package usecase

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type MediaStorage interface {
	Save(name string, content []byte) error
}

type MediaUseCase struct {
	storage MediaStorage
}

func NewMediaUseCase(storage MediaStorage) *MediaUseCase {
	return &MediaUseCase{storage: storage}
}

func (uc *MediaUseCase) UploadImage(originalName string, content []byte) (string, error) {
	extension := filepath.Ext(filepath.Base(originalName))
	if extension != "" && !validExtension(extension) {
		return "", fmt.Errorf("đuôi file không hợp lệ")
	}
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)
	if err := uc.storage.Save(name, content); err != nil {
		return "", err
	}
	return name, nil
}

func validExtension(extension string) bool {
	for _, character := range extension[1:] {
		if (character < 'a' || character > 'z') &&
			(character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') {
			return false
		}
	}
	return strings.HasPrefix(extension, ".")
}
