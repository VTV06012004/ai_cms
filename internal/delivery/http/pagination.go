package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/VTV06012004/ai_cms/internal/domain"
)

type paginationResponse struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// parsePage đọc ?page= và ?per_page=; giá trị sai hoặc thiếu dùng mặc định.
func parsePage(c *gin.Context) domain.PageRequest {
	page, _ := strconv.Atoi(c.Query("page"))
	perPage, _ := strconv.Atoi(c.Query("per_page"))
	return domain.NewPageRequest(page, perPage)
}

func toPaginationResponse(info domain.PageInfo) paginationResponse {
	return paginationResponse{
		Page:       info.Page,
		PerPage:    info.PerPage,
		Total:      info.Total,
		TotalPages: info.TotalPages(),
	}
}
