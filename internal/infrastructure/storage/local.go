package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

type LocalFileStorage struct {
	directory string
}

func NewLocalFileStorage(directory string) *LocalFileStorage {
	return &LocalFileStorage{directory: directory}
}

func (s *LocalFileStorage) Save(name string, content []byte) error {
	path := filepath.Join(s.directory, filepath.Base(name))
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("không thể tạo file upload: %w", err)
	}
	if _, err := file.Write(content); err != nil {
		_ = file.Close()
		return fmt.Errorf("không thể ghi file upload: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("không thể đóng file upload: %w", err)
	}
	return nil
}
