package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"api/internal/domain"
	"api/internal/infrastructure/security"
	"api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type routePostRepository struct {
	createErr error
	getErr    error
}

func (r *routePostRepository) GetAll(context.Context, domain.PostListCriteria, domain.PageRequest) ([]domain.Post, int64, error) {
	return nil, 0, nil
}

func (r *routePostRepository) GetByID(context.Context, uint) (*domain.Post, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return nil, domain.ErrPostNotFound
}

func (r *routePostRepository) Create(_ context.Context, post *domain.Post, _ []uint) (*domain.Post, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	return post, nil
}

func (*routePostRepository) Update(context.Context, *domain.Post, []uint) (*domain.Post, error) {
	return nil, nil
}

func (*routePostRepository) Delete(context.Context, uint) error { return nil }
func (*routePostRepository) ListDue(context.Context, time.Time, int) ([]domain.Post, error) {
	return nil, nil
}
func (*routePostRepository) PublishIfDue(context.Context, uint, time.Time) (bool, error) {
	return false, nil
}

type routeUnitOfWork struct {
	posts domain.PostRepository
}

func (u routeUnitOfWork) WithinTransaction(_ context.Context, work func(usecase.TransactionRepositories) error) error {
	return work(usecase.TransactionRepositories{Posts: u.posts})
}

func newRouteStatusRouter(t *testing.T, repository *routePostRepository) (*gin.Engine, *security.Credentials) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	credentials, err := security.NewCredentials("route-test-secret")
	if err != nil {
		t.Fatal(err)
	}
	postUseCase := usecase.NewPostUseCase(repository, routeUnitOfWork{posts: repository})
	router := gin.New()
	RegisterRoutes(router, credentials, nil, NewPostHandler(postUseCase), nil, nil, nil, nil, nil, nil, nil)
	return router, credentials
}

func routeToken(t *testing.T, credentials *security.Credentials, role string) string {
	t.Helper()
	token, err := credentials.GenerateToken(7, "author@example.com", role)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func serveRoute(router http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestProtectedPostRouteReturnsUnauthorizedWithoutToken(t *testing.T) {
	router, _ := newRouteStatusRouter(t, &routePostRepository{})
	response := serveRoute(router, http.MethodPost, "/api/posts", "", "")
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestAdminRouteReturnsForbiddenForAuthor(t *testing.T) {
	router, credentials := newRouteStatusRouter(t, &routePostRepository{})
	response := serveRoute(router, http.MethodGet, "/api/admin/users",
		routeToken(t, credentials, string(domain.UserRoleAuthor)), "")
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestPostRouteReturnsNotFoundForMissingPost(t *testing.T) {
	router, credentials := newRouteStatusRouter(t, &routePostRepository{})
	response := serveRoute(router, http.MethodGet, "/api/posts/1",
		routeToken(t, credentials, string(domain.UserRoleAuthor)), "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusNotFound, response.Body.String())
	}
}

func TestPostCreateRouteReturnsConflictForDuplicateSlug(t *testing.T) {
	router, credentials := newRouteStatusRouter(t, &routePostRepository{createErr: domain.ErrPostSlugAlreadyExists})
	response := serveRoute(router, http.MethodPost, "/api/posts",
		routeToken(t, credentials, string(domain.UserRoleAuthor)),
		`{"title":"Post","slug":"duplicate","category_id":1}`)
	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusConflict, response.Body.String())
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error != domain.ErrPostSlugAlreadyExists.Error() {
		t.Fatalf("error = %q, want duplicate slug error", body.Error)
	}
}
