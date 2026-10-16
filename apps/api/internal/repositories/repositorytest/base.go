package repositorytest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/ncruces/go-sqlite3/driver"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/enttest"
	entmigrate "github.com/unbindapp/zigzag-api/ent/migrate"
)

const PostgresURLEnv = "ZIGZAG_TEST_DATABASE_URL"

type RepositoryBaseSuite struct {
	suite.Suite
	Ctx      context.Context
	DB       *ent.Client
	Postgres bool
	dropDB   func()
}

func (suite *RepositoryBaseSuite) SetupTest() {
	suite.Ctx = context.Background()
	if suite.Postgres {
		suite.DB, suite.dropDB = openScratchPostgres(suite)
		return
	}
	suite.DB = enttest.Open(suite.T(), "sqlite3", "file:ent?mode=memory&_fk=1")
	err := suite.DB.Schema.Create(suite.Ctx)
	if err != nil {
		suite.T().Fatalf("failed creating schema resources: %v", err)
	}
}

func (suite *RepositoryBaseSuite) TearDownTest() {
	if err := suite.DB.Close(); err != nil {
		suite.T().Fatalf("failed closing database connection: %v", err)
	}
	suite.DB = nil
	if suite.dropDB != nil {
		suite.dropDB()
		suite.dropDB = nil
	}
}

func SkipWithoutPostgres(suite *RepositoryBaseSuite) {
	if !suite.Postgres {
		suite.T().Skip("needs Postgres: set " + PostgresURLEnv)
	}
}

func PostgresConfigured() bool {
	return os.Getenv(PostgresURLEnv) != ""
}

func openScratchPostgres(suite *RepositoryBaseSuite) (*ent.Client, func()) {
	t := suite.T()
	serverURL, err := url.Parse(os.Getenv(PostgresURLEnv))
	if err != nil || serverURL.Host == "" {
		t.Fatalf("%s is not a postgres:// URL", PostgresURLEnv)
	}
	admin, err := sql.Open("pgx", serverURL.String())
	if err != nil {
		t.Fatalf("connecting to postgres: %v", err)
	}
	suffix := make([]byte, 6)
	_, _ = rand.Read(suffix)
	name := "zigzag_test_" + hex.EncodeToString(suffix)
	if _, err := admin.ExecContext(suite.Ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatalf("creating scratch database: %v", err)
	}
	drop := func() {
		_, _ = admin.ExecContext(context.Background(), "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
		_ = admin.Close()
	}

	scratchURL := *serverURL
	scratchURL.Path = "/" + name
	db, err := sql.Open("pgx", scratchURL.String())
	if err != nil {
		drop()
		t.Fatalf("connecting to scratch database: %v", err)
	}
	goose.SetBaseFS(entmigrate.MigrationsFS)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		drop()
		t.Fatalf("goose dialect: %v", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		drop()
		t.Fatalf("migrating scratch database: %v", err)
	}
	return ent.NewClient(ent.Driver(entsql.OpenDB("postgres", db))), drop
}
