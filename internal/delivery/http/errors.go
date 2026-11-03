package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"api/internal/domain"
	"api/internal/usecase"
)

const (
	internalErrorMessage  = "Lỗi hệ thống, vui lòng thử lại sau"
	invalidRequestMessage = "Dữ liệu gửi lên không hợp lệ"
)

type errorRule struct {
	target error
	status int
}

// Body trả về luôn là message của chính lỗi trong bảng này,
// nên chi tiết DB bị bọc bên trong sẽ không lộ ra client.
var errorRules = []errorRule{
	{usecase.ErrInvalidCredentials, http.StatusUnauthorized},

	{domain.ErrPostNotFound, http.StatusNotFound},
	{domain.ErrRevisionNotFound, http.StatusNotFound},
	{domain.ErrUserNotFound, http.StatusNotFound},
	{domain.ErrMenuNotFound, http.StatusNotFound},

	{domain.ErrPostAccessDenied, http.StatusForbidden},
	{domain.ErrPublishNotAllowed, http.StatusForbidden},
	{domain.ErrPostStatusNotAllowed, http.StatusForbidden},
	{domain.ErrCannotChangeOwnRole, http.StatusForbidden},
	{domain.ErrCannotChangeRootAdmin, http.StatusForbidden},

	{domain.ErrPostSlugAlreadyExists, http.StatusConflict},
	{domain.ErrEmailAlreadyExists, http.StatusConflict},
	{domain.ErrCategoryAlreadyExists, http.StatusConflict},

	{domain.ErrPostNotInReview, http.StatusBadRequest},
	{domain.ErrRejectionReasonRequired, http.StatusBadRequest},
	{domain.ErrInvalidPostStatus, http.StatusBadRequest},
	{domain.ErrInvalidRole, http.StatusBadRequest},
	{domain.ErrUserNameRequired, http.StatusBadRequest},
	{domain.ErrInvalidEmail, http.StatusBadRequest},
	{domain.ErrPostTitleRequired, http.StatusBadRequest},
	{domain.ErrPostSlugRequired, http.StatusBadRequest},
	{domain.ErrPostCategoryRequired, http.StatusBadRequest},
	{domain.ErrCategoryNameRequired, http.StatusBadRequest},
	{domain.ErrCategorySlugRequired, http.StatusBadRequest},
	{domain.ErrMenuNameRequired, http.StatusBadRequest},
	{domain.ErrMenuItemTitleRequired, http.StatusBadRequest},

	{domain.ErrCategoryNotFound, http.StatusNotFound},
	{domain.ErrTagNotFound, http.StatusNotFound},

	{domain.ErrCategoryInUse, http.StatusConflict},
	{domain.ErrCategoryHasChildren, http.StatusConflict},
	{domain.ErrTagAlreadyExists, http.StatusConflict},

	{domain.ErrInvalidCategoryParent, http.StatusBadRequest},
	{domain.ErrTagNameRequired, http.StatusBadRequest},
	{domain.ErrTagSlugRequired, http.StatusBadRequest},
}

func writeError(c *gin.Context, err error) {
	for _, rule := range errorRules {
		if errors.Is(err, rule.target) {
			c.JSON(rule.status, gin.H{"error": rule.target.Error()})
			return
		}
	}
	log.Printf("internal error on %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": internalErrorMessage})
}

func writeInvalidRequest(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": invalidRequestMessage})
}
