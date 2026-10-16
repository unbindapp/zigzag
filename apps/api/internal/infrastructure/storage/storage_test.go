package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutOpenDelete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalStore(dir)
	require.NoError(t, err)
	ctx := context.Background()

	require.NoError(t, store.Put(ctx, "ws-1/file.png", strings.NewReader("hello"), ""))
	_, err = os.Stat(filepath.Join(dir, "uploads", "ws-1", "file.png"))
	require.NoError(t, err, "files live under DATA_DIR/uploads")

	f, err := store.Open(ctx, "ws-1/file.png")
	require.NoError(t, err)
	body, _ := io.ReadAll(f)
	f.Close()
	assert.Equal(t, "hello", string(body))

	require.NoError(t, store.Put(ctx, "ws-1/file.png", strings.NewReader("replaced"), ""))
	f, _ = store.Open(ctx, "ws-1/file.png")
	body, _ = io.ReadAll(f)
	f.Close()
	assert.Equal(t, "replaced", string(body))

	exists, err := store.Exists(ctx, "ws-1/file.png")
	require.NoError(t, err)
	assert.True(t, exists)

	require.NoError(t, store.Delete(ctx, "ws-1/file.png"))
	exists, err = store.Exists(ctx, "ws-1/file.png")
	require.NoError(t, err)
	assert.False(t, exists)
	_, err = store.Open(ctx, "ws-1/file.png")
	assert.ErrorIs(t, err, ErrNotFound)
	assert.NoError(t, store.Delete(ctx, "ws-1/file.png"), "deleting a missing file is not an error")
}

func TestKeysCannotEscapeTheRoot(t *testing.T) {
	store, err := NewLocalStore(t.TempDir())
	require.NoError(t, err)
	ctx := context.Background()

	for _, key := range []string{"", "../escape", "a/../../escape", "/etc/passwd", "a/../b", "a//b"} {
		assert.ErrorIs(t, store.Put(ctx, key, strings.NewReader("x"), ""), ErrInvalidKey, key)
		_, err := store.Open(ctx, key)
		assert.ErrorIs(t, err, ErrInvalidKey, key)
	}
}

func TestHasLocalUploads(t *testing.T) {
	dir := t.TempDir()
	assert.False(t, HasLocalUploads(dir))
	_, err := NewLocalStore(dir)
	require.NoError(t, err)
	assert.True(t, HasLocalUploads(dir))
}

func TestCancelledPutLeavesNothing(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewLocalStore(dir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.Error(t, store.Put(ctx, "ws/file", strings.NewReader("data"), ""))
	_, err := store.Open(context.Background(), "ws/file")
	assert.ErrorIs(t, err, ErrNotFound)
	entries, _ := os.ReadDir(filepath.Join(dir, "uploads", "ws"))
	assert.Empty(t, entries)
}
