package main

import (
	"flag"
	"log"

	receiver "transfer.agent/service/receiver"
)

func main() {
	// // intiate a test file
	// file, err := utils.CreateFile()
	// if err != nil {
	// 	return
	// }

	// defer func() {
	// 	err = file.Close()
	// }()

	// // append data into the intiated file - TEST PURPOSE: generating a file of size ~ 5 MB
	// utils.AppendData(file, 120000)

	// err = transferFiles.TransferFiles("./source_file/test_file.txt", "./generated_file/file.txt")
	// if err != nil {
	// 	return
	// }

	// _, err = verify.VerifyTransfer("./source_file/test_file.txt", "./generated_file/file.txt")
	// if err != nil {
	// 	return
	// }

	// mode := flag.String("mode", "sender", "Mode: 'sender' or 'receiver'")
	port := flag.String("port", "6789", "Port for receiver")
	// server := flag.String("server", "localhost:6780", "Server address for sender")
	// file := flag.String("file", "./source_file/test_file.txt", "File to send")

	flag.Parse()

	log.Println("[SOURCE] : Starting in RECEIVER mode")

	rvc := receiver.Init(*port, "./generated_file")

	rvc.Start()
}
