package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/VTV06012004/ai_cms/internal/domain"
	"os"
	"path/filepath"
)

type LocalFileStorage struct {
	directory string
}

func NewLocalFileStorage(directory string) *LocalFileStorage {
	return &LocalFileStorage{directory: directory}
}

func (s *LocalFileStorage) Save(ctx context.Context, name string, content []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
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

func (s *LocalFileStorage) List(ctx context.Context) ([]domain.MediaFile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.directory)
	if err != nil {
		return nil, fmt.Errorf("không thể đọc thư mục upload: %w", err)
	}
	files := make([]domain.MediaFile, 0, len(entries))
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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

func (s *LocalFileStorage) Delete(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	err := os.Remove(filepath.Join(s.directory, filepath.Base(name)))
	if errors.Is(err, os.ErrNotExist) {
		return domain.ErrMediaNotFound
	}
	if err != nil {
		return fmt.Errorf("không thể xóa file upload: %w", err)
	}
	return nil
}
