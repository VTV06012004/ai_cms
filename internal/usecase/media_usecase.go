package usecase

import (
	"api/internal/domain"
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"sort"
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
	Save(ctx context.Context, name string, content []byte) error
	List(ctx context.Context) ([]domain.MediaFile, error)
	Delete(ctx context.Context, name string) error
}

type MediaUseCase struct {
	storage MediaStorage
}

func NewMediaUseCase(storage MediaStorage) *MediaUseCase {
	return &MediaUseCase{storage: storage}
}

func (uc *MediaUseCase) UploadImage(ctx context.Context, originalName string, content []byte) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
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
	if err := uc.storage.Save(ctx, name, content); err != nil {
		return "", err
	}
	return name, nil
}

func (uc *MediaUseCase) ListImages(ctx context.Context, page domain.PageRequest) ([]domain.MediaFile, domain.PageInfo, error) {
	all, err := uc.storage.List(ctx)
	if err != nil {
		return nil, domain.PageInfo{}, err
	}
	images := make([]domain.MediaFile, 0, len(all))
	for _, file := range all {
		if isImageName(file.Name) {
			images = append(images, file)
		}
	}
	sort.Slice(images, func(i, j int) bool {
		if !images[i].ModifiedAt.Equal(images[j].ModifiedAt) {
			return images[i].ModifiedAt.After(images[j].ModifiedAt)
		}
		return images[i].Name > images[j].Name
	})
	total := int64(len(images))
	start := page.Offset()
	if start > len(images) {
		start = len(images)
	}
	end := start + page.PerPage
	if end > len(images) {
		end = len(images)
	}
	return images[start:end], page.Info(total), nil
}

func (uc *MediaUseCase) DeleteImage(ctx context.Context, name string) error {
	if name != filepath.Base(name) || !isImageName(name) {
		return domain.ErrMediaNotFound
	}
	return uc.storage.Delete(ctx, name)
}

func isImageName(name string) bool {
	if name == "" || name != filepath.Base(name) {
		return false
	}
	_, ok := supportedImageExtensions[strings.ToLower(filepath.Ext(name))]
	return ok
}

var supportedImageExtensions = map[string]string{
	".jpg":  "jpeg",
	".jpeg": "jpeg",
	".png":  "png",
	".webp": "webp",
	".gif":  "gif",
}
