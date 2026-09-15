package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	receiver "transfer.agent/service/receiver"
	sender "transfer.agent/service/sender"
)

func main() {
	mode := flag.String("mode", "sender", "Mode: 'sender' or 'receiver'")
	port := flag.String("port", "6789", "Port for receiver")
	server := flag.String("server", "localhost:6789", "Server address for sender")
	file := flag.String("file", "./source_file/test_file.txt", "File to send")
	saveDir := flag.String("savedir", "./generated_file", "Directory to save received files")

	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if *mode == "receiver" {
		log.Println("[RECEIVER] Starting in receiver mode")

		rvc := receiver.Init(*port, *saveDir)
		go func() {
			<-ctx.Done()
			rvc.Shutdown()
		}()

		if err := rvc.Start(); err != nil {
			log.Fatalf("[RECEIVER] Failed to start: %v", err)
		}
	} else {
		log.Println("[SENDER] Starting in sender mode")

		sndr := sender.Init(*server)
		if err := sndr.Send(*file); err != nil {
			log.Fatalf("[SENDER] Transfer failed: %v", err)
		}
	}
}
