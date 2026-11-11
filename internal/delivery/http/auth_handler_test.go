package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api/internal/domain"
	"api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type authLoginStub struct {
	response *usecase.LoginResponse
	err      error
}

func (s authLoginStub) Register(context.Context, usecase.RegisterDTO, string) (uint, error) {
	return 0, nil
}

func (s authLoginStub) Login(context.Context, usecase.LoginDTO, string) (*usecase.LoginResponse, error) {
	return s.response, s.err
}

type authSessionStub struct {
	issuedUserID uint
	refreshToken string
	refresh      *usecase.RefreshedSession
	refreshErr   error
	revokedToken string
	revokeErr    error
}

func (s *authSessionStub) Issue(_ context.Context, userID uint) (string, error) {
	s.issuedUserID = userID
	return s.refreshToken, nil
}

func (s *authSessionStub) Refresh(_ context.Context, token string) (*usecase.RefreshedSession, error) {
	s.revokedToken = token
	return s.refresh, s.refreshErr
}

func (s *authSessionStub) Revoke(_ context.Context, token string) error {
	s.revokedToken = token
	return s.revokeErr
}

func authTestUser() domain.User {
	return *domain.ReconstituteUser(domain.UserData{
		ID: 7, Name: "Test User", Email: "test@example.com", Role: domain.UserRoleEditor,
	})
}

func TestLoginIssuesHttpOnlyRefreshCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sessions := &authSessionStub{refreshToken: "opaque-refresh-token"}
	handler := NewAuthHandler(
		authLoginStub{response: &usecase.LoginResponse{Token: "access-token", User: authTestUser()}},
		sessions, []string{"http://localhost:3001"}, true,
	)
	router := gin.New()
	router.POST("/api/auth/login", handler.Login)

	request := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"email":"test@example.com","password":"password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Token != "access-token" {
		t.Fatalf("access token = %q", body.Token)
	}
	if sessions.issuedUserID != 7 {
		t.Fatalf("issued user ID = %d, want 7", sessions.issuedUserID)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != refreshCookieName || cookie.Value != "opaque-refresh-token" ||
		!cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteNoneMode ||
		cookie.Path != "/api/auth" {
		t.Fatalf("unexpected refresh cookie: %#v", cookie)
	}
}

func TestRefreshRequiresAllowedOriginAndRotatesCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sessions := &authSessionStub{refresh: &usecase.RefreshedSession{
		AccessToken: "new-access-token", RefreshToken: "new-refresh-token", User: authTestUser(),
	}}
	handler := NewAuthHandler(authLoginStub{}, sessions, []string{"http://localhost:3001"}, false)
	router := gin.New()
	router.POST("/api/auth/refresh", handler.Refresh)

	badRequest := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	badRequest.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "old-refresh-token"})
	badResponse := httptest.NewRecorder()
	router.ServeHTTP(badResponse, badRequest)
	if badResponse.Code != http.StatusForbidden {
		t.Fatalf("untrusted Origin status = %d, want %d", badResponse.Code, http.StatusForbidden)
	}
	if sessions.revokedToken != "" {
		t.Fatal("refresh service was called for an untrusted Origin")
	}

	request := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	request.Header.Set("Origin", "http://localhost:3001")
	request.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "old-refresh-token"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if sessions.revokedToken != "old-refresh-token" {
		t.Fatalf("refresh token sent to service = %q", sessions.revokedToken)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value != "new-refresh-token" {
		t.Fatalf("rotated cookie = %#v", cookies)
	}
	if !strings.Contains(response.Body.String(), "new-access-token") {
		t.Fatalf("response does not contain new access token: %s", response.Body.String())
	}
}

func TestLogoutRevokesRefreshTokenAndClearsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sessions := &authSessionStub{}
	handler := NewAuthHandler(authLoginStub{}, sessions, []string{"http://localhost:3001"}, false)
	router := gin.New()
	router.POST("/api/auth/logout", handler.Logout)

	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.Header.Set("Origin", "http://localhost:3001")
	request.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "active-refresh-token"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if sessions.revokedToken != "active-refresh-token" {
		t.Fatalf("revoked token = %q", sessions.revokedToken)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 || !cookies[0].HttpOnly {
		t.Fatalf("logout cookie was not cleared securely: %#v", cookies)
	}
}
