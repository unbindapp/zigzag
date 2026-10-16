package utils

import (
	"net/url"
	"path"
)

func JoinURLPaths(baseURL string, paths ...string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(append([]string{u.Path}, paths...)...)
	return u.String(), nil
}
