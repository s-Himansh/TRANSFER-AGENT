package sender

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"transfer.agent/models"
	"transfer.agent/utils"
)

type Sender struct {
	receiverAddr string
}

func Init(addr string) *Sender {
	return &Sender{receiverAddr: addr}
}

type transferResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (s *Sender) Send(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	checkSum, err := utils.CalculateChecksum(path)
	if err != nil {
		return fmt.Errorf("calculate checksum: %w", err)
	}

	meta := &models.TransferMetaData{
		FileName: fileInfo.Name(),
		FileSize: fileInfo.Size(),
		CheckSum: checkSum,
	}

	conn, err := net.DialTimeout("tcp", s.receiverAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect to receiver: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Minute))

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	if _, err := conn.Write(append(metaJSON, '\n')); err != nil {
		return fmt.Errorf("send metadata: %w", err)
	}

	buffer := make([]byte, 32*1024)
	totalBytesSent := int64(0)

	for {
		bytesRead, readErr := file.Read(buffer)
		if bytesRead > 0 {
			bytesSent, writeErr := conn.Write(buffer[:bytesRead])
			if writeErr != nil {
				return fmt.Errorf("send data: %w", writeErr)
			}
			totalBytesSent += int64(bytesSent)
			pct := float64(totalBytesSent) / float64(meta.FileSize) * 100
			log.Printf("[SENDER] Progress: %.0f%% (%d/%d bytes)", pct, totalBytesSent, meta.FileSize)
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("read file: %w", readErr)
		}
	}

	respBytes := make([]byte, 4096)
	n, err := conn.Read(respBytes)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var resp transferResponse
	if err := json.Unmarshal(respBytes[:n], &resp); err != nil {
		raw := string(respBytes[:n])
		if raw != "" {
			log.Printf("[SENDER] Response: %s", raw)
		}
		return fmt.Errorf("parse response: %w", err)
	}

	log.Printf("[SENDER] Response: %s", resp.Message)

	if resp.Status != "success" {
		return fmt.Errorf("transfer failed: %s", resp.Message)
	}

	return nil
}
