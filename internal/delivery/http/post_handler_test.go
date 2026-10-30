package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"api/internal/domain"

	"github.com/gin-gonic/gin"
)

func TestWritePostMutationErrorReturnsConflictForDuplicateSlug(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/posts", func(c *gin.Context) {
		writePostMutationError(c, domain.ErrPostSlugAlreadyExists)
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/posts", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
	if got, want := response.Body.String(), "{\"error\":\"slug đã tồn tại\"}"; got != want {
		t.Fatalf("response body = %q, want %q", got, want)
	}
}
