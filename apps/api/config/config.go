package config

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/unbindapp/zigzag-api/internal/common/log"
)

const (
	defaultExternalAPIURL = "http://localhost:8091"
	apiPathPrefix         = "/api/go"
)

type ConfigInterface interface {
	GetPostgresHost() string
	GetPostgresPort() int
	GetPostgresUser() string
	GetPostgresPassword() string
	GetPostgresDB() string
	GetPostgresSSLMode() string
}

type Config struct {
	Port             int    `env:"PORT" envDefault:"8091"`
	ExternalUIUrl    string `env:"EXTERNAL_UI_URL" envDefault:"http://localhost:5174"`
	ExternalAPIURL   string `env:"EXTERNAL_API_URL"`
	DatabaseURL      string `env:"DATABASE_URL"`
	PostgresHost     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	PostgresPort     int    `env:"POSTGRES_PORT" envDefault:"54329"`
	PostgresUser     string `env:"POSTGRES_USER" envDefault:"postgres"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" envDefault:"postgres"`
	PostgresDB       string `env:"POSTGRES_DB" envDefault:"zigzag"`
	PostgresSSLMode  string `env:"POSTGRES_SSL_MODE" envDefault:"disable"`
	CookieSecure     bool   `env:"COOKIE_SECURE" envDefault:"true"`
	TokenAudience    string `env:"TOKEN_AUDIENCE" envDefault:"zigzag-api"`
	EncryptionKey    string `env:"ENCRYPTION_KEY,required"`
	DataDir          string `env:"DATA_DIR" envDefault:"./.data"`
	MaxUploadBytes   int64  `env:"MAX_UPLOAD_BYTES" envDefault:"26214400"`
	InjectDevOrigins bool   `env:"INJECT_DEV_ORIGINS" envDefault:"false"`
	LogLevel         string `env:"LOG_LEVEL" envDefault:"info"`

	encryptionKey []byte
}

func (self *Config) GetPostgresHost() string     { return self.PostgresHost }
func (self *Config) GetPostgresPort() int        { return self.PostgresPort }
func (self *Config) GetPostgresUser() string     { return self.PostgresUser }
func (self *Config) GetPostgresPassword() string { return self.PostgresPassword }
func (self *Config) GetPostgresDB() string       { return self.PostgresDB }
func (self *Config) GetPostgresSSLMode() string  { return self.PostgresSSLMode }
func (self *Config) EncryptionKeyBytes() []byte  { return self.encryptionKey }
func (self *Config) APIPathPrefix() string       { return apiPathPrefix }
func (self *Config) ListenAddr() string          { return fmt.Sprintf(":%d", self.Port) }

func Load() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	if err := cfg.resolve(os.LookupEnv); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func NewConfig() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatal("Error parsing environment", "err", err)
	}
	return cfg
}

func (self *Config) resolve(lookup func(string) (string, bool)) error {
	key, err := parseEncryptionKey(self.EncryptionKey)
	if err != nil {
		return err
	}
	self.encryptionKey = key

	self.ExternalUIUrl = strings.TrimRight(self.ExternalUIUrl, "/")
	self.ExternalAPIURL = strings.TrimRight(self.ExternalAPIURL, "/")
	if self.ExternalAPIURL == "" {
		self.ExternalAPIURL = defaultExternalAPIURL
		if _, set := lookup("EXTERNAL_UI_URL"); set {
			self.ExternalAPIURL = self.ExternalUIUrl + apiPathPrefix
		}
	}

	if self.DatabaseURL == "" {
		return nil
	}
	return self.applyDatabaseURL(self.DatabaseURL)
}

func parseEncryptionKey(raw string) ([]byte, error) {
	if len(raw) != 64 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be 64 hex characters (32 bytes), generate one with: openssl rand -hex 32")
	}
	key, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be hex: %w", err)
	}
	return key, nil
}

func (self *Config) applyDatabaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("DATABASE_URL is not a valid url: %w", err)
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return fmt.Errorf("DATABASE_URL must start with postgres:// or postgresql://")
	}
	if parsed.Hostname() == "" {
		return fmt.Errorf("DATABASE_URL has no host")
	}

	self.PostgresHost = parsed.Hostname()
	self.PostgresPort = 5432
	if port := parsed.Port(); port != "" {
		number, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("DATABASE_URL port %q is not a number", port)
		}
		self.PostgresPort = number
	}
	if user := parsed.User.Username(); user != "" {
		self.PostgresUser = user
	}
	if password, set := parsed.User.Password(); set {
		self.PostgresPassword = password
	}
	if db := strings.TrimPrefix(parsed.Path, "/"); db != "" {
		self.PostgresDB = db
	}
	if sslMode := parsed.Query().Get("sslmode"); sslMode != "" {
		self.PostgresSSLMode = sslMode
	}
	return nil
}
