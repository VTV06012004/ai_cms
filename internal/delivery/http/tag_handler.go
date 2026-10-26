package http

import (
	"api/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	useCase *usecase.TagUseCase
}

func NewTagHandler(uc *usecase.TagUseCase) *TagHandler {
	return &TagHandler{useCase: uc}
}

func (h *TagHandler) GetAll(c *gin.Context) {
	tags, err := h.useCase.GetAllTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách thẻ"})
		return
	}
	responses := make([]tagResponse, 0, len(tags))
	for _, tag := range tags {
		responses = append(responses, toTagResponse(tag))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses})
}
