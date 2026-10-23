package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"api/models"
	"api/utils"
)

type PostHandler struct {
	DB *gorm.DB
}

func NewPostHandler(db *gorm.DB) *PostHandler {
	return &PostHandler{DB: db}
}

type PostInput struct {
	Title      string `json:"title" binding:"required"`
	Slug       string `json:"slug" binding:"required"`
	Content    string `json:"content"`
	Status     string `json:"status"`
	CategoryID uint   `json:"category_id"`
	TagIDs     []uint `json:"tag_ids"`
}

// 1. Lấy danh sách Posts
func (h *PostHandler) GetAll(c *gin.Context) {
	var posts []models.Post

	// 🛡️ Lấy thông tin role và user_id từ context (được gán qua middleware xác thực)
	userRoleValue, roleExists := c.Get("user_role")
	userRole, validRole := userRoleValue.(string)

	userIDValue, userIDExists := c.Get("user_id")
	userID, validUserID := userIDValue.(uint)

	query := h.DB.Preload("Author").Preload("Category").Preload("Tags")

	// 🛡️ NẾU LÀ AUTHOR, CHỈ TRẢ VỀ: bài của chính mình HOẶC bài đã published của người khác
	if roleExists && validRole && userRole == "author" && userIDExists && validUserID {
		query = query.Where("author_id = ? OR status = ?", userID, "published")
	}
	// Nếu là admin hoặc editor, truy vấn toàn bộ không cần điều kiện lọc

	if err := query.Order("created_at DESC").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": posts})
}

// 2. Lấy 1 Post theo ID
func (h *PostHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	var post models.Post
	if err := h.DB.Preload("Author").Preload("Category").Preload("Tags").First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy bài viết"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": post})
}

// 3. Tạo Post mới (Create)
func (h *PostHandler) Create(c *gin.Context) {
	var input PostInput
	if err := c.ShouldBindJSON(&input); err != nil {
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

	if userRole == "author" && input.Status == "published" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Author không có quyền xuất bản trực tiếp!"})
		return
	}

	if input.Status == "" {
		input.Status = "draft"
	}

	post := models.Post{
		Title:      input.Title,
		Slug:       input.Slug,
		Content:    input.Content,
		Status:     input.Status,
		CategoryID: input.CategoryID,
		AuthorID:   userID,
	}

	if err := h.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.TagIDs) > 0 {
		var tags []models.Tag
		h.DB.Find(&tags, input.TagIDs)
		h.DB.Model(&post).Association("Tags").Replace(tags)
	}

	// 📸 [REVISION] Tự động lưu phiên bản gốc (v1) ngay khi tạo bài mới
	initialRevision := models.Revision{
		PostID:     post.ID,
		Title:      post.Title,
		Slug:       post.Slug,
		Content:    post.Content,
		Status:     post.Status,
		CategoryID: post.CategoryID,
		UserID:     userID,
	}
	h.DB.Create(&initialRevision)

	// 🛡️ [AUDIT LOG] Ghi log tạo bài
	utils.RecordAuditLog(
		h.DB, c,
		"CREATE_POST",
		fmt.Sprintf("Bài viết ID: %d", post.ID),
		fmt.Sprintf("Đã tạo bài viết: '%s'", post.Title),
	)

	h.DB.Preload("Author").Preload("Category").Preload("Tags").First(&post, post.ID)
	c.JSON(http.StatusCreated, gin.H{"data": post})
}

// 4. Cập nhật Post (Update)
func (h *PostHandler) Update(c *gin.Context) {
	userIDValue, userIDExists := c.Get("user_id")
	userID, validUserID := userIDValue.(uint)
	if !userIDExists || !validUserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin người dùng"})
		return
	}

	userRoleValue, userRoleExists := c.Get("user_role")
	userRole, validUserRole := userRoleValue.(string)
	if !userRoleExists || !validUserRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "Không tìm thấy thông tin phân quyền"})
		return
	}

	id := c.Param("id")
	var post models.Post
	if err := h.DB.Preload("Tags").First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy bài viết"})
		return
	}
	if userRole == "author" && post.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn chỉ có thể chỉnh sửa bài viết của mình"})
		return
	}

	var input PostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if userRole == "author" && input.Status == "published" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Author không có quyền chuyển sang Published!"})
		return
	}

	// 📸 [REVISION QUAN TRỌNG]: Lưu lại TRẠNG THÁI CŨ trước khi ghi đè dữ liệu mới
	oldRevision := models.Revision{
		PostID:     post.ID,
		Title:      post.Title,      // Lưu tiêu đề cũ
		Slug:       post.Slug,       // Lưu slug cũ
		Content:    post.Content,    // Lưu nguyên vẹn "đống chữ và ảnh" ban đầu vào lịch sử
		Status:     post.Status,     // Lưu trạng thái cũ
		CategoryID: post.CategoryID, // Lưu danh mục cũ
		UserID:     userID,          // Người thực hiện lần sửa này
	}
	h.DB.Create(&oldRevision)

	// Sau khi đã lưu bản cũ vào revisions, tiến hành cập nhật dữ liệu mới vào Post
	post.Title = input.Title
	post.Slug = input.Slug
	post.Content = input.Content
	post.Status = input.Status
	post.CategoryID = input.CategoryID

	if err := h.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể cập nhật bài viết"})
		return
	}

	var tags []models.Tag
	if len(input.TagIDs) > 0 {
		h.DB.Find(&tags, input.TagIDs)
	}
	h.DB.Model(&post).Association("Tags").Replace(tags)

	// 🛡️ [AUDIT LOG] Ghi log cập nhật bài
	utils.RecordAuditLog(
		h.DB, c,
		"UPDATE_POST",
		fmt.Sprintf("Bài viết ID: %d", post.ID),
		fmt.Sprintf("Đã cập nhật bài viết: '%s'", post.Title),
	)

	h.DB.Preload("Author").Preload("Category").Preload("Tags").First(&post, post.ID)
	c.JSON(http.StatusOK, gin.H{"data": post})
}

// 5. Xóa Post (Delete)
func (h *PostHandler) Delete(c *gin.Context) {
	userIDValue, userIDExists := c.Get("user_id")
	userID, validUserID := userIDValue.(uint)
	if !userIDExists || !validUserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin người dùng"})
		return
	}

	userRoleValue, userRoleExists := c.Get("user_role")
	userRole, validUserRole := userRoleValue.(string)
	if !userRoleExists || !validUserRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "Không tìm thấy thông tin phân quyền"})
		return
	}

	id := c.Param("id")
	var post models.Post
	if err := h.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy bài viết"})
		return
	}

	if userRole == "author" && post.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn chỉ có thể xóa bài viết của mình"})
		return
	}

	postTitle := post.Title
	postId := post.ID

	if err := h.DB.Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa bài viết"})
		return
	}

	// 🛡️ [AUDIT LOG] Ghi log xóa bài
	utils.RecordAuditLog(
		h.DB, c,
		"DELETE_POST",
		fmt.Sprintf("Bài viết ID: %d", postId),
		fmt.Sprintf("Đã xóa bài viết: '%s'", postTitle),
	)

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa bài viết thành công"})
}

// Admin cấp quyền duyệt bài cho user
func (h *PostHandler) ToggleApprovePermission(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		CanApprove bool `json:"can_approve"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	if err := h.DB.Model(&models.User{}).Where("id = ?", userID).Update("can_approve", req.CanApprove).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cập nhật quyền duyệt bài thất bại"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật quyền duyệt bài thành công"})
}

func (h *PostHandler) ModeratePost(c *gin.Context) {
	postID := c.Param("id")

	var req struct {
		Status          string `json:"status" binding:"required"` // 'approved' hoặc 'rejected'
		RejectionReason string `json:"rejection_reason"`          // Bắt buộc nếu là 'rejected'
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	if req.Status != "published" && req.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Trạng thái không hợp lệ"})
		return
	}

	if req.Status == "rejected" && req.RejectionReason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phải nhập lý do từ chối bài viết"})
		return
	}

	var post models.Post
	if err := h.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy bài viết"})
		return
	}

	// 🔒 CHẶN: Chỉ cho phép duyệt/từ chối khi bài viết đang ở trạng thái 'review' (Chờ duyệt)
	if post.Status != "review" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chỉ có thể kiểm duyệt bài viết đang ở trạng thái chờ duyệt (Review)!"})
		return
	}

	// Cập nhật trạng thái và lý do
	post.Status = req.Status
	if req.Status == "published" {
		post.RejectionReason = "" // Xóa lý do cũ nếu có
	} else {
		post.RejectionReason = req.RejectionReason
	}

	if err := h.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể cập nhật trạng thái bài viết vào Database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã kiểm duyệt bài viết thành công", "data": post})
}
