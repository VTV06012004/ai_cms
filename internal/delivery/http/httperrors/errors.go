package httperrors

import (
	"errors"
	"log"
	"net/http"

	"api/internal/domain"
	"api/internal/infrastructure/security"
	"api/internal/usecase"

	"github.com/gin-gonic/gin"
)

const (
	internalErrorMessage  = "Lỗi hệ thống, vui lòng thử lại sau"
	invalidRequestMessage = "Dữ liệu gửi lên không hợp lệ"
)

var (
	ErrInvalidRequest        = errors.New(invalidRequestMessage)
	ErrInvalidID             = errors.New("ID không hợp lệ")
	ErrMissingUserIdentity   = errors.New("Không tìm thấy thông tin người dùng")
	ErrMissingRoleIdentity   = errors.New("Không tìm thấy thông tin phân quyền")
	ErrMissingUploadFile     = errors.New("Không tìm thấy file tải lên")
	ErrAuthRateLimitExceeded = errors.New("quá nhiều yêu cầu xác thực, vui lòng thử lại sau")
	ErrMissingToken          = errors.New("Không tìm thấy token xác thực")
	ErrMalformedToken        = errors.New("Định dạng token không hợp lệ")
	ErrForbidden             = errors.New("Bạn không có quyền thực hiện hành động này")
	ErrInvalidRole           = errors.New("Vai trò người dùng không hợp lệ")
)

type errorRule struct {
	target error
	status int
}

var errorRules = []errorRule{
	{usecase.ErrInvalidCredentials, http.StatusUnauthorized},
	{ErrMissingUserIdentity, http.StatusUnauthorized},
	{ErrMissingRoleIdentity, http.StatusForbidden},
	{ErrAuthRateLimitExceeded, http.StatusTooManyRequests},
	{ErrInvalidRequest, http.StatusBadRequest},
	{ErrInvalidID, http.StatusBadRequest},
	{ErrMissingUploadFile, http.StatusBadRequest},
	{usecase.ErrInvalidImage, http.StatusUnsupportedMediaType},
	{usecase.ErrImageTooLarge, http.StatusRequestEntityTooLarge},
	{ErrMissingToken, http.StatusUnauthorized},
	{ErrMalformedToken, http.StatusUnauthorized},
	{security.ErrInvalidToken, http.StatusUnauthorized},
	{ErrForbidden, http.StatusForbidden},
	{ErrInvalidRole, http.StatusForbidden},

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
	{domain.ErrInvalidScheduleTime, http.StatusBadRequest},
	{domain.ErrPostNotDue, http.StatusBadRequest},
	{domain.ErrMediaNotFound, http.StatusNotFound},
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
	{domain.ErrRevisionPostMismatch, http.StatusBadRequest},
	{domain.ErrCategoryNotFound, http.StatusNotFound},
	{domain.ErrTagNotFound, http.StatusNotFound},
	{domain.ErrCategoryInUse, http.StatusConflict},
	{domain.ErrCategoryHasChildren, http.StatusConflict},
	{domain.ErrTagAlreadyExists, http.StatusConflict},
	{domain.ErrInvalidCategoryParent, http.StatusBadRequest},
	{domain.ErrTagNameRequired, http.StatusBadRequest},
	{domain.ErrTagSlugRequired, http.StatusBadRequest},
}

func WriteError(c *gin.Context, err error) {
	for _, rule := range errorRules {
		if errors.Is(err, rule.target) {
			c.JSON(rule.status, gin.H{"error": rule.target.Error()})
			return
		}
	}
	log.Printf("internal error on %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": internalErrorMessage})
}
