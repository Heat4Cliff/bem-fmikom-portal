package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var ErrInvalidKey = errors.New("invalid private document storage key")

type DocumentObject struct {
	SizeBytes int64
	SHA256    string
}

type DocumentReader interface {
	io.ReadCloser
	io.ReaderAt
	Stat() (os.FileInfo, error)
}

type DocumentStorage interface {
	Save(ctx context.Context, key string, source io.Reader, maxBytes int64) (DocumentObject, error)
	Open(ctx context.Context, key string) (DocumentReader, error)
	Promote(ctx context.Context, temporaryKey, storageKey string) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Stat(ctx context.Context, key string) (os.FileInfo, error)
}

type LocalPrivateDocuments struct {
	directory string
}

func NewLocalPrivateDocuments(directory string) (*LocalPrivateDocuments, error) {
	if strings.TrimSpace(directory) == "" {
		return nil, errors.New("document storage path is required")
	}
	abs, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("resolve private document directory")
	}
	for _, segment := range strings.Split(strings.ToLower(filepath.Clean(abs)), string(filepath.Separator)) {
		if segment == "public" || segment == "static" || segment == "wwwroot" {
			return nil, errors.New("private document storage cannot be inside a public/static web directory")
		}
	}
	if err := os.MkdirAll(abs, 0o700); err != nil {
		return nil, fmt.Errorf("create private document directory: %w", err)
	}
	return &LocalPrivateDocuments{directory: abs}, nil
}

func (store *LocalPrivateDocuments) openRoot() (*os.Root, error) {
	root, err := os.OpenRoot(store.directory)
	if err != nil {
		return nil, fmt.Errorf("open private document directory: %w", err)
	}
	return root, nil
}

func (store *LocalPrivateDocuments) Save(ctx context.Context, key string, source io.Reader, maxBytes int64) (DocumentObject, error) {
	clean, err := validateKey(key)
	if err != nil || maxBytes < 1 {
		return DocumentObject{}, ErrInvalidKey
	}
	if err := ctx.Err(); err != nil {
		return DocumentObject{}, err
	}
	root, err := store.openRoot()
	if err != nil {
		return DocumentObject{}, err
	}
	defer root.Close()
	if err := root.MkdirAll(path.Dir(clean), 0o700); err != nil {
		return DocumentObject{}, fmt.Errorf("create private object directory: %w", err)
	}
	file, err := root.OpenFile(clean, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return DocumentObject{}, fmt.Errorf("create private temporary object: %w", err)
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
	if err != nil {
		return DocumentObject{}, fmt.Errorf("stream private document: %w", err)
	}
	if size < 1 || size > maxBytes {
		return DocumentObject{}, ErrSizeLimit
	}
	if err := ctx.Err(); err != nil {
		return DocumentObject{}, err
	}
	if err := file.Sync(); err != nil {
		return DocumentObject{}, fmt.Errorf("sync private document: %w", err)
	}
	if err := file.Close(); err != nil {
		return DocumentObject{}, fmt.Errorf("close private document: %w", err)
	}
	removeOnFailure = false
	return DocumentObject{SizeBytes: size, SHA256: hex.EncodeToString(hasher.Sum(nil))}, nil
}

var ErrSizeLimit = errors.New("document exceeds configured size limit")

func (store *LocalPrivateDocuments) Open(ctx context.Context, key string) (DocumentReader, error) {
	clean, err := validateKey(key)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root, err := store.openRoot()
	if err != nil {
		return nil, err
	}
	file, err := root.Open(clean)
	if err != nil {
		_ = root.Close()
		return nil, err
	}
	return &rootedReadCloser{File: file, root: root}, nil
}

type rootedReadCloser struct {
	*os.File
	root *os.Root
}

func (reader *rootedReadCloser) Close() error {
	fileErr := reader.File.Close()
	rootErr := reader.root.Close()
	return errors.Join(fileErr, rootErr)
}

func (store *LocalPrivateDocuments) Promote(ctx context.Context, temporaryKey, storageKey string) error {
	source, sourceErr := validateKey(temporaryKey)
	destination, destinationErr := validateKey(storageKey)
	if sourceErr != nil || destinationErr != nil || !strings.HasPrefix(source, ".staging/") || !strings.HasPrefix(destination, "documents/") {
		return ErrInvalidKey
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	root, err := store.openRoot()
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.MkdirAll(path.Dir(destination), 0o700); err != nil {
		return fmt.Errorf("create final document directory: %w", err)
	}
	if _, err := root.Lstat(destination); err == nil {
		return errors.New("private document object already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect final document object: %w", err)
	}
	if err := root.Rename(source, destination); err != nil {
		return fmt.Errorf("promote private document object: %w", err)
	}
	return nil
}

func (store *LocalPrivateDocuments) Delete(ctx context.Context, key string) error {
	clean, err := validateKey(key)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	root, err := store.openRoot()
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.Remove(clean); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove private document object: %w", err)
	}
	return nil
}

func (store *LocalPrivateDocuments) Exists(ctx context.Context, key string) (bool, error) {
	info, err := store.Stat(ctx, key)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.Mode().IsRegular(), nil
}

func (store *LocalPrivateDocuments) Stat(ctx context.Context, key string) (os.FileInfo, error) {
	clean, err := validateKey(key)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root, err := store.openRoot()
	if err != nil {
		return nil, err
	}
	defer root.Close()
	info, err := root.Stat(clean)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("private document object is not a regular file")
	}
	return info, nil
}

func validateKey(key string) (string, error) {
	if key == "" || strings.ContainsAny(key, "\\:%\x00") || strings.HasPrefix(key, "/") || filepath.IsAbs(key) {
		return "", ErrInvalidKey
	}
	for _, segment := range strings.Split(key, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", ErrInvalidKey
		}
	}
	if path.Clean(key) != key {
		return "", ErrInvalidKey
	}
	return key, nil
}
