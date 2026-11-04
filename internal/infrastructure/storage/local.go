package storage

import (
	"api/internal/domain"
	"errors"
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

func (s *LocalFileStorage) List() ([]domain.MediaFile, error) {
	entries, err := os.ReadDir(s.directory)
	if err != nil {
		return nil, fmt.Errorf("không thể đọc thư mục upload: %w", err)
	}
	files := make([]domain.MediaFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, domain.MediaFile{Name: entry.Name(), Size: info.Size(), ModifiedAt: info.ModTime()})
	}
	return files, nil
}

func (s *LocalFileStorage) Delete(name string) error {
	err := os.Remove(filepath.Join(s.directory, filepath.Base(name)))
	if errors.Is(err, os.ErrNotExist) {
		return domain.ErrMediaNotFound
	}
	if err != nil {
		return fmt.Errorf("không thể xóa file upload: %w", err)
	}
	return nil
}
