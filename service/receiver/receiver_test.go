package receiver_test

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
	"transfer.agent/models"
	"transfer.agent/service/receiver"
	"transfer.agent/service/sender"
)

func TestReceiverSuccessfulTransfer(t *testing.T) {
	dir := t.TempDir()
	r := receiver.Init("19789", dir)
	go r.Start()
	time.Sleep(100 * time.Millisecond)
	t.Cleanup(func() { r.Shutdown() })

	sndr := sender.Init("localhost:19789")

	srcPath := filepath.Join(t.TempDir(), "send.txt")
	content := []byte("hello transfer agent")
	os.WriteFile(srcPath, content, 0644)

	err := sndr.Send(srcPath)
	if err != nil {
		t.Fatalf("transfer failed: %v", err)
	}

	destPath := filepath.Join(dir, "send.txt")
	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("file not saved: %v", err)
	}

	if string(got) != string(content) {
		t.Fatalf("content mismatch: got %q, want %q", got, content)
	}
}

func TestReceiverChecksumMismatch(t *testing.T) {
	dir := t.TempDir()
	r := receiver.Init("19790", dir)
	go r.Start()
	time.Sleep(100 * time.Millisecond)
	t.Cleanup(func() { r.Shutdown() })

	conn, err := net.DialTimeout("tcp", "localhost:19790", time.Second)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	meta := models.TransferMetaData{
		FileName: "bad.txt",
		FileSize: 5,
		CheckSum: "0000000000000000000000000000000000000000000000000000000000000000",
	}
	metaJSON, _ := json.Marshal(meta)
	conn.Write(append(metaJSON, '\n'))
	conn.Write([]byte("hello"))

	respBytes := make([]byte, 4096)
	n, _ := conn.Read(respBytes)
	resp := string(respBytes[:n])

	if resp == "" {
		t.Fatal("expected response")
	}

	savedPath := filepath.Join(dir, "bad.txt")
	if _, err := os.Stat(savedPath); !os.IsNotExist(err) {
		t.Fatal("corrupted file should be deleted")
	}
}
