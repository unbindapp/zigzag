package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

const uploadsPrefix = "uploads"

var (
	ErrNotFound   = errors.New("stored file not found")
	ErrInvalidKey = errors.New("invalid storage key")
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i StoreInterface -p storage -s LocalStore -o storage_iface.go
type LocalStore struct {
	root string
}

func NewLocalStore(dataDir string) (*LocalStore, error) {
	root, err := filepath.Abs(filepath.Join(dataDir, uploadsPrefix))
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, fmt.Errorf("create upload directory: %w", err)
	}
	return &LocalStore{root: root}, nil
}

func HasLocalUploads(dataDir string) bool {
	info, err := os.Stat(filepath.Join(dataDir, uploadsPrefix))
	return err == nil && info.IsDir()
}

func (self *LocalStore) Put(ctx context.Context, key string, reader io.Reader, contentType string) error {
	target, err := self.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".upload-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, contextReader{ctx: ctx, reader: reader}); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), target)
}

func (self *LocalStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	target, err := self.path(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(target)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	return f, err
}

func (self *LocalStore) Exists(ctx context.Context, key string) (bool, error) {
	target, err := self.path(key)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(target)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

func (self *LocalStore) Delete(ctx context.Context, key string) error {
	target, err := self.path(key)
	if err != nil {
		return err
	}
	err = os.Remove(target)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (self *LocalStore) path(key string) (string, error) {
	if err := validateKey(key); err != nil {
		return "", err
	}
	return filepath.Join(self.root, filepath.FromSlash(key)), nil
}

func validateKey(key string) error {
	if key == "" || path.Clean(key) != key || !filepath.IsLocal(filepath.FromSlash(key)) {
		return ErrInvalidKey
	}
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (self contextReader) Read(p []byte) (int, error) {
	if err := self.ctx.Err(); err != nil {
		return 0, err
	}
	return self.reader.Read(p)
}
