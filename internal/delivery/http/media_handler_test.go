package http

import (
	"bytes"
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

	"api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type mediaStorageTestStub struct {
	calls int
}

func (*mediaStorageTestStub) Save(string, []byte) error {
	return nil
}

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
	handler := NewMediaHandler(usecase.NewMediaUseCase(storage))
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
	router.GET("/uploads/*filepath", NewMediaHandler(nil).ServeUploadedFile)
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
