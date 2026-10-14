package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalPublicMediaStorageConfinementAndLifecycle(t *testing.T) {
	root := t.TempDir()
	if _, err := NewLocalPublicMediaStorage(filepath.Join(root, "web", "public", "uploads")); err == nil {
		t.Fatal("media storage accepted a path inside a web static directory")
	}
	store, err := NewLocalPublicMediaStorage(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(context.Background(), "../escape.png", bytes.NewReader([]byte("image")), 1024); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("path traversal key error = %v", err)
	}
	value := []byte("valid-bytes-for-storage-layer")
	object, err := store.Save(context.Background(), "abc123.png", bytes.NewReader(value), 1024)
	if err != nil {
		t.Fatal(err)
	}
	if object.SizeBytes != int64(len(value)) || len(object.SHA256) != 64 {
		t.Fatalf("unexpected media object metadata: %+v", object)
	}
	reader, err := store.Open(context.Background(), "abc123.png")
	if err != nil {
		t.Fatal(err)
	}
	got, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil || !bytes.Equal(got, value) {
		t.Fatalf("read saved media: read=%v close=%v data=%q", readErr, closeErr, got)
	}
	if err := store.Delete(context.Background(), "abc123.png"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "abc123.png")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("media object remains after deletion: %v", err)
	}
}
