package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"api/internal/domain"
	"api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	useCase       *usecase.MediaUseCase
	publicBaseURL string
}

// NewMediaHandler tạo handler ảnh. publicBaseURL là gốc URL công khai của API (ví dụ
// https://cms.example.com); để trống thì URL ảnh là đường dẫn tương đối /uploads/...
// URL không bao giờ được dựng từ header Host của request, vì client có thể gửi giá trị giả.
func NewMediaHandler(uc *usecase.MediaUseCase, publicBaseURL string) *MediaHandler {
	return &MediaHandler{useCase: uc, publicBaseURL: strings.TrimRight(publicBaseURL, "/")}
}

func (h *MediaHandler) uploadURL(name string) string {
	return h.publicBaseURL + "/uploads/" + name
}

func (h *MediaHandler) UploadImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, usecase.MaxImageSize+(64<<10))
	file, err := c.FormFile("file")
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(c, usecase.ErrImageTooLarge)
			return
		}
		writeError(c, errMissingUploadFile)
		return
	}
	if file.Size > usecase.MaxImageSize {
		writeError(c, usecase.ErrImageTooLarge)
		return
	}
	source, err := file.Open()
	if err != nil {
		writeError(c, fmt.Errorf("open uploaded file: %w", err))
		return
	}
	defer source.Close()
	content, err := io.ReadAll(io.LimitReader(source, usecase.MaxImageSize+1))
	if err != nil {
		writeError(c, fmt.Errorf("read uploaded file: %w", err))
		return
	}
	if len(content) > usecase.MaxImageSize {
		writeError(c, usecase.ErrImageTooLarge)
		return
	}
	filename, err := h.useCase.UploadImage(c.Request.Context(), file.Filename, content)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Upload ảnh thành công!",
		"url":     h.uploadURL(filename),
	})
}

func (h *MediaHandler) ServeUploadedFile(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	filename := filepath.Base(c.Param("filepath"))
	if filename == "." || filename == string(filepath.Separator) {
		writeError(c, domain.ErrMediaNotFound)
		return
	}
	c.File(filepath.Join("uploads", filename))
}

type mediaResponse struct {
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (h *MediaHandler) List(c *gin.Context) {
	files, info, err := h.useCase.ListImages(c.Request.Context(), parsePage(c))
	if err != nil {
		writeError(c, err)
		return
	}
	responses := make([]mediaResponse, 0, len(files))
	for _, file := range files {
		responses = append(responses, mediaResponse{
			Name: file.Name, URL: h.uploadURL(file.Name),
			Size: file.Size, UploadedAt: file.ModifiedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": responses, "pagination": toPaginationResponse(info)})
}

func (h *MediaHandler) Delete(c *gin.Context) {
	if err := h.useCase.DeleteImage(c.Request.Context(), c.Param("name")); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa ảnh thành công"})
}
