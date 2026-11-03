package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"api/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

func performRequest(router *gin.Engine, authorization string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestRequireRolesAllowsOnlyListedRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name    string
		role    any
		setRole bool
		want    int
	}{
		{"admin is allowed", "admin", true, http.StatusNoContent},
		{"editor is allowed", "editor", true, http.StatusNoContent},
		{"author is forbidden", "author", true, http.StatusForbidden},
		{"missing role is forbidden", nil, false, http.StatusForbidden},
		{"non-string role is forbidden", 42, true, http.StatusForbidden},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/protected",
				func(c *gin.Context) {
					if test.setRole {
						c.Set("user_role", test.role)
					}
				},
				RequireRoles("admin", "editor"),
				func(c *gin.Context) { c.Status(http.StatusNoContent) },
			)

			if got := performRequest(router, "").Code; got != test.want {
				t.Fatalf("status = %d, want %d", got, test.want)
			}
		})
	}
}

func TestAuthMiddlewareRejectsMissingMalformedAndInvalidTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	credentials, err := security.NewCredentials("test-secret-with-enough-length")
	if err != nil {
		t.Fatal(err)
	}
	other, err := security.NewCredentials("another-secret-with-enough-length")
	if err != nil {
		t.Fatal(err)
	}
	foreignToken, err := other.GenerateToken(1, "a@example.com", "admin")
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET("/protected", AuthMiddleware(credentials), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	headers := map[string]string{
		"missing header":          "",
		"no Bearer prefix":        "Token abc",
		"garbage token":           "Bearer not-a-token",
		"signed with another key": "Bearer " + foreignToken,
	}
	for name, header := range headers {
		if got := performRequest(router, header).Code; got != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, want 401", name, got)
		}
	}
}

func TestAuthMiddlewareStoresIdentityFromValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	credentials, err := security.NewCredentials("test-secret-with-enough-length")
	if err != nil {
		t.Fatal(err)
	}
	token, err := credentials.GenerateToken(7, "editor@example.com", "editor")
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET("/protected", AuthMiddleware(credentials), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		role, _ := c.Get("user_role")
		if userID != uint(7) || role != "editor" {
			t.Errorf("context identity = (%v, %v), want (7, editor)", userID, role)
		}
		c.Status(http.StatusNoContent)
	})

	if got := performRequest(router, "Bearer "+token).Code; got != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", got)
	}
}
