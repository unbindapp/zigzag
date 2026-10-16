package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinURLPaths(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		paths    []string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple join",
			baseURL:  "https://google.com",
			paths:    []string{"search", "fiveguys"},
			expected: "https://google.com/search/fiveguys",
			wantErr:  false,
		},
		{
			name:     "base URL with trailing slash",
			baseURL:  "https://google.com/",
			paths:    []string{"search", "fiveguys"},
			expected: "https://google.com/search/fiveguys",
			wantErr:  false,
		},
		{
			name:     "paths with leading slashes",
			baseURL:  "https://google.com",
			paths:    []string{"/search", "/fiveguys"},
			expected: "https://google.com/search/fiveguys",
			wantErr:  false,
		},
		{
			name:     "paths with trailing slashes",
			baseURL:  "https://google.com",
			paths:    []string{"search/", "fiveguys/"},
			expected: "https://google.com/search/fiveguys",
			wantErr:  false,
		},
		{
			name:     "empty paths",
			baseURL:  "https://google.com",
			paths:    []string{},
			expected: "https://google.com",
			wantErr:  false,
		},
		{
			name:     "with query parameters",
			baseURL:  "https://google.com?param=value",
			paths:    []string{"search", "fiveguys"},
			expected: "https://google.com/search/fiveguys?param=value",
			wantErr:  false,
		},
		{
			name:     "with fragment",
			baseURL:  "https://google.com#fragment",
			paths:    []string{"search", "fiveguys"},
			expected: "https://google.com/search/fiveguys#fragment",
			wantErr:  false,
		},
		{
			name:     "with existing path",
			baseURL:  "https://google.com/api",
			paths:    []string{"search", "fiveguys"},
			expected: "https://google.com/api/search/fiveguys",
			wantErr:  false,
		},
		{
			name:     "with double slashes",
			baseURL:  "https://google.com",
			paths:    []string{"search//", "//fiveguys"},
			expected: "https://google.com/search/fiveguys",
			wantErr:  false,
		},
		{
			name:     "invalid URL",
			baseURL:  ":invalid-url",
			paths:    []string{"search", "fiveguys"},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := JoinURLPaths(tt.baseURL, tt.paths...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
