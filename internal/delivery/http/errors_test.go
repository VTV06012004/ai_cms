package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"api/internal/domain"
	"api/internal/usecase"
)

func runWriteError(err error) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/x", func(c *gin.Context) { writeError(c, err) })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/x", nil))
	return response
}

func TestWriteErrorMapsBusinessErrorsToStatusCodes(t *testing.T) {
	cases := []struct {
		err    error
		status int
	}{
		{usecase.ErrInvalidCredentials, http.StatusUnauthorized},
		{domain.ErrPostNotFound, http.StatusNotFound},
		{domain.ErrUserNotFound, http.StatusNotFound},
		{domain.ErrPostAccessDenied, http.StatusForbidden},
		{domain.ErrCannotChangeRootAdmin, http.StatusForbidden},
		{domain.ErrPostSlugAlreadyExists, http.StatusConflict},
		{domain.ErrEmailAlreadyExists, http.StatusConflict},
		{domain.ErrCategoryAlreadyExists, http.StatusConflict},
		{domain.ErrInvalidRole, http.StatusBadRequest},
		{domain.ErrCategoryNotFound, http.StatusNotFound},
		{domain.ErrCategoryInUse, http.StatusConflict},
		{domain.ErrInvalidCategoryParent, http.StatusBadRequest},
	}
	for _, tc := range cases {
		response := runWriteError(tc.err)
		if response.Code != tc.status {
			t.Errorf("%v: status = %d, want %d", tc.err, response.Code, tc.status)
		}
	}
}

func TestWriteErrorDoesNotLeakWrappedDetails(t *testing.T) {
	wrapped := fmt.Errorf("insert users: %w: %w", domain.ErrEmailAlreadyExists,
		errors.New("Error 1062: Duplicate entry 'a@b.c' for key 'users.email'"))
	response := runWriteError(wrapped)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", response.Code)
	}
	if strings.Contains(response.Body.String(), "1062") {
		t.Fatalf("response leaked database details: %s", response.Body.String())
	}
}

func TestWriteErrorHidesUnexpectedErrorsBehindGeneric500(t *testing.T) {
	response := runWriteError(errors.New("dial tcp 10.0.0.5:3306: connection refused"))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
	if strings.Contains(response.Body.String(), "10.0.0.5") {
		t.Fatalf("response leaked internal error: %s", response.Body.String())
	}
}
