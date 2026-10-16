package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"Hello World":         "hello-world",
		"  Fix: the  BUG!! ":  "fix-the-bug",
		"Ünïcode":             "n-code",
		"---":                 "",
		"already-a-slug-2026": "already-a-slug-2026",
	}
	for in, want := range tests {
		assert.Equal(t, want, Slugify(in), in)
	}
}

func TestTruncateSlug(t *testing.T) {
	assert.Equal(t, "abc", TruncateSlug("abc", 10))
	assert.Equal(t, "abc", TruncateSlug("abc-def", 4))
	assert.Equal(t, "abc-d", TruncateSlug("abc-def", 5))
}

func TestSlugWithSuffix(t *testing.T) {
	slug, err := SlugWithSuffix("Mobile App", 4)
	require.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`^mobile-app-[0-9a-f]{8}$`), slug)

	slug, err = SlugWithSuffix("!!!", 4)
	require.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`^untitled-[0-9a-f]{8}$`), slug)
}

func TestRandomTokensDiffer(t *testing.T) {
	a, err := RandomToken(32)
	require.NoError(t, err)
	b, err := RandomToken(32)
	require.NoError(t, err)
	assert.NotEqual(t, a, b)
	assert.Len(t, a, 43)
	assert.Equal(t, HashToken(a), HashToken(a))
	assert.Len(t, HashToken(a), 64)
}
