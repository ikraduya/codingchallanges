package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mattn/go-sqlite3"
	"github.com/xyproto/randomstring"
	"ikraduya.dev/urlshortener/backend/model"
	"ikraduya.dev/urlshortener/backend/service"
)

type URLHandler struct {
	Repository model.Repository
}

// Post handles HTTP Post - /
func (h *URLHandler) PostURL(c *gin.Context) {
	var newURL model.PostURLData

	if err := c.ShouldBindJSON(&newURL); err != nil {
		errorStrings := make([]string, 0)
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				if fieldError.Tag() == "required" {
					errorStrings = append(errorStrings, fmt.Sprintf("Missing field: %s", strings.ToLower(fieldError.Field())))
				}
			}
		}

		c.String(http.StatusBadRequest, "%s\n", strings.Join(errorStrings, "\n"))
		return
	}

	// service logics
	checkIsLongURLAlreadyExist := func(longURL string) (bool, error) {
		isLongURLExist, key, err := h.Repository.IsLongURLExists(newURL.URL)
		if err != nil {
			return false, err
		}
		if !isLongURLExist {
			return isLongURLExist, nil
		}

		protocol := "http"
		if c.Request.TLS != nil {
			protocol = "https"
		}
		shortURL := fmt.Sprintf("%s://%s/%s", protocol, c.Request.Host, key)
		URLResponse := model.PostURLResponse{Key: key, LongURL: longURL, ShortURL: shortURL}

		c.IndentedJSON(http.StatusCreated, URLResponse)
		return true, nil
	}
	isAlreadyExist, err := checkIsLongURLAlreadyExist(newURL.URL)
	if err != nil {
		log.Printf("DB repository error: %v\n", err)
		c.String(http.StatusInternalServerError, "An internal server error occurred\n")
		return
	}
	if isAlreadyExist {
		return
	}

	hashLongURL := func(longURL string) (hashOutput string, err error) {
		hashOutput, err = service.GetHash(longURL, 6)
		if err != nil {
			return "", err
		}

		return hashOutput, nil
	}
	key, err := hashLongURL(newURL.URL)
	if err != nil {
		log.Printf("Service logic error: %v\n", err)
		c.String(http.StatusInternalServerError, "An internal server error occurred\n")
		return
	}
	key = "cdb4d8"
	url := model.URL{LongURL: newURL.URL, Key: key}
	log.Println("url", url)

	// if already exists, return already exists response code
	err = h.Repository.Create(url)
	if err != nil {
		// if err == ErrConstraintUnique, rehash and try the logic once more
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				errInStr := fmt.Sprint(sqliteErr)
				if strings.Contains(errInStr, ".key") {
					extraSuffix := randomstring.String(5)
					key, err = hashLongURL(newURL.URL + extraSuffix)
					if err != nil {
						log.Printf("Service logic error: %v\n", err)
						c.String(http.StatusInternalServerError, "An internal server error occurred\n")
						return
					}

					url.Key = key
					err = h.Repository.Create(url)
					if err != nil {
						log.Printf("DB repository error: %v\n", err)
						c.String(http.StatusInternalServerError, "An internal server error occurred\n")
						return
					} // else, continue
				}
			}
		} else {
			log.Printf("DB repository error: %v\n", err)
			c.String(http.StatusInternalServerError, "An internal server error occurred\n")
			return
		}
	}

	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	shortURL := fmt.Sprintf("%s://%s/%s", protocol, c.Request.Host, url.Key)
	URLResponse := model.PostURLResponse{Key: url.Key, LongURL: url.LongURL, ShortURL: shortURL}

	c.IndentedJSON(http.StatusCreated, URLResponse)
}

// Get handles URL redirection
func (h *URLHandler) GetURL(c *gin.Context) {
	key := c.Param("key")
	longURL, err := h.Repository.Retrieve(key)
	if err != nil {
		if err == sql.ErrNoRows {
			c.String(http.StatusNotFound, "URL not found\n")
		} else {
			log.Printf("DB repository error: %v\n", err)
			c.String(http.StatusInternalServerError, "An internal server error occurred\n")
		}
		return
	}

	c.Redirect(http.StatusFound, longURL)
}

func (h *URLHandler) DeleteURL(c *gin.Context) {
	key := c.Param("key")
	success, err := h.Repository.Delete(key)
	if err != nil {
		log.Printf("DB repository error: %v\n", err)
		c.String(http.StatusInternalServerError, "An internal server error occurred\n")
	}

	if !success {
		c.String(http.StatusNotFound, "URL not found\n")
	} else {
		c.Status(http.StatusAccepted)
	}
}
