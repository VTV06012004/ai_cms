package http

import (
	"net/http"

	"github.com/VTV06012004/ai_cms/internal/usecase"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	useCase *usecase.TagUseCase
}

func NewTagHandler(uc *usecase.TagUseCase) *TagHandler {
	return &TagHandler{useCase: uc}
}

type tagRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

func (h *TagHandler) GetAll(c *gin.Context) {
	tags, err := h.useCase.GetAllTags(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	responses := make([]tagResponse, 0, len(tags))
	for _, tag := range tags {
		responses = append(responses, toTagResponse(tag))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func (h *TagHandler) Create(c *gin.Context) {
	var request tagRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeInvalidRequest(c)
		return
	}
	tag, err := h.useCase.CreateTag(c.Request.Context(), currentUserID(c), request.Name, request.Slug, c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": toTagResponse(tag)})
}

func (h *TagHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var request tagRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeInvalidRequest(c)
		return
	}
	tag, err := h.useCase.UpdateTag(c.Request.Context(), currentUserID(c), id, request.Name, request.Slug, c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toTagResponse(tag)})
}

func (h *TagHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.useCase.DeleteTag(c.Request.Context(), currentUserID(c), id, c.ClientIP()); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa thẻ thành công"})
}
