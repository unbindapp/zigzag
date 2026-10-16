package names

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
)

func TestSlug(t *testing.T) {
	valid := map[string]string{
		"acme":          "acme",
		"  Acme Corp  ": "acme-corp",
		"Zig_Zag 2026":  "zig-zag-2026",
		"ab":            "ab",
	}
	for in, want := range valid {
		got, err := Slug(in)
		assert.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	for _, in := range []string{"", "   ", "a", "!!", "this-slug-is-definitely-longer-than-32"} {
		_, err := Slug(in)
		assert.ErrorIs(t, err, errdefs.ErrInvalidInput, in)
	}
}

func TestDisplayName(t *testing.T) {
	for in, want := range map[string]string{"Ada": "ada", "@grace.hopper": "grace.hopper", " linus_t ": "linus_t", "a": "a"} {
		got, err := DisplayName(in)
		assert.NoError(t, err, in)
		assert.Equal(t, want, got)
	}
	for _, in := range []string{"", "-ada", "ada lovelace", "ada@home", "abcdefghijklmnopqrstuvwxyz0123456"} {
		_, err := DisplayName(in)
		assert.ErrorIs(t, err, errdefs.ErrInvalidInput, in)
	}
}

func TestTeamKey(t *testing.T) {
	for in, want := range map[string]string{"eng": "ENG", " Web2 ": "WEB2", "A": "A", "ABCDEFG": "ABCDEFG"} {
		got, err := TeamKey(in)
		assert.NoError(t, err, in)
		assert.Equal(t, want, got)
	}
	for _, in := range []string{"", "1ENG", "EN-G", "ABCDEFGH"} {
		_, err := TeamKey(in)
		assert.ErrorIs(t, err, errdefs.ErrInvalidInput, in)
	}
}

func TestTeamKeyFromName(t *testing.T) {
	for in, want := range map[string]string{
		"Acme Corp":                    "AC",
		"acme":                         "ACM",
		"Zig-zag labs":                 "ZZL",
		"a b c d e f g h i":            "ABCDEFG",
		"2026 Product Team":            "PT",
		"Ab":                           "AB",
		"Q":                            "Q",
		"42":                           "TEAM",
		"Café Ünited":                  "CN",
		"Ölbaum":                       "LBA",
		"🚀":                            "TEAM",
		"Unbind 2":                     "U2",
		"   the   quick brown  fox   ": "TQBF",
	} {
		got := TeamKeyFromName(in)
		assert.Equal(t, want, got, in)
		_, err := TeamKey(got)
		assert.NoError(t, err, in)
	}
}

func TestEmail(t *testing.T) {
	got, err := Email("  Ada@Example.COM ")
	assert.NoError(t, err)
	assert.Equal(t, "ada@example.com", got)
	for _, in := range []string{"", "ada", "@example.com", "ada@", "a da@example.com"} {
		_, err := Email(in)
		assert.ErrorIs(t, err, errdefs.ErrInvalidInput, in)
	}
}

func TestSuggestDisplayName(t *testing.T) {
	assert.Equal(t, "ada.lovelace", SuggestDisplayName("Ada.Lovelace@example.com"))
	assert.Equal(t, "user", SuggestDisplayName("+++@example.com"))
}

func TestColor(t *testing.T) {
	got, err := Color(" #5E6AD2 ")
	assert.NoError(t, err)
	assert.Equal(t, "#5e6ad2", got)
	for _, in := range []string{"", "5e6ad2", "#fff", "#gggggg"} {
		_, err := Color(in)
		assert.ErrorIs(t, err, errdefs.ErrInvalidInput, in)
	}
}
