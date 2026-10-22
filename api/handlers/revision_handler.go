package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"api/models"
)

type RevisionHandler struct {
	DB *gorm.DB
}

func NewRevisionHandler(db *gorm.DB) *RevisionHandler {
	return &RevisionHandler{DB: db}
}

// 1. Lấy danh sách lịch sử phiên bản của 1 Post
func (h *RevisionHandler) GetByPostID(c *gin.Context) {
	postID := c.Param("postId")
	var revisions []models.Revision

	if err := h.DB.Preload("User").Where("post_id = ?", postID).Order("created_at desc").Find(&revisions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy lịch sử phiên bản"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": revisions})
}

// 2. Khôi phục bài viết về một phiên bản cũ (Restore)
func (h *RevisionHandler) Restore(c *gin.Context) {
	revisionID := c.Param("id")

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

	// Lấy thông tin bản Revision
	var revision models.Revision
	if err := h.DB.First(&revision, revisionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy bản khôi phục này"})
		return
	}

	// Lấy bài viết hiện tại
	var post models.Post
	if err := h.DB.First(&post, revision.PostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy bài viết"})
		return
	}

	// Kiểm tra quyền: Author chỉ khôi phục bài của mình
	if userRole == "author" && post.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn chỉ có thể khôi phục bài viết của chính mình"})
		return
	}

	// Tạo lại 1 bản snapshot cho dữ liệu HIỆN TẠI trước khi bị đè bởi bản khôi phục
	currentRevision := models.Revision{
		PostID:     post.ID,
		Title:      post.Title,
		Slug:       post.Slug,
		Content:    post.Content,
		Status:     post.Status,
		CategoryID: post.CategoryID,
		UserID:     userID,
	}
	h.DB.Create(&currentRevision)

	// Đè thông tin từ Revision được chọn vào Bài viết hiện tại
	post.Title = revision.Title
	post.Slug = revision.Slug
	post.Content = revision.Content
	post.CategoryID = revision.CategoryID

	if err := h.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể khôi phục bài viết"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã khôi phục bài viết về phiên bản được chọn thành công", "data": post})
}
