package http

import (
	"api/internal/delivery/http/httperrors"

	"github.com/gin-gonic/gin"
)

var (
	errInvalidRequest        = httperrors.ErrInvalidRequest
	errInvalidID             = httperrors.ErrInvalidID
	errMissingUserIdentity   = httperrors.ErrMissingUserIdentity
	errMissingRoleIdentity   = httperrors.ErrMissingRoleIdentity
	errMissingUploadFile     = httperrors.ErrMissingUploadFile
	errAuthRateLimitExceeded = httperrors.ErrAuthRateLimitExceeded
	errForbidden             = httperrors.ErrForbidden
)

func writeError(c *gin.Context, err error) {
	httperrors.WriteError(c, err)
}

func writeInvalidRequest(c *gin.Context) {
	writeError(c, errInvalidRequest)
}
