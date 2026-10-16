package config

import (
	"strings"
	"testing"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKey = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

func lookupOf(vars map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		v, ok := vars[name]
		return v, ok
	}
}

func baseConfig() *Config {
	return &Config{
		ExternalUIUrl:    "http://localhost:5174",
		PostgresHost:     "localhost",
		PostgresPort:     54329,
		PostgresUser:     "postgres",
		PostgresPassword: "postgres",
		PostgresDB:       "zigzag",
		PostgresSSLMode:  "disable",
		EncryptionKey:    testKey,
	}
}

func TestEncryptionKeyMustBe32HexBytes(t *testing.T) {
	for name, key := range map[string]string{
		"empty":   "",
		"short":   "abcd",
		"not hex": strings.Repeat("z", 64),
	} {
		t.Run(name, func(t *testing.T) {
			cfg := baseConfig()
			cfg.EncryptionKey = key
			assert.Error(t, cfg.resolve(lookupOf(nil)))
		})
	}

	cfg := baseConfig()
	require.NoError(t, cfg.resolve(lookupOf(nil)))
	assert.Len(t, cfg.EncryptionKeyBytes(), 32)
}

func TestExternalAPIURLDefaults(t *testing.T) {
	t.Run("local dev", func(t *testing.T) {
		cfg := baseConfig()
		require.NoError(t, cfg.resolve(lookupOf(nil)))
		assert.Equal(t, "http://localhost:8091", cfg.ExternalAPIURL)
	})
	t.Run("only the ui url set", func(t *testing.T) {
		cfg := baseConfig()
		cfg.ExternalUIUrl = "https://zigzag.unbind.app/"
		require.NoError(t, cfg.resolve(lookupOf(map[string]string{"EXTERNAL_UI_URL": "https://zigzag.unbind.app/"})))
		assert.Equal(t, "https://zigzag.unbind.app", cfg.ExternalUIUrl)
		assert.Equal(t, "https://zigzag.unbind.app/api/go", cfg.ExternalAPIURL)
	})
	t.Run("explicit api url wins", func(t *testing.T) {
		cfg := baseConfig()
		cfg.ExternalAPIURL = "https://api.example.com/"
		require.NoError(t, cfg.resolve(lookupOf(map[string]string{"EXTERNAL_UI_URL": "https://example.com"})))
		assert.Equal(t, "https://api.example.com", cfg.ExternalAPIURL)
	})
}

func TestDatabaseURLOverridesPostgresFields(t *testing.T) {
	cfg := baseConfig()
	cfg.DatabaseURL = "postgres://zz:s3cr%2Ft@db.internal:6543/tracker?sslmode=require"
	require.NoError(t, cfg.resolve(lookupOf(nil)))
	assert.Equal(t, "db.internal", cfg.GetPostgresHost())
	assert.Equal(t, 6543, cfg.GetPostgresPort())
	assert.Equal(t, "zz", cfg.GetPostgresUser())
	assert.Equal(t, "s3cr/t", cfg.GetPostgresPassword())
	assert.Equal(t, "tracker", cfg.GetPostgresDB())
	assert.Equal(t, "require", cfg.GetPostgresSSLMode())
}

func TestDatabaseURLDefaultsPortAndKeepsSSLMode(t *testing.T) {
	cfg := baseConfig()
	cfg.DatabaseURL = "postgresql://u:p@pg/app"
	require.NoError(t, cfg.resolve(lookupOf(nil)))
	assert.Equal(t, 5432, cfg.GetPostgresPort())
	assert.Equal(t, "disable", cfg.GetPostgresSSLMode())
	assert.Equal(t, "app", cfg.GetPostgresDB())
}

func TestListenAddrFollowsPort(t *testing.T) {
	cfg := baseConfig()
	cfg.Port = 9000
	assert.Equal(t, ":9000", cfg.ListenAddr())
}

func TestDatabaseURLRejectsOtherSchemes(t *testing.T) {
	for _, raw := range []string{"mysql://u:p@h/db", "postgres:///nohost", "postgres://h:port/db"} {
		cfg := baseConfig()
		cfg.DatabaseURL = raw
		assert.Error(t, cfg.resolve(lookupOf(nil)), raw)
	}
}

func parseEnvironment(t *testing.T, vars map[string]string) (*Config, error) {
	t.Helper()
	vars["ENCRYPTION_KEY"] = testKey
	cfg := Config{}
	require.NoError(t, env.ParseWithOptions(&cfg, env.Options{Environment: vars}))
	return &cfg, cfg.resolve(lookupOf(vars))
}

func TestS3IsOffWithoutABucket(t *testing.T) {
	cfg, err := parseEnvironment(t, map[string]string{"S3_ENDPOINT": "http://silo:9000"})
	require.NoError(t, err)
	assert.False(t, cfg.S3Enabled())
	assert.Equal(t, "./.data", cfg.DataDir)
}

func TestS3Defaults(t *testing.T) {
	cfg, err := parseEnvironment(t, map[string]string{
		"S3_ENDPOINT":          "http://silo:9000",
		"S3_BUCKET":            "zigzag",
		"S3_ACCESS_KEY_ID":     "key",
		"S3_SECRET_ACCESS_KEY": "secret",
	})
	require.NoError(t, err)
	assert.True(t, cfg.S3Enabled())
	assert.Equal(t, "us-east-1", cfg.S3Region)
	assert.True(t, cfg.S3UsePathStyle)
	assert.Equal(t, "http://silo:9000", cfg.S3Endpoint)
	assert.Empty(t, cfg.S3KeyPrefix)
}

func TestS3ForCloudflareR2(t *testing.T) {
	cfg, err := parseEnvironment(t, map[string]string{
		"S3_ENDPOINT":          "https://0123456789abcdef.r2.cloudflarestorage.com",
		"S3_BUCKET":            "zigzag",
		"S3_ACCESS_KEY_ID":     "key",
		"S3_SECRET_ACCESS_KEY": "secret",
		"S3_REGION":            "auto",
		"S3_KEY_PREFIX":        "dux/",
	})
	require.NoError(t, err)
	assert.Equal(t, "auto", cfg.S3Region)
	assert.True(t, cfg.S3UsePathStyle)
	assert.Equal(t, "dux/", cfg.S3KeyPrefix)
}

func TestS3Overrides(t *testing.T) {
	cfg, err := parseEnvironment(t, map[string]string{
		"S3_BUCKET":            "zigzag",
		"S3_ACCESS_KEY_ID":     "key",
		"S3_SECRET_ACCESS_KEY": "secret",
		"S3_REGION":            "auto",
		"S3_USE_PATH_STYLE":    "false",
	})
	require.NoError(t, err)
	assert.Equal(t, "auto", cfg.S3Region)
	assert.False(t, cfg.S3UsePathStyle)
	assert.Empty(t, cfg.S3Endpoint, "no endpoint means AWS S3")
}

func TestS3Validation(t *testing.T) {
	for name, vars := range map[string]map[string]string{
		"missing access key": {"S3_BUCKET": "zigzag", "S3_SECRET_ACCESS_KEY": "secret"},
		"missing secret":     {"S3_BUCKET": "zigzag", "S3_ACCESS_KEY_ID": "key"},
		"endpoint not a url": {"S3_BUCKET": "zigzag", "S3_ACCESS_KEY_ID": "key", "S3_SECRET_ACCESS_KEY": "secret", "S3_ENDPOINT": "silo:9000"},
		"endpoint scheme":    {"S3_BUCKET": "zigzag", "S3_ACCESS_KEY_ID": "key", "S3_SECRET_ACCESS_KEY": "secret", "S3_ENDPOINT": "ftp://silo"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := parseEnvironment(t, vars)
			assert.Error(t, err)
		})
	}
}
