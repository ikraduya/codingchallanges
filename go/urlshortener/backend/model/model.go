package model

type PostURLData struct {
	URL string `json:"url" binding:"required"`
}

type PostURLResponse struct {
	Key      string `json:"key"`
	LongURL  string `json:"long_url"`
	ShortURL string `json:"short_url"`
}

type URL struct {
	LongURL string
	Key     string
}

type Repository interface {
	Create(URL) error
	Retrieve(key string) (longURL string, err error)
	Delete(key string) (success bool, err error)
	IsLongURLExists(longURL string) (isExist bool, key string, err error)
}
