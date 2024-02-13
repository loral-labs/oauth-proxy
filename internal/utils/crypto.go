package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateNonce(byteLength int) (string, error) {
	nonce := make([]byte, byteLength)
	if _, err := rand.Read(nonce); err != nil {
		return "", err // return the error if random read fails
	}
	return base64.URLEncoding.EncodeToString(nonce), nil // encode to base64 for easy handling and URL safety
}
