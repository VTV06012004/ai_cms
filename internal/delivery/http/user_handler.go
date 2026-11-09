package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"api/internal/usecase"
)

type UserHandler struct {
	useCase *usecase.UserUseCase
}

func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{useCase: uc}
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, info, err := h.useCase.GetAllUsers(c.Request.Context(), parsePage(c))
	if err != nil {
		writeError(c, err)
		return
	}
	responses := make([]userResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, toUserResponse(user))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses, "pagination": toPaginationResponse(info)})
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		writeInvalidRequest(c)
		return
	}

	userIDValue, _ := c.Get("user_id")
	currentUserID, _ := userIDValue.(uint)

	if err := h.useCase.UpdateUserRole(c.Request.Context(), currentUserID, id, input.Role, c.ClientIP()); err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật vai trò người dùng thành công"})
}
