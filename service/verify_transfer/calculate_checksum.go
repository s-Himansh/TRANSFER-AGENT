package service

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
)

func calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("[TRANSFER_AGENT] : Error while opening source file : %v", err)

		return "", err
	}
	defer file.Close()

	hasher := sha256.New()

	// this is great for large file, it doesn't loads the file in hasher in one go. Instead this is done in chunks
	if _, err := io.Copy(hasher, file); err != nil {
		log.Printf("[TRANSFER_AGENT] : Error while copying file data to hasher : %v", err)

		return "", err
	}

	// conversion is done to hex for easy comparison of hashes
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
