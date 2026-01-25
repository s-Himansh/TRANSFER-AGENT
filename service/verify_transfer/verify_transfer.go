package service

import (
	"log"

	"transfer.agent/utils"
)

func VerifyTransfer(sourcePath, destinationPath string) (transferStatus bool, err error) {
	log.Println("[TRANSFER_AGENT] : Verifying transfer...")

	source, err := utils.CalculateChecksum(sourcePath)
	if err != nil {
		log.Printf("[TRANSFER_AGENT] : Error while generating checksum for source file : %v", err)

		return false, err
	}

	destination, err := utils.CalculateChecksum(destinationPath)
	if err != nil {
		log.Printf("[TRANSFER_AGENT] : Error while generating checksum for destination file : %v", err)

		return false, err
	}

	if source != destination {
		return false, nil
	}

	transferStatus = true

	log.Printf("[TRANSFER_AGENT] : Transfer verified : Validation -> %v", transferStatus)

	return
}
