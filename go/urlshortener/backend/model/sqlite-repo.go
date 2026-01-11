package model

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepo struct {
	DB *sql.DB
}

func OpenSQLiteRepo(sourceName string) (*SQLiteRepo, error) {
	db, err := sql.Open("sqlite3", sourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS url (
		"id" INTEGER PRIMARY KEY AUTOINCREMENT,
		"longurl" TEXT NOT NULL,
		"key" TEXT NOT NULL UNIQUE
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	return &SQLiteRepo{db}, nil
}

func (r *SQLiteRepo) Close() error {
	err := r.DB.Close()
	return err
}

func (r *SQLiteRepo) Create(url URL) error {
	query := "INSERT INTO url(longurl, key) VALUES(?, ?);"
	_, err := r.DB.Exec(query, url.LongURL, url.Key)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteRepo) Retrieve(key string) (longURL string, err error) {
	query := "SELECT longurl FROM url WHERE key = ? LIMIT 1;"
	if err := r.DB.QueryRow(query, key).Scan(&longURL); err != nil {
		return "", err
	}

	return longURL, nil
}

func (r *SQLiteRepo) Delete(key string) (success bool, err error) {
	query := "DELETE FROM url WHERE key = ?"
	result, err := r.DB.Exec(query, key)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if rowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

func (r *SQLiteRepo) IsLongURLExists(longURL string) (isExist bool, key string, err error) {
	query := "SELECT key FROM url WHERE longurl = ? LIMIT 1;"
	if err = r.DB.QueryRow(query, longURL).Scan(&key); err != nil {
		fmt.Println("err islongurlexists", err)
		if err == sql.ErrNoRows {
			return false, "", nil
		}
		return false, "", err
	}

	return true, key, nil
}
