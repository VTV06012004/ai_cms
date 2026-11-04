package http

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

var (
	specPathLine = regexp.MustCompile(`^  (/\S*):\s*$`)
	specOpLine   = regexp.MustCompile(`^    (get|post|put|delete|patch):\s*$`)
	routeParam   = regexp.MustCompile(`:([A-Za-z]+)`)
)

func specOperations() map[string]bool {
	operations := map[string]bool{}
	currentPath := ""
	for _, line := range strings.Split(string(openAPISpec), "\n") {
		line = strings.TrimRight(line, "\r")
		if match := specPathLine.FindStringSubmatch(line); match != nil {
			currentPath = match[1]
			continue
		}
		if match := specOpLine.FindStringSubmatch(line); match != nil && currentPath != "" {
			operations[strings.ToUpper(match[1])+" "+currentPath] = true
		}
	}
	return operations
}

func registeredOperations() map[string]bool {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router, nil,
		(*AuthHandler)(nil), (*PostHandler)(nil), (*CategoryHandler)(nil), (*TagHandler)(nil),
		(*MenuHandler)(nil), (*RevisionHandler)(nil), (*AuditHandler)(nil),
		(*MediaHandler)(nil), (*UserHandler)(nil),
	)
	operations := map[string]bool{}
	for _, route := range router.Routes() {
		// Route phục vụ file tĩnh hoặc trang tài liệu không thuộc API JSON.
		if strings.Contains(route.Path, "*") || strings.HasPrefix(route.Path, "/docs") {
			continue
		}
		operations[route.Method+" "+routeParam.ReplaceAllString(route.Path, "{$1}")] = true
	}
	return operations
}

// Tài liệu luôn khớp với route thật: thêm route mà quên viết tài liệu sẽ làm test này fail.
func TestOpenAPISpecMatchesRegisteredRoutes(t *testing.T) {
	documented := specOperations()
	registered := registeredOperations()

	var missingDocs, staleDocs []string
	for operation := range registered {
		if !documented[operation] {
			missingDocs = append(missingDocs, operation)
		}
	}
	for operation := range documented {
		if !registered[operation] {
			staleDocs = append(staleDocs, operation)
		}
	}
	sort.Strings(missingDocs)
	sort.Strings(staleDocs)
	if len(missingDocs) > 0 {
		t.Errorf("route chưa có trong openapi.yaml: %v", missingDocs)
	}
	if len(staleDocs) > 0 {
		t.Errorf("openapi.yaml mô tả route không tồn tại: %v", staleDocs)
	}
}

func TestDocsEndpointsServeSpecAndUI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/docs", ServeDocs)
	router.GET("/docs/openapi.yaml", ServeOpenAPISpec)

	page := httptest.NewRecorder()
	router.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "SwaggerUIBundle") {
		t.Fatalf("/docs returned %d without Swagger UI", page.Code)
	}

	spec := httptest.NewRecorder()
	router.ServeHTTP(spec, httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil))
	if spec.Code != http.StatusOK || !strings.HasPrefix(spec.Body.String(), "openapi: 3.0.3") {
		t.Fatalf("/docs/openapi.yaml returned %d with unexpected body", spec.Code)
	}
}
