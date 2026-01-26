package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"transfer.agent/models"
	"transfer.agent/service"
	"transfer.agent/utils"
)

type Sender struct {
	receiverAddr string
}

func Init(addr string) service.Sender {
	return &Sender{receiverAddr: addr}
}

func (s *Sender) Send(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("[TRANSFER_AGENT | SENDER] : Error while opening source file : %v", err)

		return err
	}

	defer file.Close()

	fileMeta, err := file.Stat()
	if err != nil {
		log.Printf("[TRANSFER_AGENT | SENDER] : Error while retreiving source file : %v", err)

		return err
	}

	checkSum, err := utils.CalculateChecksum(path)
	if err != nil {
		log.Printf("[TRANSFER_AGENT | SENDER] : Error while retreiving checksum for the file : %v", err)

		return err
	}

	meta := &models.TransferMetaData{FileName: fileMeta.Name(), FileSize: fileMeta.Size(), CheckSum: checkSum}

	request, err := net.Dial("tcp", s.receiverAddr)
	if err != nil {
		log.Printf("[TRANSFER_AGENT | SENDER] : Error while connecting to receiver : %v", err)

		return err
	}

	metaJson, err := json.Marshal(meta)
	if err != nil {
		log.Printf("[TRANSFER_AGENT | SENDER] : Error while marshalling meta data : %v", err)

		return err
	}

	request.Write(append(metaJson, '\n'))

	// the receiver is configured to recieve the data in 32KB chunks and we'll be sending the data in same memory constraint
	buffer := make([]byte, 32*1024)

	totalBytesSent := int64(0)

	for {
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			log.Printf("[TRANSFER_AGENT | SENDER] : Error while reading data from file : %v", err)

			return err
		}

		// bytesRead is zero states the whole file is already read and possible transfered successfully
		if bytesRead == 0 {
			break
		}

		bytesSent, err := request.Write(buffer[:bytesRead])
		if err != nil {
			log.Printf("[TRANSFER_AGENT | SENDER] : Error while sending data from file : %v", err)

			return err
		}

		totalBytesSent += int64(bytesSent)

		log.Printf("[TRANSFER_AGENT | SENDER] : Progress: %.0f%% (%d/%d bytes)", float64(totalBytesSent)/float64(meta.FileSize)*100, totalBytesSent, meta.FileSize)

		if err == io.EOF {
			break
		}
	}

	respBytes := make([]byte, 1024)

	idx, err := request.Read(respBytes)
	if err != nil {
		log.Printf("[TRANSFER_AGENT | SENDER] : Error while reading response from receiver : %v", err)

		return err
	}

	resp := string(respBytes[:idx])

	log.Printf("[TRANSFER_AGENT | SENDER] : Response : %s", resp)

	if strings.Contains(resp, "successfull") {
		return nil
	} else {
		return fmt.Errorf("[TRANSFER_AGENT | SENDER] : Transfer failed : %s", resp)
	}
}
