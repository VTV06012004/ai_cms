package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAuthRateLimiterLimitsByIPAndEmail(t *testing.T) {
	now := time.Now()

	t.Run("IP limit", func(t *testing.T) {
		limiter := newAuthRateLimiter()
		for i := 0; i < authRateLimit; i++ {
			email := "user" + string(rune('a'+i)) + "@example.com"
			if !limiter.allow("192.0.2.1", email, now) {
				t.Fatalf("request %d unexpectedly rate limited", i+1)
			}
		}
		if limiter.allow("192.0.2.1", "another@example.com", now) {
			t.Fatal("request beyond IP limit was allowed")
		}
		if !limiter.allow("192.0.2.1", "another@example.com", now.Add(authRateWindow)) {
			t.Fatal("request was still limited after the window expired")
		}
	})

	t.Run("email limit", func(t *testing.T) {
		limiter := newAuthRateLimiter()
		for i := 0; i < authRateLimit; i++ {
			ip := "192.0.2." + string(rune('1'+i))
			if !limiter.allow(ip, "user@example.com", now) {
				t.Fatalf("request %d unexpectedly rate limited", i+1)
			}
		}
		if limiter.allow("198.51.100.1", " USER@example.com ", now) {
			t.Fatal("email limit was bypassed with another IP or casing")
		}
	})
}

func TestAuthIPRateLimitMiddlewareRejectsExcessRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &AuthHandler{rateLimiter: newAuthRateLimiter()}
	router := gin.New()
	router.POST("/api/auth/login", handler.RateLimitIP(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for i := 0; i < authRateLimit; i++ {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/auth/login", nil))
		if response.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want %d", i+1, response.Code, http.StatusNoContent)
		}
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/auth/login", nil))
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("excess request status = %d, want %d", response.Code, http.StatusTooManyRequests)
	}
}
