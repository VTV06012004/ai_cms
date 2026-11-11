package middleware

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddlewareSetsResponseHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	requestID := response.Header().Get(requestIDHeader)
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(requestID) {
		t.Fatalf("request ID = %q, want 32 lowercase hex characters", requestID)
	}
	if _, err := hex.DecodeString(requestID); err != nil {
		t.Fatalf("request ID is not valid hex: %v", err)
	}
}

func TestStructuredRecoveryReturnsInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID(), StructuredRecovery())
	router.GET("/", func(*gin.Context) {
		panic("test panic")
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
}
