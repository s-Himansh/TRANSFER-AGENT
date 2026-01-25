package service

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"

	"transfer.agent/models"
	"transfer.agent/utils"
)

type Receiver struct {
	port          string
	saveDirectory string
}

func InitReceiever(port, directory string) *Receiver {
	return &Receiver{port: port, saveDirectory: directory}
}

func (r *Receiver) Start() error {
	// intial directory creation should be dynamic
	err := os.MkdirAll(r.saveDirectory, 0755)
	if err != nil {
		log.Printf("[TRANSFER_AGENT | RECEIVER] : Error while creating target directory for storing file %v", err)

		return err
	}

	listener, err := net.Listen("tcp", ":"+r.port)
	if err != nil {
		log.Printf("[TRANSFER_AGENT | RECEIVER] : Error while intialising listener to the client %v", err)

		return err
	}

	defer listener.Close()

	for {
		request, err := listener.Accept()
		if err != nil {
			log.Printf("[TRANSFER_AGENT | RECEIVER] : Error while accepting connections from client %v", err)

			continue
		}

		log.Printf("[TRANSFER_AGENT | RECEIVER] : New connection from: %s", request.RemoteAddr())

		go r.handleIncomingRequests(request)
	}
}

func (r *Receiver) handleIncomingRequests(request net.Conn) {
	defer request.Close() // closing requests is imp. to not end up loosing unecassary memory and crash the server

	reader := bufio.NewReader(request)

	metaDataStr, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("[TRANSFER_AGENT | RECEIVER] : Error while reading meta data for the request came from client %v", err)

		request.Write([]byte(`[RECEIVER] : Failed to read meta data`))

		return
	}

	parsedMeta := &models.TransferMetaData{}

	err = json.Unmarshal([]byte(metaDataStr), &parsedMeta)
	if err != nil {
		log.Printf("[TRANSFER_AGENT | RECEIVER] : Error while parsing meta data for the request came from client %v", err)

		request.Write([]byte(`[RECEIVER] : Failed to read meta data`))

		return
	}

	log.Printf("[TRANSFER_AGENT | RECEIVER] : Receiving: %s {%.2f MB}", parsedMeta.FileName, float64(parsedMeta.FileSize)/(1024*1024))

	file, err := os.Create(r.saveDirectory + "/" + parsedMeta.FileName)
	if err != nil {
		log.Printf("[TRANSFER_AGENT | RECEIVER] : Error while creating file in target path for the request came from client %v", err)

		request.Write([]byte(`[RECEIVER] : Failed to create target file`))

		return
	}

	defer file.Close()

	// we need to consume the data in same chunks as it is being sent with same memory buffer

	bytesRecieved := int64(0)

	buffer := make([]byte, 32*1024)

	// bytesRecieved keeps track of how many we bytes we need to read and break the flow acc.
	for bytesRecieved < parsedMeta.FileSize {
		bytesToRead := int64(len(buffer))
		remainingBytes := parsedMeta.FileSize - bytesRecieved

		// in case the chunk of data is less than the defined buffer size, update bytesToRead acc.
		if remainingBytes < bytesToRead {
			bytesToRead = remainingBytes
		}

		// question may arise, buffer can only consume within it's capacity so why slicing ?
		// answer to this is if buffer is having less bytes of data possibly when only single chunk is left to consume, the parsing will make sure to only read the left chunk
		// buffer might block read as it'll not be able to consume data to it's full allocated space
		bytesRead, err := reader.Read(buffer[:bytesToRead])
		if err != nil && err != io.EOF {
			log.Printf("[TRANSFER_AGENT | RECEIVER] : Error while reading source file for the request came from client %v", err)

			request.Write([]byte(`[RECEIVER] : Failed to read sent file`))

			return
		}

		if bytesRead > 0 {
			_, err = file.Write(buffer[:bytesRead])
			if err != nil {
				log.Printf("[TRANSFER_AGENT | RECEIVER] : Error while writing source chunk to target for the request came from client %v", err)

				request.Write([]byte(`[RECEIVER] : Failed to write sent file into target`))

				return
			}

			bytesRecieved += int64(bytesRead)

			log.Printf("[TRANSFER_AGENT | RECEIVER] : Progress: %.0f%% (%d/%d bytes)", float64(bytesRecieved)/float64(parsedMeta.FileSize)*100, bytesRecieved, parsedMeta.FileSize)
		}

		if err == io.EOF {
			break
		}
	}

	checkSum, err := utils.CalculateChecksum(r.saveDirectory + "/" + parsedMeta.FileName)
	if err != nil {
		log.Printf("[TRANSFER_AGENT | RECEIVER] : Error while calculating checksum for recieved file %v", err)

		request.Write([]byte(`[RECEIVER] : Failed to validate sent file into target`))

		return
	}

	if parsedMeta.CheckSum != checkSum {
		log.Printf("[TRANSFER_AGENT | RECEIVER] : Checksum validation failed due to misatch")

		request.Write([]byte(`[RECEIVER] : Checksum validation failed due to misatch`))

		return
	}

	log.Printf("[TRANSFER_AGENT | RECEIVER] : File transfer and validation successfull")

	request.Write([]byte(`[RECEIVER] : File transfer and validation successfull`))
}
