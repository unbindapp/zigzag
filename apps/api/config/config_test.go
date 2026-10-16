package config

import (
	"strings"
	"testing"

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
