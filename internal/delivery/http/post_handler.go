package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"api/internal/usecase"
)

type PostHandler struct {
	useCase *usecase.PostUseCase
}

func NewPostHandler(uc *usecase.PostUseCase) *PostHandler {
	return &PostHandler{useCase: uc}
}

func (h *PostHandler) GetAll(c *gin.Context) {
	userRoleValue, _ := c.Get("user_role")
	userRole, _ := userRoleValue.(string)

	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)

	posts, info, err := h.useCase.GetAllPosts(c.Request.Context(), userRole, userID, parsePage(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toPostResponses(posts), "pagination": toPaginationResponse(info)})
}

func (h *PostHandler) GetByID(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userRoleValue, _ := c.Get("user_role")
	userRole, _ := userRoleValue.(string)
	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)

	post, err := h.useCase.GetPostByID(c.Request.Context(), id, userRole, userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toPostResponse(*post)})
}

func (h *PostHandler) Create(c *gin.Context) {
	var request struct {
		Title       string     `json:"title" binding:"required"`
		Slug        string     `json:"slug" binding:"required"`
		Content     string     `json:"content"`
		Status      string     `json:"status"`
		CategoryID  uint       `json:"category_id"`
		TagIDs      []uint     `json:"tag_ids"`
		ScheduledAt *time.Time `json:"scheduled_at"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		writeInvalidRequest(c)
		return
	}

	userIDValue, exists := c.Get("user_id")
	userID, validUserID := userIDValue.(uint)
	if !exists || !validUserID {
		writeError(c, errMissingUserIdentity)
		return
	}

	userRoleValue, roleExists := c.Get("user_role")
	userRole, validRole := userRoleValue.(string)
	if !roleExists || !validRole {
		writeError(c, errMissingRoleIdentity)
		return
	}

	post, err := h.useCase.CreatePost(c.Request.Context(), userID, userRole, usecase.CreatePostDTO{
		Title: request.Title, Slug: request.Slug, Content: request.Content,
		Status: request.Status, CategoryID: request.CategoryID, TagIDs: request.TagIDs,
		ScheduledAt: request.ScheduledAt,
	}, c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": toPostResponse(*post)})
}

func (h *PostHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userIDValue, exists := c.Get("user_id")
	userID, _ := userIDValue.(uint)
	if !exists {
		writeError(c, errMissingUserIdentity)
		return
	}

	userRoleValue, _ := c.Get("user_role")
	userRole, _ := userRoleValue.(string)

	var request struct {
		Title       string     `json:"title" binding:"required"`
		Slug        string     `json:"slug" binding:"required"`
		Content     string     `json:"content"`
		Status      string     `json:"status"`
		CategoryID  uint       `json:"category_id"`
		TagIDs      []uint     `json:"tag_ids"`
		ScheduledAt *time.Time `json:"scheduled_at"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		writeInvalidRequest(c)
		return
	}

	updatedPost, err := h.useCase.UpdatePost(c.Request.Context(), id, userID, userRole, usecase.CreatePostDTO{
		Title: request.Title, Slug: request.Slug, Content: request.Content,
		Status: request.Status, CategoryID: request.CategoryID, TagIDs: request.TagIDs,
		ScheduledAt: request.ScheduledAt,
	}, c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toPostResponse(*updatedPost)})
}

func (h *PostHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)
	userRoleValue, _ := c.Get("user_role")
	userRole, _ := userRoleValue.(string)

	if err := h.useCase.DeletePost(c.Request.Context(), id, userID, userRole, c.ClientIP()); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa bài viết thành công"})
}

func (h *PostHandler) Moderate(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req struct {
		Status          string `json:"status" binding:"required"`
		RejectionReason string `json:"rejection_reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		writeInvalidRequest(c)
		return
	}

	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)
	post, err := h.useCase.ModeratePost(c.Request.Context(), id, userID, req.Status, req.RejectionReason, c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã kiểm duyệt bài viết thành công", "data": toPostResponse(*post)})
}
