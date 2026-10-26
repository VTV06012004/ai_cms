package http

import (
	"api/internal/usecase"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	useCase *usecase.MediaUseCase
}

func NewMediaHandler(uc *usecase.MediaUseCase) *MediaHandler {
	return &MediaHandler{useCase: uc}
}

func (h *MediaHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy file tải lên"})
		return
	}
	source, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể đọc file tải lên"})
		return
	}
	defer source.Close()
	content, err := io.ReadAll(source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể đọc file tải lên"})
		return
	}
	filename, err := h.useCase.UploadImage(file.Filename, content)
	if err != nil {
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
