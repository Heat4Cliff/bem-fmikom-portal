package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type MediaObject struct {
	SizeBytes int64
	SHA256    string
}

// PublicMediaStorage keeps public image bytes behind an API-controlled read path.
// Implementations can later be replaced by an object store without changing services.
type PublicMediaStorage interface {
	Save(ctx context.Context, key string, source io.Reader, maxBytes int64) (MediaObject, error)
	Open(ctx context.Context, key string) (DocumentReader, error)
	Delete(ctx context.Context, key string) error
}

type LocalPublicMediaStorage struct{ root string }

func NewLocalPublicMediaStorage(directory string) (*LocalPublicMediaStorage, error) {
	if directory == "" {
		return nil, errors.New("public media storage path is required")
	}
	root, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("resolve public media storage path")
	}
	for _, segment := range strings.Split(strings.ToLower(filepath.Clean(root)), string(filepath.Separator)) {
		if segment == "public" || segment == "static" || segment == "wwwroot" {
			return nil, errors.New("public media storage cannot be mounted inside a web static directory")
		}
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, fmt.Errorf("create public media storage: %w", err)
	}
	return &LocalPublicMediaStorage{root: root}, nil
}

func (s *LocalPublicMediaStorage) Save(ctx context.Context, key string, source io.Reader, maxBytes int64) (MediaObject, error) {
	clean, err := validateKey(key)
	if err != nil || maxBytes < 1 {
		return MediaObject{}, ErrInvalidKey
	}
	if err := ctx.Err(); err != nil {
		return MediaObject{}, err
	}
	root, err := os.OpenRoot(s.root)
	if err != nil {
		return MediaObject{}, fmt.Errorf("open public media root: %w", err)
	}
	defer root.Close()
	file, err := root.OpenFile(clean, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return MediaObject{}, fmt.Errorf("create public media object: %w", err)
	}
	removeOnFailure := true
	defer func() {
		_ = file.Close()
		if removeOnFailure {
			_ = root.Remove(clean)
		}
	}()
	hasher := sha256.New()
	size, err := io.Copy(io.MultiWriter(file, hasher), io.LimitReader(source, maxBytes+1))
	if err != nil || size < 1 || size > maxBytes {
		return MediaObject{}, ErrSizeLimit
	}
	if err := ctx.Err(); err != nil {
		return MediaObject{}, err
	}
	if err := file.Sync(); err != nil {
		return MediaObject{}, fmt.Errorf("sync public media object: %w", err)
	}
	if err := file.Close(); err != nil {
		return MediaObject{}, fmt.Errorf("close public media object: %w", err)
	}
	removeOnFailure = false
	return MediaObject{SizeBytes: size, SHA256: hex.EncodeToString(hasher.Sum(nil))}, nil
}

func (s *LocalPublicMediaStorage) Open(ctx context.Context, key string) (DocumentReader, error) {
	clean, err := validateKey(key)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(s.root)
	if err != nil {
		return nil, fmt.Errorf("open public media root: %w", err)
	}
	file, err := root.Open(clean)
	if err != nil {
		_ = root.Close()
		return nil, err
	}
	return &rootedReadCloser{File: file, root: root}, nil
}

func (s *LocalPublicMediaStorage) Delete(ctx context.Context, key string) error {
	clean, err := validateKey(key)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	root, err := os.OpenRoot(s.root)
	if err != nil {
		return fmt.Errorf("open public media root: %w", err)
	}
	defer root.Close()
	if err := root.Remove(clean); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove public media object: %w", err)
	}
	return nil
}
