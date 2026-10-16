package names

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
)

var (
	slugPattern        = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	displayNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,31}$`)
	teamKeyPattern     = regexp.MustCompile(`^[A-Z][A-Z0-9]{0,6}$`)
)

func Slug(raw string) (string, error) {
	cleaned, err := Clean(raw)
	if err != nil {
		return "", errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Slug cannot be empty")
	}
	slug := utils.Slugify(cleaned)
	if len(slug) < 2 || len(slug) > MaxLength || !slugPattern.MatchString(slug) {
		return "", errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Slug must be 2 to 32 lowercase letters, digits or dashes")
	}
	return slug, nil
}

func DisplayName(raw string) (string, error) {
	handle := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "@")))
	if !displayNamePattern.MatchString(handle) {
		return "", errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Display name must be 1 to 32 lowercase letters, digits, dots, dashes or underscores, starting with a letter or digit")
	}
	return handle, nil
}

func TeamKey(raw string) (string, error) {
	key := strings.ToUpper(strings.TrimSpace(raw))
	if !teamKeyPattern.MatchString(key) {
		return "", errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Team key must be 1 to 7 letters or digits, starting with a letter")
	}
	return key, nil
}

const (
	maxTeamKeyLength      = 7
	fallbackTeamKeyLength = 3
	fallbackTeamKey       = "TEAM"
)

func TeamKeyFromName(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	initials, letters := "", ""
	for _, word := range words {
		ascii := strings.Map(keepTeamKeyRune, strings.ToUpper(word))
		if ascii == "" {
			continue
		}
		initials += ascii[:1]
		letters += ascii
	}
	if key := teamKeyPrefix(initials, maxTeamKeyLength); len(key) > 1 {
		return key
	}
	if key := teamKeyPrefix(letters, fallbackTeamKeyLength); key != "" {
		return key
	}
	return fallbackTeamKey
}

func teamKeyPrefix(candidate string, length int) string {
	key := strings.TrimLeft(candidate, "0123456789")
	if len(key) > length {
		key = key[:length]
	}
	return key
}

func keepTeamKeyRune(r rune) rune {
	if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
		return r
	}
	return -1
}

func Email(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	at := strings.LastIndex(email, "@")
	if at < 1 || at == len(email)-1 || strings.ContainsAny(email, " \t\r\n") {
		return "", errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Email address is not valid")
	}
	return email, nil
}

func SuggestDisplayName(email string) string {
	local, _, _ := strings.Cut(strings.ToLower(email), "@")
	handle := strings.Trim(regexp.MustCompile(`[^a-z0-9._-]+`).ReplaceAllString(local, ""), "._-")
	if handle == "" {
		handle = "user"
	}
	if len(handle) > 24 {
		handle = handle[:24]
	}
	return handle
}

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func Color(raw string) (string, error) {
	color := strings.ToLower(strings.TrimSpace(raw))
	if !colorPattern.MatchString(color) {
		return "", errdefs.NewCustomError(errdefs.ErrTypeInvalidInput, "Color must be a hex color such as #5e6ad2")
	}
	return color, nil
}

func OptionalColor(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	return Color(raw)
}
