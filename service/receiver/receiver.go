package receiver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"transfer.agent/models"
	"transfer.agent/utils"
)

type Receiver struct {
	port          string
	saveDirectory string
	listener      net.Listener
	wg            sync.WaitGroup
}

func Init(port, directory string) *Receiver {
	return &Receiver{port: port, saveDirectory: directory}
}

func (r *Receiver) Start() error {
	if err := os.MkdirAll(r.saveDirectory, 0755); err != nil {
		return fmt.Errorf("create save directory: %w", err)
	}

	var err error
	r.listener, err = net.Listen("tcp", ":"+r.port)
	if err != nil {
		return fmt.Errorf("listen on port %s: %w", r.port, err)
	}

	log.Printf("[RECEIVER] Listening on port %s", r.port)

	for {
		conn, err := r.listener.Accept()
		if err != nil {
			select {
			case <-context.Background().Done():
				return nil
			default:
				log.Printf("[RECEIVER] Accept error: %v", err)
				continue
			}
		}
		log.Printf("[RECEIVER] New connection from: %s", conn.RemoteAddr())
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			r.handleConnection(conn)
		}()
	}
}

func (r *Receiver) Shutdown() {
	if r.listener != nil {
		r.listener.Close()
	}
	r.wg.Wait()
}

func (r *Receiver) handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(5 * time.Minute))

	reader := bufio.NewReader(conn)

	metaDataStr, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("[RECEIVER] Error reading metadata: %v", err)
		conn.Write([]byte(`{"status":"error","message":"failed to read metadata"}`))
		return
	}

	var parsedMeta models.TransferMetaData
	if err := json.Unmarshal([]byte(metaDataStr), &parsedMeta); err != nil {
		log.Printf("[RECEIVER] Error parsing metadata: %v", err)
		conn.Write([]byte(`{"status":"error","message":"invalid metadata"}`))
		return
	}

	// Sanitize filename to prevent path traversal
	cleanName := filepath.Base(parsedMeta.FileName)
	if cleanName == "." || cleanName == "/" {
		conn.Write([]byte(`{"status":"error","message":"invalid filename"}`))
		return
	}

	log.Printf("[RECEIVER] Receiving: %s (%.2f MB)", cleanName, float64(parsedMeta.FileSize)/(1024*1024))

	filePath := filepath.Join(r.saveDirectory, cleanName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("[RECEIVER] Error creating file: %v", err)
		conn.Write([]byte(`{"status":"error","message":"failed to create file"}`))
		return
	}
	defer file.Close()

	bytesReceived := int64(0)
	buffer := make([]byte, 32*1024)

	for bytesReceived < parsedMeta.FileSize {
		remaining := parsedMeta.FileSize - bytesReceived
		toRead := int64(len(buffer))
		if remaining < toRead {
			toRead = remaining
		}

		bytesRead, err := reader.Read(buffer[:toRead])
		if bytesRead > 0 {
			if _, writeErr := file.Write(buffer[:bytesRead]); writeErr != nil {
				log.Printf("[RECEIVER] Error writing data: %v", writeErr)
				conn.Write([]byte(`{"status":"error","message":"write failed"}`))
				return
			}
			bytesReceived += int64(bytesRead)
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("[RECEIVER] Error reading data: %v", err)
			conn.Write([]byte(`{"status":"error","message":"read failed"}`))
			return
		}
	}

	file.Close()

	checkSum, err := utils.CalculateChecksum(filePath)
	if err != nil {
		log.Printf("[RECEIVER] Error calculating checksum: %v", err)
		conn.Write([]byte(`{"status":"error","message":"checksum calculation failed"}`))
		return
	}

	if parsedMeta.CheckSum != checkSum {
		log.Printf("[RECEIVER] Checksum mismatch: expected %s, got %s", parsedMeta.CheckSum, checkSum)
		os.Remove(filePath)
		conn.Write([]byte(`{"status":"error","message":"checksum mismatch"}`))
		return
	}

	log.Printf("[RECEIVER] Transfer successful: %s", cleanName)
	fmt.Fprintf(conn, `{"status":"success","message":"file %s transferred successfully"}`, cleanName)
}

