package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateChecksum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	hash1, err := CalculateChecksum(path)
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := CalculateChecksum(path)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 != hash2 {
		t.Fatalf("checksums differ: %s != %s", hash1, hash2)
	}

	if hash1 == "" {
		t.Fatal("checksum is empty")
	}
}

func TestCalculateChecksum_DifferentFiles(t *testing.T) {
	dir := t.TempDir()

	path1 := filepath.Join(dir, "a.txt")
	path2 := filepath.Join(dir, "b.txt")
	os.WriteFile(path1, []byte("content A"), 0644)
	os.WriteFile(path2, []byte("content B"), 0644)

	h1, _ := CalculateChecksum(path1)
	h2, _ := CalculateChecksum(path2)

	if h1 == h2 {
		t.Fatal("different files should have different checksums")
	}
}

func TestCalculateChecksum_MissingFile(t *testing.T) {
	_, err := CalculateChecksum("/nonexistent/file.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
