package http

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/VTV06012004/ai_cms/internal/domain"
	"github.com/VTV06012004/ai_cms/internal/usecase"

	"github.com/gin-gonic/gin"
)

type mediaStorageTestStub struct{}

func (*mediaStorageTestStub) Save(context.Context, string, []byte) error {
	return nil
}

func (*mediaStorageTestStub) List(context.Context) ([]domain.MediaFile, error) { return nil, nil }

func (*mediaStorageTestStub) Delete(context.Context, string) error { return nil }

func newMultipartUploadRequest(t *testing.T, filename string, content []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func validPNGForHandlerTest(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	var content bytes.Buffer
	if err := png.Encode(&content, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}
	return content.Bytes()
}

func TestUploadImageHandlerAcceptsImageAndRejectsInvalidOrOversizedFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storage := &mediaStorageTestStub{}
	handler := NewMediaHandler(usecase.NewMediaUseCase(storage), "")
	router := gin.New()
	router.POST("/api/upload", handler.UploadImage)

	tests := []struct {
		name       string
		filename   string
		content    []byte
		wantStatus int
	}{
		{name: "valid png", filename: "image.png", content: validPNGForHandlerTest(t), wantStatus: http.StatusOK},
		{name: "html disguised as image", filename: "image.jpg", content: []byte("<html>not an image</html>"), wantStatus: http.StatusUnsupportedMediaType},
		{name: "over five megabytes", filename: "large.png", content: bytes.Repeat([]byte("x"), usecase.MaxImageSize+1), wantStatus: http.StatusRequestEntityTooLarge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, newMultipartUploadRequest(t, test.filename, test.content))
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, test.wantStatus, response.Body.String())
			}
		})
	}
}

func TestServeUploadedFileSetsActiveContentProtectionHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current directory: %v", err)
	}
	tempDirectory := t.TempDir()
	if err := os.Mkdir(filepath.Join(tempDirectory, "uploads"), 0755); err != nil {
		t.Fatalf("create uploads directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDirectory, "uploads", "old.html"), []byte("<script>alert(1)</script>"), 0600); err != nil {
		t.Fatalf("create legacy active-content file: %v", err)
	}
	if err := os.Chdir(tempDirectory); err != nil {
		t.Fatalf("change to temporary directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(workingDirectory); err != nil {
			t.Errorf("restore current directory: %v", err)
		}
	})

	router := gin.New()
	router.GET("/uploads/*filepath", NewMediaHandler(nil, "").ServeUploadedFile)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/uploads/old.html", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if got := response.Header().Get("Content-Security-Policy"); !strings.Contains(got, "sandbox") {
		t.Fatalf("Content-Security-Policy = %q, want sandbox restriction", got)
	}
}

type mediaListStorageStub struct{ mediaStorageTestStub }

func (*mediaListStorageStub) List(context.Context) ([]domain.MediaFile, error) {
	return []domain.MediaFile{{Name: "a.png", Size: 1, ModifiedAt: time.Now()}}, nil
}

func TestMediaURLsUseConfiguredBaseAndIgnoreRequestHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		base       string
		wantPrefix string
	}{
		{name: "configured https base", base: "https://cms.example.com/", wantPrefix: "https://cms.example.com/uploads/"},
		{name: "no base gives a relative path", base: "", wantPrefix: "/uploads/"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewMediaHandler(usecase.NewMediaUseCase(&mediaListStorageStub{}), test.base)
			router := gin.New()
			router.POST("/api/upload", handler.UploadImage)
			router.GET("/api/media", handler.List)

			upload := newMultipartUploadRequest(t, "image.png", validPNGForHandlerTest(t))
			upload.Host = "evil.example"
			upload.Header.Set("X-Forwarded-Host", "evil.example")
			upload.Header.Set("X-Forwarded-Proto", "http")
			uploadResponse := httptest.NewRecorder()
			router.ServeHTTP(uploadResponse, upload)
			if uploadResponse.Code != http.StatusOK {
				t.Fatalf("upload status = %d, body=%s", uploadResponse.Code, uploadResponse.Body.String())
			}
			var uploaded struct {
				URL string `json:"url"`
			}
			if err := json.Unmarshal(uploadResponse.Body.Bytes(), &uploaded); err != nil {
				t.Fatalf("decode upload response: %v", err)
			}
			if !strings.HasPrefix(uploaded.URL, test.wantPrefix) || strings.Contains(uploaded.URL, "evil.example") {
				t.Fatalf("upload url = %q, want prefix %q and nothing from request headers", uploaded.URL, test.wantPrefix)
			}

			list := httptest.NewRequest(http.MethodGet, "/api/media", nil)
			list.Host = "evil.example"
			listResponse := httptest.NewRecorder()
			router.ServeHTTP(listResponse, list)
			if listResponse.Code != http.StatusOK {
				t.Fatalf("list status = %d, body=%s", listResponse.Code, listResponse.Body.String())
			}
			var listed struct {
				Data []struct {
					URL string `json:"url"`
				} `json:"data"`
			}
			if err := json.Unmarshal(listResponse.Body.Bytes(), &listed); err != nil {
				t.Fatalf("decode list response: %v", err)
			}
			if len(listed.Data) != 1 || listed.Data[0].URL != test.wantPrefix+"a.png" {
				t.Fatalf("list data = %+v, want one url %q", listed.Data, test.wantPrefix+"a.png")
			}
		})
	}
}
