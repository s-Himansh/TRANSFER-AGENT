package utils

import (
	"crypto/sha256"
)

func CreateSHA256Hash(data []byte) []byte {
	hasher := sha256.New()

	hasher.Write(data)

	return hasher.Sum(nil)
}
