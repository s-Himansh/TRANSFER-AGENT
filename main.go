package main

import (
	"log"

	transferFiles "transfer.agent/service/transfer_files"
	verify "transfer.agent/service/verify_transfer"
	"transfer.agent/utils"
)

func main() {
	// intiate a test file
	file, err := utils.CreateFile()
	if err != nil {
		return
	}

	defer file.Close()

	// append data into the intiated file - TEST PURPOSE: generating a file of size ~ 5 MB
	utils.AppendData(file, 120000)

	err = transferFiles.TransferFiles("./source_file/test_file.txt", "./generated_file/file.txt")
	if err != nil {
		return
	}

	transferStatus, err := verify.VerifyTransfer("./source_file/test_file.txt", "./generated_file/file.txt")
	if err != nil {
		return
	}

	log.Printf("[TRANSFER_AGENT] : Transfered verified : Status -> %v", transferStatus)
}
