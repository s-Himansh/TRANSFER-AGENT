package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
)

// CreateFile creates a test_file.txt file
func CreateFile() (*os.File, error) {
	file, err := os.Create("test_file.txt")
	if err != nil {
		log.Printf("Error while open a dummy file : %v", err)

		return nil, err
	}

	return file, nil
}

// AppendData inputs the file, number of lines that needs to be generated and appends those many lines to the file
func AppendData(file *os.File, numOfLines int) {
	if file == nil {
		return
	}

	for i := 1; i <= numOfLines; i++ {
		str := fmt.Sprintf("Generated line -> %d\n", i)

		_, err := file.WriteString(str)
		if err != nil {
			log.Printf("Error while writing data to file : %v", err)
		}
	}

	logFileSize(file)
}

// logFileSize logs the file size
func logFileSize(file *os.File) {
	if file == nil {
		return
	}

	info, err := os.Stat(file.Name())
	if err != nil {
		log.Printf("Error while retrieving info for a file : %v", err)

		return
	}

	log.Printf("[TRANSFER_AGENT] : File size: %d bytes", info.Size())
}

func CalculateChecksum(path string) (string, error) {
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
