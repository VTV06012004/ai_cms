package http

import (
	"github.com/VTV06012004/ai_cms/internal/usecase"
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
	logs, info, err := h.useCase.GetAllAuditLogs(c.Request.Context(), parsePage(c))
	if err != nil {
		writeError(c, err)
		return
	}

	responses := make([]auditLogResponse, 0, len(logs))
	for _, log := range logs {
		responses = append(responses, toAuditLogResponse(log))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses, "pagination": toPaginationResponse(info)})
}
