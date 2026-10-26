package http

import (
	"api/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	useCase *usecase.AuditUseCase
}

func NewAuditHandler(uc *usecase.AuditUseCase) *AuditHandler {
	return &AuditHandler{useCase: uc}
}

func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	logs, err := h.useCase.GetAllAuditLogs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách nhật ký hệ thống"})
		return
	}

	responses := make([]auditLogResponse, 0, len(logs))
	for _, log := range logs {
		responses = append(responses, toAuditLogResponse(log))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses})
}
