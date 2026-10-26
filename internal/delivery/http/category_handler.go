package http

import (
	"api/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	useCase *usecase.CategoryUseCase
}

func NewCategoryHandler(uc *usecase.CategoryUseCase) *CategoryHandler {
	return &CategoryHandler{useCase: uc}
}

func (h *CategoryHandler) GetAll(c *gin.Context) {
	categories, err := h.useCase.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách chuyên mục"})
		return
	}
	responses := make([]categoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, toCategoryResponse(category))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Slug     string `json:"slug" binding:"required"`
		ParentID *uint  `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.useCase.CreateCategory(input.Name, input.Slug, input.ParentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": toCategoryResponse(*category)})
}
