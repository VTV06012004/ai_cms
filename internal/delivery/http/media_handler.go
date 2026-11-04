package http

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"api/internal/domain"
	"api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	useCase *usecase.MediaUseCase
}

func NewMediaHandler(uc *usecase.MediaUseCase) *MediaHandler {
	return &MediaHandler{useCase: uc}
}

func (h *MediaHandler) UploadImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, usecase.MaxImageSize+(64<<10))
	file, err := c.FormFile("file")
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": usecase.ErrImageTooLarge.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy file tải lên"})
		return
	}
	if file.Size > usecase.MaxImageSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": usecase.ErrImageTooLarge.Error()})
		return
	}
	source, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể đọc file tải lên"})
		return
	}
	defer source.Close()
	content, err := io.ReadAll(io.LimitReader(source, usecase.MaxImageSize+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể đọc file tải lên"})
		return
	}
	if len(content) > usecase.MaxImageSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": usecase.ErrImageTooLarge.Error()})
		return
	}
	filename, err := h.useCase.UploadImage(file.Filename, content)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidImage) {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": usecase.ErrInvalidImage.Error()})
			return
		}
		if errors.Is(err, usecase.ErrImageTooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": usecase.ErrImageTooLarge.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu file trên server"})
		return
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Upload ảnh thành công!",
		"url":     scheme + "://" + c.Request.Host + "/uploads/" + filename,
	})
}

func (h *MediaHandler) ServeUploadedFile(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	filename := filepath.Base(c.Param("filepath"))
	if filename == "." || filename == string(filepath.Separator) {
		c.Status(http.StatusNotFound)
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

func requestBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}

func (h *MediaHandler) List(c *gin.Context) {
	files, info, err := h.useCase.ListImages(parsePage(c))
	if err != nil {
		writeError(c, err)
		return
	}
	base := requestBaseURL(c)
	responses := make([]mediaResponse, 0, len(files))
	for _, file := range files {
		responses = append(responses, mediaResponse{
			Name: file.Name, URL: base + "/uploads/" + file.Name,
			Size: file.Size, UploadedAt: file.ModifiedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": responses, "pagination": toPaginationResponse(info)})
}

func (h *MediaHandler) Delete(c *gin.Context) {
	if err := h.useCase.DeleteImage(c.Param("name")); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa ảnh thành công"})
}

var _ = domain.ErrMediaNotFound
