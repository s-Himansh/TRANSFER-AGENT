package main

import (
	"flag"
	"log"

	"transfer.agent/service/agent"
	receiver "transfer.agent/service/receiver"
	// sender "transfer.agent/service/sender"
)

func main() {

	// >>>> PHASE - 1

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

	// >>>> PHASE - 2

	// mode := flag.String("mode", "sender", "Mode: 'sender' or 'receiver'")
	// port := flag.String("port", "6789", "Port for receiver")
	// server := flag.String("server", "localhost:6789", "Server address for sender")
	// file := flag.String("file", "./source_file/test_file.txt", "File to send")

	// flag.Parse()

	// if *mode == "receiver" {
	// 	log.Println("[DESTINATION] : Starting in RECEIVER mode")

	// 	rvc := receiver.Init(*port, "./generated_file")

	// 	rvc.Start()
	// } else {
	// 	log.Println("[SOURCE] : Starting in SENDER mode")

	// 	sndr := sender.Init(*server)

	// 	sndr.Send(*file)
	// }

	// >>>> PHASE - 3

	mode := flag.String("mode", "sender", "Mode: 'sender' or 'receiver'")
	port := flag.String("port", "6789", "Port for receiver")
	server := flag.String("server", "localhost:6789", "Server address for sender")
	// file := flag.String("file", "./source_file/test_file.txt", "File to send")

	workers := flag.Int("workers", 3, "Number of concurrent workers")

	flag.Parse()

	if *mode == "receiver" {
		runReceiver(*port)
	} else {
		runSenderAgent(*server, *workers)
	}
}

func runReceiver(port string) {
	log.Println("[DESTINATION] : Starting in RECEIVER mode")

	rvc := receiver.Init(port, "./generated_file")

	rvc.Start()
}

func runSenderAgent(server string, workers int) {
	log.Println("[SOURCE] : Starting in SENDER mode")

	// this will initialise an agent assigned to work out the transfer
	agent := agent.NewAgent("sneilh", server, workers)

	// this will fire workers in memory and workers will continously look for sent files
	agent.Start()

	// once all the jobs are free, shutdown the agent
	defer agent.Stop()

	// CONTINUATION : CLI to view status of transfers
}
