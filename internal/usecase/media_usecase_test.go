package usecase

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"
)

type mediaStorageStub struct {
	name    string
	content []byte
	calls   int
}

func (s *mediaStorageStub) Save(name string, content []byte) error {
	s.name = name
	s.content = append([]byte(nil), content...)
	s.calls++
	return nil
}

func validRasterImages(t *testing.T) map[string][]byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	images := make(map[string][]byte)
	for extension, encode := range map[string]func(*bytes.Buffer) error{
		".png": func(buffer *bytes.Buffer) error { return png.Encode(buffer, img) },
		".jpg": func(buffer *bytes.Buffer) error { return jpeg.Encode(buffer, img, nil) },
		".gif": func(buffer *bytes.Buffer) error {
			return gif.Encode(buffer, img, nil)
		},
	} {
		var buffer bytes.Buffer
		if err := encode(&buffer); err != nil {
			t.Fatalf("encode test %s image: %v", extension, err)
		}
		images[extension] = buffer.Bytes()
	}
	images[".webp"], _ = base64.StdEncoding.DecodeString(
		"UklGRiIAAABXRUJQVlA4IBYAAAAwAQCdASoBAAEAAUAmJaQAA3AA/v89WAAAAA==",
	)
	return images
}

func TestUploadImageAcceptsSupportedRasterImageContent(t *testing.T) {
	for extension, content := range validRasterImages(t) {
		t.Run(extension, func(t *testing.T) {
			storage := &mediaStorageStub{}
			useCase := NewMediaUseCase(storage)

			name, err := useCase.UploadImage("user-supplied"+extension, content)
			if err != nil {
				t.Fatalf("UploadImage rejected valid image: %v", err)
			}
			if !strings.HasSuffix(name, extension) {
				t.Fatalf("stored name %q does not preserve validated extension", name)
			}
			if storage.calls != 1 || !bytes.Equal(storage.content, content) {
				t.Fatal("validated image was not saved exactly once")
			}
		})
	}
}

func TestUploadImageRejectsNonImagesAndExtensionMismatch(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  []byte
	}{
		{name: "html named html", filename: "payload.html", content: []byte("<script>alert(1)</script>")},
		{name: "html disguised as jpg", filename: "payload.jpg", content: []byte("<html>not an image</html>")},
		{name: "svg image", filename: "payload.svg", content: []byte("<svg xmlns=\"http://www.w3.org/2000/svg\"/>")},
		{name: "png bytes named jpg", filename: "payload.jpg", content: validRasterImages(t)[".png"]},
		{name: "invalid jpeg bytes", filename: "payload.jpg", content: []byte{0xff, 0xd8, 0xff}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := &mediaStorageStub{}
			_, err := NewMediaUseCase(storage).UploadImage(test.filename, test.content)
			if err != ErrInvalidImage {
				t.Fatalf("error = %v, want ErrInvalidImage", err)
			}
			if storage.calls != 0 {
				t.Fatal("invalid image content was written to storage")
			}
		})
	}
}

func TestUploadImageRejectsFilesOverFiveMegabytes(t *testing.T) {
	storage := &mediaStorageStub{}
	content := make([]byte, MaxImageSize+1)

	_, err := NewMediaUseCase(storage).UploadImage("large.png", content)
	if err != ErrImageTooLarge {
		t.Fatalf("error = %v, want ErrImageTooLarge", err)
	}
	if storage.calls != 0 {
		t.Fatal("oversized file was written to storage")
	}
}
