package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct{}

func NewMediaHandler() *MediaHandler {
	return &MediaHandler{}
}

func (h *MediaHandler) UploadImage(c *gin.Context) {
	// Lấy file từ request form key là "file"
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy file tải lên"})
		return
	}

	// Đặt tên file độc đáo tránh trùng lặp bằng timestamp
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
	dst := filepath.Join("uploads", filename)

	// Lưu file vào thư mục uploads/
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu file trên server"})
		return
	}

	// Trả về URL để frontend hiển thị ảnh
	// Giả sử backend chạy ở cổng 8080
	fileUrl := fmt.Sprintf("http://localhost:8080/uploads/%s", filename)

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload ảnh thành công!",
		"url":     fileUrl,
	})
}
