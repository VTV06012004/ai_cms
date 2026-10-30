package usecase

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"strings"
	"time"

	_ "golang.org/x/image/webp"
)

const MaxImageSize = 5 << 20

var (
	ErrInvalidImage  = errors.New("chỉ chấp nhận ảnh JPG, PNG, WebP hoặc GIF hợp lệ")
	ErrImageTooLarge = errors.New("kích thước ảnh tối đa là 5 MB")
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
	if len(content) > MaxImageSize {
		return "", ErrImageTooLarge
	}
	extension := strings.ToLower(filepath.Ext(filepath.Base(originalName)))
	expectedFormat, validExtension := supportedImageExtensions[extension]
	if !validExtension {
		return "", ErrInvalidImage
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil || format != expectedFormat {
		return "", ErrInvalidImage
	}
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)
	if err := uc.storage.Save(name, content); err != nil {
		return "", err
	}
	return name, nil
}

var supportedImageExtensions = map[string]string{
	".jpg":  "jpeg",
	".jpeg": "jpeg",
	".png":  "png",
	".webp": "webp",
	".gif":  "gif",
}
