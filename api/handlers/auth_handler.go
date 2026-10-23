package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"api/models"
	"api/utils"
)

type AuthHandler struct {
	DB *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"` // admin, editor, author
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi mã hóa mật khẩu"})
		return
	}

	role := input.Role
	if role == "" {
		role = "author"
	}

	user := models.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email đã tồn tại trên hệ thống"})
		return
	}

	// Ghi Audit Log cho hành động Đăng ký tài khoản mới
	utils.RecordAuditLog(h.DB, c, "REGISTER", fmt.Sprintf("User #%d", user.ID), "Đăng ký tài khoản mới: "+user.Email)

	c.JSON(http.StatusCreated, gin.H{"message": "Đăng ký tài khoản thành công!", "user_id": user.ID})
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email hoặc mật khẩu không chính xác"})
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email hoặc mật khẩu không chính xác"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo token"})
		return
	}

	// Gán tạm userID vào Context để hàm RecordAuditLog nhận diện được chính xác ai đang đăng nhập
	c.Set("userID", user.ID)

	// 📝 Ghi Audit Log cho hành động Đăng nhập thành công
	utils.RecordAuditLog(h.DB, c, "LOGIN", fmt.Sprintf("User #%d", user.ID), "Đăng nhập thành công vào hệ thống")

	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng nhập thành công",
		"token":   token,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}
