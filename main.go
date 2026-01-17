package main

import (
	service "transfer.agent/service/transfer_files"
	"transfer.agent/utils"
)

func main() {
	// intiate a test file
	file, err := utils.CreateFile()
	if err != nil {
		return
	}

	defer file.Close()

	// append data into the intiated file
	utils.AppendData(file, 120000)

	err = service.TransferFiles("./test_file.txt", "./generated_file/file.txt")
	if err != nil {
		return
	}
}
