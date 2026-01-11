package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/xyproto/randomstring"

	"ikraduya.dev/urlshortener/backend/handlers"
	"ikraduya.dev/urlshortener/backend/model"
)

func main() {
	randomstring.Seed()

	repo, err := model.OpenSQLiteRepo("url.db")
	if err != nil {
		log.Fatal("Can't open SQLite DB")
	}
	defer repo.Close()
	handlers := &handlers.URLHandler{Repository: repo}

	router := gin.Default()
	router.Use(cors.Default())
	router.POST("/", handlers.PostURL)
	router.GET("/:key", handlers.GetURL)
	router.DELETE("/:key", handlers.DeleteURL)

	router.Run(":8080")
}
