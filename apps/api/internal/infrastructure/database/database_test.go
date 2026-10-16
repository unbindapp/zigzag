package database

import (
	"testing"

	"entgo.io/ent/dialect"
	_ "github.com/ncruces/go-sqlite3/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unbindapp/zigzag-api/config"
)

type sqliteConn struct{}

func (sqliteConn) DSN() string     { return "file:entclient?mode=memory&_pragma=foreign_keys(1)" }
func (sqliteConn) Dialect() string { return dialect.SQLite }
func (sqliteConn) Driver() string  { return "sqlite3" }

func TestGetSqlDbConnPostgres(t *testing.T) {
	conn, err := GetSqlDbConn(&config.Config{
		PostgresDB:       "zigzag",
		PostgresHost:     "127.0.0.1",
		PostgresPassword: "p@ss/word",
		PostgresPort:     54329,
		PostgresUser:     "user",
	})
	require.NoError(t, err)

	assert.Equal(t, "postgres://user:p%40ss%2Fword@127.0.0.1:54329/zigzag?sslmode=disable", conn.DSN())
	assert.Equal(t, dialect.Postgres, conn.Dialect())
	assert.Equal(t, "pgx", conn.Driver())
}

func TestGetSqlDbConnNeedsDatabase(t *testing.T) {
	_, err := GetSqlDbConn(&config.Config{PostgresUser: "user"})
	assert.Error(t, err)
}

func TestNewEntClientRegistersSchemaDefaults(t *testing.T) {
	client, _, err := NewEntClient(sqliteConn{})
	require.NoError(t, err)
	require.NoError(t, client.Schema.Create(t.Context()))

	user, err := client.User.Create().SetEmail("a@example.com").SetDisplayName("a").Save(t.Context())
	require.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.False(t, user.CreatedAt.IsZero())
	assert.Equal(t, "UTC", user.Timezone)
	assert.True(t, user.Active)
}
