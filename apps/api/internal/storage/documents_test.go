package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLocalPrivateDocumentsStreamsAndPromotesObjects(t *testing.T) {
	root := filepath.Join(t.TempDir(), "private-docs")
	store, err := NewLocalPrivateDocuments(root)
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("private bytes")
	key := ".staging/random.part"
	obj, err := store.Save(context.Background(), key, bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatal(err)
	}
	if obj.SizeBytes != int64(len(content)) || obj.SHA256 == "" {
		t.Fatalf("unexpected object hash/size: %+v", obj)
	}
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(key)))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("private object permissions are too broad: %o", info.Mode().Perm())
	}
	if err := store.Promote(context.Background(), key, "documents/2026/10/object.bin"); err != nil {
		t.Fatal(err)
	}
	reader, err := store.Open(context.Background(), "documents/2026/10/object.bin")
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(reader)
	reader.Close()
	if err != nil || !bytes.Equal(got, content) {
		t.Fatalf("promoted bytes differ: %q err=%v", got, err)
	}
}

func TestLocalPrivateDocumentsPromotesRegistrationAttachments(t *testing.T) {
	store, err := NewLocalPrivateDocuments(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(context.Background(), ".staging/registration.part", bytes.NewReader([]byte("private image")), 64); err != nil {
		t.Fatal(err)
	}
	if err := store.Promote(context.Background(), ".staging/registration.part", "registration/opaque.png"); err != nil {
		t.Fatal(err)
	}
	if exists, err := store.Exists(context.Background(), "registration/opaque.png"); err != nil || !exists {
		t.Fatalf("registration attachment exists=%v err=%v", exists, err)
	}
	if err := store.Promote(context.Background(), ".staging/other.part", "arbitrary/opaque.png"); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("unexpected destination prefix accepted: %v", err)
	}
}

func TestLocalPrivateDocumentsRejectsTraversalAndPublicRoots(t *testing.T) {
	if _, err := NewLocalPrivateDocuments(filepath.Join(t.TempDir(), "public", "docs")); err == nil {
		t.Fatal("public web root was accepted")
	}
	store, err := NewLocalPrivateDocuments(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"../escape", "documents\\escape", "/absolute", "documents/%2e%2e/escape", "documents//empty"} {
		if _, err := store.Save(context.Background(), key, bytes.NewReader([]byte("x")), 5); !errors.Is(err, ErrInvalidKey) {
			t.Errorf("key %q error = %v", key, err)
		}
	}
}

func TestLocalPrivateDocumentsEnforcesLimitAndCleansPartialFile(t *testing.T) {
	root := t.TempDir()
	store, err := NewLocalPrivateDocuments(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(context.Background(), ".staging/too-large.part", bytes.NewReader([]byte("12345")), 4); !errors.Is(err, ErrSizeLimit) {
		t.Fatalf("oversize error=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".staging", "too-large.part")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial upload was not cleaned: %v", err)
	}
}
