package http

import (
	"api/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct {
	useCase *usecase.MenuUseCase
}

func NewMenuHandler(uc *usecase.MenuUseCase) *MenuHandler {
	return &MenuHandler{useCase: uc}
}

func (h *MenuHandler) GetAll(c *gin.Context) {
	menus, err := h.useCase.GetAllMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	responses := make([]menuResponse, 0, len(menus))
	for _, menu := range menus {
		responses = append(responses, toMenuResponse(menu))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func (h *MenuHandler) Create(c *gin.Context) {
	var request struct {
		Name      string `json:"name" binding:"required"`
		Location  string `json:"location"`
		MenuItems []struct {
			Title     string `json:"title" binding:"required"`
			Type      string `json:"type"`
			TargetID  uint   `json:"target_id"`
			URL       string `json:"url"`
			SortOrder int    `json:"sort_order"`
		} `json:"menu_items"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input := usecase.MenuInputDTO{Name: request.Name, Location: request.Location}
	for _, item := range request.MenuItems {
		input.MenuItems = append(input.MenuItems, usecase.MenuItemInputDTO{
			Title: item.Title, Type: item.Type, TargetID: item.TargetID, URL: item.URL, SortOrder: item.SortOrder,
		})
	}

	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)

	menu, err := h.useCase.CreateMenu(userID, input, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": toMenuResponse(*menu)})
}

func (h *MenuHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return
	}

	userIDValue, _ := c.Get("user_id")
	userID, _ := userIDValue.(uint)

	if err := h.useCase.DeleteMenu(userID, uint(id), c.ClientIP()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa menu thành công"})
}
