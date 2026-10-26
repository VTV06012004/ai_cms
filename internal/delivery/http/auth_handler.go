package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"api/internal/usecase"
)

type AuthHandler struct {
	userUseCase *usecase.UserUseCase
}

func NewAuthHandler(uc *usecase.UserUseCase) *AuthHandler {
	return &AuthHandler{userUseCase: uc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var request struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.userUseCase.Register(usecase.RegisterDTO{
		Name: request.Name, Email: request.Email, Password: request.Password,
	}, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Đăng ký tài khoản thành công!", "user_id": userID})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.userUseCase.Login(usecase.LoginDTO{Email: request.Email, Password: request.Password}, c.ClientIP())
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng nhập thành công",
		"token":   res.Token,
		"user": gin.H{
			"id":    res.User.Data().ID,
			"name":  res.User.Data().Name,
			"email": string(res.User.Data().Email),
			"role":  string(res.User.Data().Role),
		},
	})
}
