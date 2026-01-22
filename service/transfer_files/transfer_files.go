package service

import (
	"io"
	"log"
	"os"
)

func TransferFiles(sourcePath, destinationPath string) error {
	srcFile, err := os.Open(sourcePath)
	if err != nil {
		log.Printf("[TRANSFER_AGENT] : Error while opening source file : %v", err)

		return err
	}

	defer srcFile.Close()

	destFile, err := os.Create(destinationPath)
	if err != nil {
		log.Printf("[TRANSFER_AGENT] : Error while creating destination file : %v", err)

		return err
	}

	defer destFile.Close()

	// create a buffer of 4 KB to transfer the file with it's data in chunks to faciltate transfer of huge size files
	buffer := make([]byte, 4096)

	for {
		bytesRead, err := srcFile.Read(buffer)
		if err != nil && err != io.EOF {
			log.Printf("[TRANSFER_AGENT] : Error while reading bytes from source file : %v", err)

			return err
		}

		// bytesRead is zero states the whole file is already read and possible transfered successfully
		if bytesRead == 0 {
			break
		}

		_, err = destFile.Write(buffer[:bytesRead])
		if err != nil {
			log.Printf("[TRANSFER_AGENT] : Error while writing bytes to destination file : %v", err)

			return err
		}
	}

	return nil
}
