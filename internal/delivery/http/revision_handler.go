package http

import (
	"context"
	"net/http"

	"github.com/VTV06012004/ai_cms/internal/domain"

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
	postID, ok := parseIDParam(c, "postId")
	if !ok {
		return
	}

	userIDValue, exists := c.Get("user_id")
	userID, validUserID := userIDValue.(uint)
	if !exists || !validUserID {
		writeError(c, errMissingUserIdentity)
		return
	}
	roleValue, exists := c.Get("user_role")
	userRole, validRole := roleValue.(string)
	if !exists || !validRole {
		writeError(c, errMissingRoleIdentity)
		return
	}

	revisions, err := h.useCase.GetByPostID(c.Request.Context(), postID, userID, userRole)
	if err != nil {
		writeError(c, err)
		return
	}
	responses := make([]revisionResponse, 0, len(revisions))
	for _, revision := range revisions {
		responses = append(responses, toRevisionResponse(revision))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func (h *RevisionHandler) Restore(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)
	roleValue, _ := c.Get("user_role")
	role, _ := roleValue.(string)

	post, err := h.useCase.RestoreRevision(c.Request.Context(), id, userID, role, c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã khôi phục bài viết về phiên bản được chọn thành công", "data": toPostResponse(*post)})
}

func (h *RevisionHandler) Compare(c *gin.Context) {
	fromID, err := parseIDValue(c.Query("from"))
	if err != nil {
		writeError(c, err)
		return
	}
	var toID uint
	if to := c.Query("to"); to != "" && to != "current" {
		toID, err = parseIDValue(to)
		if err != nil {
			writeError(c, err)
			return
		}
	}

	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)
	roleValue, _ := c.Get("user_role")
	role, _ := roleValue.(string)

	comparison, err := h.useCase.CompareRevisions(c.Request.Context(), fromID, toID, userID, role)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toRevisionComparisonResponse(*comparison)})
}
