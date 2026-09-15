package storage

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Transfer struct {
	ID         string    `json:"id"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	CheckSum   string    `json:"check_sum"`
	Status     string    `json:"status"`
	Progress   float64   `json:"progress"`
	Mode       string    `json:"mode"`
	RemoteAddr string    `json:"remote_addr,omitempty"`
	Link       string    `json:"link"`
	CreatedAt  time.Time `json:"created_at"`
}

type FileRecord struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	CheckSum string `json:"check_sum"`
}

type Store struct {
	mu        sync.RWMutex
	transfers map[string]*Transfer
	files     []FileRecord
}

func New() *Store {
	return &Store{
		transfers: make(map[string]*Transfer),
		files:     make([]FileRecord, 0),
	}
}

func generateID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Store) CreateTransfer(fileName string, fileSize int64, mode string) *Transfer {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID()
	t := &Transfer{
		ID:        id,
		FileName:  fileName,
		FileSize:  fileSize,
		Status:    "queued",
		Mode:      mode,
		Link:      "/d/" + id,
		CreatedAt: time.Now(),
	}
	s.transfers[id] = t
	return t
}

func (s *Store) GetTransfer(id string) *Transfer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.transfers[id]
}

func (s *Store) UpdateTransfer(id string, fn func(t *Transfer)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.transfers[id]; ok {
		fn(t)
	}
}

func (s *Store) DeleteTransfer(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.transfers[id]; ok {
		delete(s.transfers, id)
		return true
	}
	return false
}

func (s *Store) ListTransfers() []*Transfer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Transfer, 0, len(s.transfers))
	for _, t := range s.transfers {
		list = append(list, t)
	}
	return list
}

func (s *Store) AddFile(name string, size int64, checkSum string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files = append(s.files, FileRecord{Name: name, Size: size, CheckSum: checkSum})
}

func (s *Store) RemoveFile(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, f := range s.files {
		if f.Name == name {
			s.files = append(s.files[:i], s.files[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) ListFiles() []FileRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]FileRecord, len(s.files))
	copy(out, s.files)
	return out
}
