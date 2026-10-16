package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCacheExpires(t *testing.T) {
	c := NewMemoryCache[string]()
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	c.now = func() time.Time { return now }
	ctx := context.Background()

	_, err := c.Get(ctx, "a")
	assert.ErrorIs(t, err, ErrMiss)

	assert.NoError(t, c.SetWithExpiration(ctx, "a", "one", time.Minute))
	v, err := c.Get(ctx, "a")
	assert.NoError(t, err)
	assert.Equal(t, "one", v)

	now = now.Add(time.Minute)
	_, err = c.Get(ctx, "a")
	assert.ErrorIs(t, err, ErrMiss)
}
