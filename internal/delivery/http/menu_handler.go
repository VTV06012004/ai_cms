package http

import (
	"net/http"

	"github.com/VTV06012004/ai_cms/internal/usecase"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct {
	useCase *usecase.MenuUseCase
}

func NewMenuHandler(uc *usecase.MenuUseCase) *MenuHandler {
	return &MenuHandler{useCase: uc}
}

type menuRequest struct {
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

func (r menuRequest) toInput() usecase.MenuInputDTO {
	input := usecase.MenuInputDTO{Name: r.Name, Location: r.Location}
	for _, item := range r.MenuItems {
		input.MenuItems = append(input.MenuItems, usecase.MenuItemInputDTO{
			Title: item.Title, Type: item.Type, TargetID: item.TargetID, URL: item.URL, SortOrder: item.SortOrder,
		})
	}
	return input
}

func (h *MenuHandler) GetAll(c *gin.Context) {
	menus, err := h.useCase.GetAllMenus(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	responses := make([]menuResponse, 0, len(menus))
	for _, menu := range menus {
		responses = append(responses, toMenuResponse(menu))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func (h *MenuHandler) Create(c *gin.Context) {
	var request menuRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeInvalidRequest(c)
		return
	}
	menu, err := h.useCase.CreateMenu(c.Request.Context(), currentUserID(c), request.toInput(), c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": toMenuResponse(*menu)})
}

func (h *MenuHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var request menuRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeInvalidRequest(c)
		return
	}
	menu, err := h.useCase.UpdateMenu(c.Request.Context(), currentUserID(c), id, request.toInput(), c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toMenuResponse(*menu)})
}

func (h *MenuHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.useCase.DeleteMenu(c.Request.Context(), currentUserID(c), id, c.ClientIP()); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa menu thành công"})
}
