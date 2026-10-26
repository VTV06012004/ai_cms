package http

import (
	"api/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RevisionHandler struct {
	useCase *usecase.RevisionUseCase
}

func NewRevisionHandler(uc *usecase.RevisionUseCase) *RevisionHandler {
	return &RevisionHandler{useCase: uc}
}

func (h *RevisionHandler) GetByPostID(c *gin.Context) {
	postId, err := strconv.ParseUint(c.Param("postId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID bài viết không hợp lệ"})
		return
	}

	revisions, err := h.useCase.GetByPostID(uint(postId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	post, err := h.useCase.RestoreRevision(uint(id), userID, role, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã khôi phục bài viết về phiên bản được chọn thành công", "data": toPostResponse(*post)})
}
