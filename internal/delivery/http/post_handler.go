package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"api/internal/domain"
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

	posts, err := h.useCase.GetAllPosts(userRole, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toPostResponses(posts)})
}

func (h *PostHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	userRoleValue, _ := c.Get("user_role")
	userRole, _ := userRoleValue.(string)
	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)

	post, err := h.useCase.GetPostByID(uint(id), userRole, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy bài viết"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toPostResponse(*post)})
}

func (h *PostHandler) Create(c *gin.Context) {
	var request struct {
		Title      string `json:"title" binding:"required"`
		Slug       string `json:"slug" binding:"required"`
		Content    string `json:"content"`
		Status     string `json:"status"`
		CategoryID uint   `json:"category_id"`
		TagIDs     []uint `json:"tag_ids"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDValue, exists := c.Get("user_id")
	userID, validUserID := userIDValue.(uint)
	if !exists || !validUserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin người dùng"})
		return
	}

	userRoleValue, roleExists := c.Get("user_role")
	userRole, validRole := userRoleValue.(string)
	if !roleExists || !validRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "Không tìm thấy thông tin phân quyền"})
		return
	}

	post, err := h.useCase.CreatePost(userID, userRole, usecase.CreatePostDTO{
		Title: request.Title, Slug: request.Slug, Content: request.Content,
		Status: request.Status, CategoryID: request.CategoryID, TagIDs: request.TagIDs,
	}, c.ClientIP())
	if err != nil {
		writePostMutationError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": toPostResponse(*post)})
}

func (h *PostHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	userIDValue, exists := c.Get("user_id")
	userID, _ := userIDValue.(uint)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin người dùng"})
		return
	}

	userRoleValue, _ := c.Get("user_role")
	userRole, _ := userRoleValue.(string)

	var request struct {
		Title      string `json:"title" binding:"required"`
		Slug       string `json:"slug" binding:"required"`
		Content    string `json:"content"`
		Status     string `json:"status"`
		CategoryID uint   `json:"category_id"`
		TagIDs     []uint `json:"tag_ids"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedPost, err := h.useCase.UpdatePost(uint(id), userID, userRole, usecase.CreatePostDTO{
		Title: request.Title, Slug: request.Slug, Content: request.Content,
		Status: request.Status, CategoryID: request.CategoryID, TagIDs: request.TagIDs,
	}, c.ClientIP())
	if err != nil {
		writePostMutationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toPostResponse(*updatedPost)})
}

func writePostMutationError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrPostSlugAlreadyExists) {
		c.JSON(http.StatusConflict, gin.H{"error": domain.ErrPostSlugAlreadyExists.Error()})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func (h *PostHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)
	userRoleValue, _ := c.Get("user_role")
	userRole, _ := userRoleValue.(string)

	if err := h.useCase.DeletePost(uint(id), userID, userRole, c.ClientIP()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa bài viết thành công"})
}

func (h *PostHandler) Moderate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	var req struct {
		Status          string `json:"status" binding:"required"`
		RejectionReason string `json:"rejection_reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)
	post, err := h.useCase.ModeratePost(uint(id), userID, req.Status, req.RejectionReason, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã kiểm duyệt bài viết thành công", "data": toPostResponse(*post)})
}
