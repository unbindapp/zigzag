package oauthserver

import (
	"errors"
	"net/url"
	"slices"
	"strings"
)

const maxRedirectURILen = 512

var nativeAppSchemes = []string{"cursor"}

func ValidateRedirectURI(raw string) error {
	if raw == "" {
		return errors.New("redirect uri is required")
	}
	if len(raw) > maxRedirectURILen {
		return errors.New("redirect uri is too long")
	}
	if strings.Contains(raw, "#") {
		return errors.New("redirect uri must not contain a fragment")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.User != nil {
		return errors.New("redirect uri must be an absolute url without credentials")
	}
	switch parsed.Scheme {
	case "https":
		return nil
	case "http":
		if IsLoopbackHost(parsed.Hostname()) {
			return nil
		}
		return errors.New("http redirect uris are only allowed on localhost")
	default:
		if IsNativeAppScheme(parsed.Scheme) {
			return nil
		}
		return errors.New("redirect uri must use https")
	}
}

func IsNativeAppScheme(scheme string) bool {
	return slices.Contains(nativeAppSchemes, strings.ToLower(scheme))
}

func isOnDevice(redirect *url.URL) bool {
	return IsLoopbackHost(redirect.Hostname()) || IsNativeAppScheme(redirect.Scheme)
}

func IsLoopbackHost(hostname string) bool {
	switch strings.ToLower(hostname) {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	return false
}

func RedirectURIMatches(registered, requested string) bool {
	if registered == requested {
		return true
	}
	reg, err := url.Parse(registered)
	if err != nil {
		return false
	}
	req, err := url.Parse(requested)
	if err != nil {
		return false
	}
	if reg.Scheme != "http" || req.Scheme != "http" {
		return false
	}
	if !IsLoopbackHost(reg.Hostname()) || !IsLoopbackHost(req.Hostname()) {
		return false
	}
	reg.Host = reg.Hostname()
	req.Host = req.Hostname()
	return reg.String() == req.String()
}

func LoopbackOnly(redirectURIs []string) bool {
	if len(redirectURIs) == 0 {
		return false
	}
	for _, raw := range redirectURIs {
		parsed, err := url.Parse(raw)
		if err != nil || !isOnDevice(parsed) {
			return false
		}
	}
	return true
}

func HostOf(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if IsNativeAppScheme(parsed.Scheme) {
		return parsed.Scheme + "://" + parsed.Host
	}
	return parsed.Host
}

func BuildRedirect(redirectURI string, params url.Values) string {
	parsed, err := url.Parse(redirectURI)
	if err != nil {
		return redirectURI
	}
	query := parsed.Query()
	for key, values := range params {
		for _, value := range values {
			query.Set(key, value)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
