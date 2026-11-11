package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VTV06012004/ai_cms/internal/delivery/http/middleware"
	"github.com/VTV06012004/ai_cms/internal/domain"
	"github.com/VTV06012004/ai_cms/internal/infrastructure/security"

	"github.com/gin-gonic/gin"
)

type revisionUseCaseStub struct {
	revisions []domain.Revision
	err       error
	userID    uint
	userRole  string
	postID    uint
	ctx       context.Context
}

type revisionContextKey struct{}

func (s *revisionUseCaseStub) GetByPostID(ctx context.Context, postID, userID uint, userRole string) ([]domain.Revision, error) {
	s.ctx = ctx
	s.postID = postID
	s.userID = userID
	s.userRole = userRole
	return s.revisions, s.err
}

func (*revisionUseCaseStub) RestoreRevision(context.Context, uint, uint, string, string) (*domain.Post, error) {
	return nil, nil
}

func (*revisionUseCaseStub) CompareRevisions(context.Context, uint, uint, uint, string) (*domain.RevisionComparison, error) {
	return nil, nil
}

func TestRevisionHandlerRequiresIdentityAndPassesAuthorizationContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	revision := domain.ReconstituteRevision(domain.RevisionData{
		ID: 3, PostID: 7, Content: "private version",
	})
	stub := &revisionUseCaseStub{revisions: []domain.Revision{*revision}}
	router := gin.New()
	router.GET("/posts/:postId/revisions", func(c *gin.Context) {
		c.Set("user_id", uint(17))
		c.Set("user_role", string(domain.UserRoleAuthor))
		c.Next()
	}, NewRevisionHandler(stub).GetByPostID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/posts/7/revisions", nil).
		WithContext(context.WithValue(context.Background(), revisionContextKey{}, "request-context"))
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if stub.postID != 7 || stub.userID != 17 || stub.userRole != string(domain.UserRoleAuthor) {
		t.Fatalf("use case received post=%d user=%d role=%q", stub.postID, stub.userID, stub.userRole)
	}
	if got := stub.ctx.Value(revisionContextKey{}); got != "request-context" {
		t.Fatalf("use case context value = %v, want request context", got)
	}
	if !strings.Contains(response.Body.String(), "private version") {
		t.Fatal("authorized post owner did not receive revision data")
	}
}

func TestRevisionHandlerRejectsMissingIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &revisionUseCaseStub{}
	credentials, err := security.NewCredentials("test-secret")
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/posts/:postId/revisions", middleware.AuthMiddleware(credentials), NewRevisionHandler(stub).GetByPostID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/posts/7/revisions", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if stub.postID != 0 {
		t.Fatal("handler called use case without authenticated identity")
	}
}

func TestRevisionHandlerDoesNotExposeContentWhenUseCaseDeniesAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &revisionUseCaseStub{err: domain.ErrPostAccessDenied}
	router := gin.New()
	router.GET("/posts/:postId/revisions", func(c *gin.Context) {
		c.Set("user_id", uint(99))
		c.Set("user_role", string(domain.UserRoleAuthor))
		c.Next()
	}, NewRevisionHandler(stub).GetByPostID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/posts/7/revisions", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
	if strings.Contains(response.Body.String(), "private version") {
		t.Fatal("denied response exposed revision content")
	}
}

func TestRevisionHandlerReturnsNotFoundForMissingPost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &revisionUseCaseStub{err: domain.ErrPostNotFound}
	router := gin.New()
	router.GET("/posts/:postId/revisions", func(c *gin.Context) {
		c.Set("user_id", uint(99))
		c.Set("user_role", string(domain.UserRoleAuthor))
		c.Next()
	}, NewRevisionHandler(stub).GetByPostID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/posts/7/revisions", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestRevisionHandlerHidesUnexpectedDatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &revisionUseCaseStub{err: errors.New("sensitive database detail")}
	router := gin.New()
	router.GET("/posts/:postId/revisions", func(c *gin.Context) {
		c.Set("user_id", uint(99))
		c.Set("user_role", string(domain.UserRoleAuthor))
		c.Next()
	}, NewRevisionHandler(stub).GetByPostID)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/posts/7/revisions", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if strings.Contains(response.Body.String(), "sensitive database detail") {
		t.Fatal("unexpected database error was exposed")
	}
}
