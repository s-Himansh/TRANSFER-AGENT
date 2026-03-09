package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"transfer.agent/models"
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

	commandLineInterface(agent)
}

func commandLineInterface(agent *agent.Agent) {
	fmt.Println("\n=== File Transfer Agent CLI ===")
	fmt.Println("Commands:")
	fmt.Println("  send <filepath>  - Submit a file for transfer")
	fmt.Println("  status <id>      - Check transfer status")
	fmt.Println("  list             - List all transfers")
	fmt.Println("  quit             - Exit")
	fmt.Println("===============================\n")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		command := parts[0]

		switch command {
		case "send":
			if len(parts) < 2 {
				fmt.Println("Usage: send <filepath>")
				continue
			}

			filePath := parts[1]

			_, err := os.Stat(filePath)
			if os.IsNotExist(err) {
				fmt.Printf("File is not present at %s", filePath)

				continue
			}

			jobID, err := agent.SubmitTransfer(filePath)
			if err != nil {
				fmt.Printf("ERROR : %v/n", err)

				continue
			}

			fmt.Printf("Transfer submitted. The assigned JOB-ID is %s\n", jobID)

		case "status":
			if len(parts) < 2 {
				fmt.Println("Usage: status <job-id>")
				continue
			}

			jobID := parts[1]

			job, err := agent.RetrieveJob(jobID)
			if err != nil {
				fmt.Printf("ERROR : %v/n", err)

				continue
			}

			printJobStatus(job)

		case "list":
			jobs := agent.ListJobs()

			if len(jobs) == 0 {
				fmt.Println("No transfer history found")

				continue
			}

			// fmt.Printf("\n%-20s %-15s %-30s %-10s\n", "ID", "Status", "File", "Duration")

			fmt.Println(strings.Repeat("-", 80))

			for _, job := range jobs {
				duration := "-"

				if job.CompletedAt != nil {
					duration = fmt.Sprintf("%.1fs", job.CompletedAt.Sub(*job.SubmittedAt).Seconds())
				}

				fmt.Printf("%-20s %-15s %-30s %-10s\n", job.ID, job.Status, job.SourcePath, duration)
			}

			fmt.Println()

		case "quit", "exit":
			fmt.Println("Shutting down...")

			return

		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}
}

func printJobStatus(job *models.TransferJob) {
	fmt.Println("\n=== Transfer Status ===")
	fmt.Printf("ID:         %s\n", job.ID)
	fmt.Printf("Status:     %s\n", job.Status)
	fmt.Printf("File:       %s\n", job.SourcePath)
	fmt.Printf("Submitted:  %s\n", job.SubmittedAt.Format(time.RFC3339))

	if job.StartedAt != nil {
		fmt.Printf("Started:    %s\n", job.StartedAt.Format(time.RFC3339))
	}

	if job.CompletedAt != nil {
		fmt.Printf("Completed:  %s\n", job.CompletedAt.Format(time.RFC3339))
		duration := job.CompletedAt.Sub(*job.SubmittedAt)
		fmt.Printf("Duration:   %.2fs\n", duration.Seconds())
	}

	if job.Error != "" {
		fmt.Printf("Error:      %s\n", job.Error)
	}

	fmt.Println("=======================\n")
}
