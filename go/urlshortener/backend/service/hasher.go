package service

import (
	"crypto/sha256"
	"encoding/hex"
)

func GetHash(data string, length int) (string, error) {
	hasher := sha256.New()

	_, err := hasher.Write([]byte(data))
	if err != nil {
		return "", err
	}

	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	if len(hashString) > length {
		return hashString[:length], nil
	}
	return hashString, nil
}
