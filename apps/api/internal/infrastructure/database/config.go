package database

import (
	"fmt"
	"net/url"

	"entgo.io/ent/dialect"
	"github.com/unbindapp/zigzag-api/config"
)

type SqlDBConn interface {
	DSN() string
	Dialect() string
	Driver() string
}

type PostgresConn struct {
	Host     string
	Port     int
	Password string
	User     string
	DBName   string
	SSLMode  string
}

func (c *PostgresConn) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:     "/" + c.DBName,
		RawQuery: url.Values{"sslmode": {c.SSLMode}}.Encode(),
	}
	return u.String()
}

func (c *PostgresConn) Dialect() string { return dialect.Postgres }

func (c *PostgresConn) Driver() string { return "pgx" }

func GetSqlDbConn(cfg config.ConfigInterface) (SqlDBConn, error) {
	if cfg.GetPostgresDB() == "" || cfg.GetPostgresUser() == "" {
		return nil, fmt.Errorf("postgres database and user must be set")
	}
	sslMode := cfg.GetPostgresSSLMode()
	if sslMode == "" {
		sslMode = "disable"
	}
	return &PostgresConn{
		Host:     cfg.GetPostgresHost(),
		Port:     cfg.GetPostgresPort(),
		Password: cfg.GetPostgresPassword(),
		User:     cfg.GetPostgresUser(),
		DBName:   cfg.GetPostgresDB(),
		SSLMode:  sslMode,
	}, nil
}
