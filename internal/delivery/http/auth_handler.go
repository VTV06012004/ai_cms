package http

import (
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
		writeInvalidRequest(c)
		return
	}

	userID, err := h.userUseCase.Register(c.Request.Context(), usecase.RegisterDTO{
		Name: request.Name, Email: request.Email, Password: request.Password,
	}, c.ClientIP())
	if err != nil {
		writeError(c, err)
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
		writeInvalidRequest(c)
		return
	}

	res, err := h.userUseCase.Login(c.Request.Context(), usecase.LoginDTO{Email: request.Email, Password: request.Password}, c.ClientIP())
	if err != nil {
		writeError(c, err)
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
