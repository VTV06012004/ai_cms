package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"api/internal/domain"

	"github.com/gin-gonic/gin"
)

type revisionUseCase interface {
	GetByPostID(context.Context, uint, uint, string) ([]domain.Revision, error)
	RestoreRevision(context.Context, uint, uint, string, string) (*domain.Post, error)
	CompareRevisions(context.Context, uint, uint, uint, string) (*domain.RevisionComparison, error)
}

type RevisionHandler struct {
	useCase revisionUseCase
}

func NewRevisionHandler(uc revisionUseCase) *RevisionHandler {
	return &RevisionHandler{useCase: uc}
}

func (h *RevisionHandler) GetByPostID(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("postId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID bài viết không hợp lệ"})
		return
	}

	userIDValue, exists := c.Get("user_id")
	userID, validUserID := userIDValue.(uint)
	if !exists || !validUserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin người dùng"})
		return
	}
	roleValue, exists := c.Get("user_role")
	userRole, validRole := roleValue.(string)
	if !exists || !validRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "Không tìm thấy thông tin phân quyền"})
		return
	}

	revisions, err := h.useCase.GetByPostID(c.Request.Context(), uint(postID), userID, userRole)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPostNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy bài viết"})
		case errors.Is(err, domain.ErrPostAccessDenied):
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền xem lịch sử bài viết này"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tải lịch sử bài viết"})
		}
		return
	}
	responses := make([]revisionResponse, 0, len(revisions))
	for _, revision := range revisions {
		responses = append(responses, toRevisionResponse(revision))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func (h *RevisionHandler) Restore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID revision không hợp lệ"})
		return
	}

	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)
	roleValue, _ := c.Get("user_role")
	role, _ := roleValue.(string)

	post, err := h.useCase.RestoreRevision(c.Request.Context(), uint(id), userID, role, c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã khôi phục bài viết về phiên bản được chọn thành công", "data": toPostResponse(*post)})
}

func (h *RevisionHandler) Compare(c *gin.Context) {
	fromID, err := strconv.ParseUint(c.Query("from"), 10, 32)
	if err != nil || fromID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID phiên bản gốc không hợp lệ"})
		return
	}
	var toID uint64
	if to := c.Query("to"); to != "" && to != "current" {
		toID, err = strconv.ParseUint(to, 10, 32)
		if err != nil || toID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID phiên bản so sánh không hợp lệ"})
			return
		}
	}

	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)
	roleValue, _ := c.Get("user_role")
	role, _ := roleValue.(string)

	comparison, err := h.useCase.CompareRevisions(c.Request.Context(), uint(fromID), uint(toID), userID, role)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toRevisionComparisonResponse(*comparison)})
}
