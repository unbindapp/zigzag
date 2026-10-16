package database

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/unbindapp/zigzag-api/ent"
	// without it every insert panics at runtime: schema defaults are registered there
	_ "github.com/unbindapp/zigzag-api/ent/runtime"
)

func NewEntClient(connInfo SqlDBConn) (*ent.Client, *sql.DB, error) {
	db, err := sql.Open(connInfo.Driver(), connInfo.DSN())
	if err != nil {
		return nil, nil, err
	}

	drv := entsql.OpenDB(connInfo.Dialect(), db)
	return ent.NewClient(ent.Driver(drv)), db, nil
}
