package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"api/internal/domain"
)

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestSaveWritesTheContent(t *testing.T) {
	directory := t.TempDir()
	store := NewLocalFileStorage(directory)

	if err := store.Save(context.Background(), "a.png", []byte("image-bytes")); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(directory, "a.png"))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if string(got) != "image-bytes" {
		t.Fatalf("saved content = %q, want %q", got, "image-bytes")
	}
}

func TestSaveRefusesToOverwriteAnExistingFile(t *testing.T) {
	directory := t.TempDir()
	store := NewLocalFileStorage(directory)
	if err := store.Save(context.Background(), "a.png", []byte("first")); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	if err := store.Save(context.Background(), "a.png", []byte("second")); err == nil {
		t.Fatal("second Save with the same name succeeded, want an error")
	}

	got, _ := os.ReadFile(filepath.Join(directory, "a.png"))
	if string(got) != "first" {
		t.Fatalf("file content = %q, want the original %q", got, "first")
	}
}

func TestSaveKeepsPathTraversalNamesInsideTheDirectory(t *testing.T) {
	parent := t.TempDir()
	directory := filepath.Join(parent, "uploads")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	store := NewLocalFileStorage(directory)

	if err := store.Save(context.Background(), "../escaped.png", []byte("x")); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := os.Stat(filepath.Join(parent, "escaped.png")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file was written outside the upload directory (stat error: %v)", err)
	}
	if _, err := os.Stat(filepath.Join(directory, "escaped.png")); err != nil {
		t.Fatalf("file was not written inside the upload directory: %v", err)
	}
}

func TestSaveFailsWhenTheDirectoryDoesNotExist(t *testing.T) {
	store := NewLocalFileStorage(filepath.Join(t.TempDir(), "missing"))

	if err := store.Save(context.Background(), "a.png", []byte("x")); err == nil {
		t.Fatal("Save into a missing directory succeeded, want an error")
	}
}

func TestListReturnsFilesWithSizeAndSkipsDirectories(t *testing.T) {
	directory := t.TempDir()
	store := NewLocalFileStorage(directory)
	if err := store.Save(context.Background(), "a.png", []byte("12345")); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), "b.png", []byte("123")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(directory, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	files, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	sizes := map[string]int64{}
	var names []string
	for _, file := range files {
		sizes[file.Name] = file.Size
		names = append(names, file.Name)
		if file.ModifiedAt.IsZero() {
			t.Errorf("%s has no modification time", file.Name)
		}
	}
	sort.Strings(names)
	if len(names) != 2 || names[0] != "a.png" || names[1] != "b.png" {
		t.Fatalf("listed names = %v, want [a.png b.png]", names)
	}
	if sizes["a.png"] != 5 || sizes["b.png"] != 3 {
		t.Fatalf("sizes = %v, want a.png=5 b.png=3", sizes)
	}
}

func TestListOfAnEmptyDirectoryIsEmpty(t *testing.T) {
	files, err := NewLocalFileStorage(t.TempDir()).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("List = %v, want no files", files)
	}
}

func TestListFailsWhenTheDirectoryDoesNotExist(t *testing.T) {
	store := NewLocalFileStorage(filepath.Join(t.TempDir(), "missing"))

	if _, err := store.List(context.Background()); err == nil {
		t.Fatal("List of a missing directory succeeded, want an error")
	}
}

func TestDeleteRemovesTheFile(t *testing.T) {
	directory := t.TempDir()
	store := NewLocalFileStorage(directory)
	if err := store.Save(context.Background(), "a.png", []byte("x")); err != nil {
		t.Fatal(err)
	}

	if err := store.Delete(context.Background(), "a.png"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := os.Stat(filepath.Join(directory, "a.png")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file still exists after Delete (stat error: %v)", err)
	}
}

func TestDeleteReportsMediaNotFoundForAMissingFile(t *testing.T) {
	store := NewLocalFileStorage(t.TempDir())

	err := store.Delete(context.Background(), "missing.png")

	if !errors.Is(err, domain.ErrMediaNotFound) {
		t.Fatalf("Delete of a missing file error = %v, want ErrMediaNotFound", err)
	}
}

func TestDeleteDoesNotLeaveTheUploadDirectory(t *testing.T) {
	parent := t.TempDir()
	directory := filepath.Join(parent, "uploads")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(parent, "keep.png")
	if err := os.WriteFile(outside, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := NewLocalFileStorage(directory)

	err := store.Delete(context.Background(), "../keep.png")

	if !errors.Is(err, domain.ErrMediaNotFound) {
		t.Fatalf("Delete with a traversal name error = %v, want ErrMediaNotFound", err)
	}
	if _, statErr := os.Stat(outside); statErr != nil {
		t.Fatalf("file outside the upload directory was removed: %v", statErr)
	}
}

func TestOperationsStopWhenTheContextIsCanceled(t *testing.T) {
	directory := t.TempDir()
	store := NewLocalFileStorage(directory)
	if err := store.Save(context.Background(), "a.png", []byte("x")); err != nil {
		t.Fatal(err)
	}
	ctx := canceledContext()

	if err := store.Save(ctx, "b.png", []byte("x")); !errors.Is(err, context.Canceled) {
		t.Errorf("Save error = %v, want context.Canceled", err)
	}
	if _, err := store.List(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("List error = %v, want context.Canceled", err)
	}
	if err := store.Delete(ctx, "a.png"); !errors.Is(err, context.Canceled) {
		t.Errorf("Delete error = %v, want context.Canceled", err)
	}

	if _, err := os.Stat(filepath.Join(directory, "b.png")); !errors.Is(err, os.ErrNotExist) {
		t.Error("a canceled Save still created the file")
	}
	if _, err := os.Stat(filepath.Join(directory, "a.png")); err != nil {
		t.Errorf("a canceled Delete still removed the file: %v", err)
	}
}
