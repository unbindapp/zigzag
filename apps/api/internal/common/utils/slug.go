package utils

import (
	"regexp"
	"strings"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(displayName string) string {
	slug := strings.ToLower(displayName)
	slug = nonAlphanumeric.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

func TruncateSlug(slug string, maxLength int) string {
	if len(slug) <= maxLength {
		return slug
	}
	return strings.TrimRight(slug[:maxLength], "-")
}

func SlugWithSuffix(displayName string, suffixBytes int) (string, error) {
	slug := Slugify(displayName)
	if slug == "" {
		slug = "untitled"
	}
	suffix, err := RandomHex(suffixBytes)
	if err != nil {
		return "", err
	}
	return slug + "-" + suffix, nil
}
