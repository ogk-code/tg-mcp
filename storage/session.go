package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/gotd/td/session"
)

type FileStorage struct {
	path string
	mu   sync.Mutex
}

func NewFileStorage(path string) *FileStorage {
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, ".tg-mcp-session.json")
	}
	return &FileStorage{path: path}
}

func (s *FileStorage) LoadSession(_ context.Context) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil, session.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var stored storedSession
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}

	return stored.Data, nil
}

func (s *FileStorage) StoreSession(_ context.Context, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored := storedSession{Data: data}
	jsonData, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(s.path, jsonData, 0600)
}

func (s *FileStorage) Path() string {
	return s.path
}

func (s *FileStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.Remove(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

type storedSession struct {
	Data []byte `json:"data"`
}
