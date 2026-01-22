package service

import (
	"bytes"
	"io"
	"log"
	"os"

	"transfer.agent/models"
	"transfer.agent/utils"
)

func generateChecksum(content []byte) *models.Checksum {
	checkSumHash := utils.CreateSHA256Hash(content)

	return &models.Checksum{Hash: checkSumHash}
}

func VerifyTransfer(sourcePath, destinationPath string) (string, error) {
	srcFile, err := os.Open(sourcePath)
	if err != nil {
		log.Printf("[TRANSFER_AGENT] : Error while opening source file : %v", err)

		return "FAILED", err
	}

	defer srcFile.Close()

	destFile, err := os.Open(destinationPath)
	if err != nil {
		log.Printf("[TRANSFER_AGENT] : Error while opening destination file : %v", err)

		return "FAILED", err
	}

	defer destFile.Close()

	sourceFileContent, err := io.ReadAll(srcFile)
	if err != nil {
		log.Printf("[TRANSFER_AGENT] : Error while reading source file : %v", err)

		return "FAILED", err
	}

	destinationFileContent, err := io.ReadAll(destFile)
	if err != nil {
		log.Printf("[TRANSFER_AGENT] : Error while reading destination file : %v", err)

		return "FAILED", err
	}

	sourceChecksum := generateChecksum(sourceFileContent)
	destinationChecksum := generateChecksum(destinationFileContent)

	isSuccess := bytes.Equal(sourceChecksum.Hash, destinationChecksum.Hash)

	if !isSuccess {
		return "FAILED", nil
	}

	return "SUCCESS", nil
}
