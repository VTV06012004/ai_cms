package http

import (
	"net/http"

	"api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	useCase *usecase.CategoryUseCase
}

func NewCategoryHandler(uc *usecase.CategoryUseCase) *CategoryHandler {
	return &CategoryHandler{useCase: uc}
}

type categoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
	ParentID    *uint  `json:"parent_id"`
}

func (r categoryRequest) toInput() usecase.CategoryInputDTO {
	return usecase.CategoryInputDTO{
		Name: r.Name, Slug: r.Slug, Description: r.Description, ParentID: r.ParentID,
	}
}

func (h *CategoryHandler) GetAll(c *gin.Context) {
	categories, err := h.useCase.GetAllCategories()
	if err != nil {
		writeError(c, err)
		return
	}
	responses := make([]categoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, toCategoryResponse(category))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var request categoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeInvalidRequest(c)
		return
	}
	category, err := h.useCase.CreateCategory(currentUserID(c), request.toInput(), c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": toCategoryResponse(*category)})
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var request categoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeInvalidRequest(c)
		return
	}
	category, err := h.useCase.UpdateCategory(currentUserID(c), id, request.toInput(), c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toCategoryResponse(*category)})
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.useCase.DeleteCategory(currentUserID(c), id, c.ClientIP()); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa danh mục thành công"})
}
