package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/internal/common/log"
)

//go:generate go run -mod=mod github.com/vburenin/ifacemaker -f "*.go" -i BaseRepositoryInterface -p repository -s BaseRepository -o repository_iface.go
type BaseRepository struct {
	DB *ent.Client
}

func NewBaseRepository(db *ent.Client) *BaseRepository {
	return &BaseRepository{DB: db}
}

func (r *BaseRepository) WithTx(ctx context.Context, fn func(tx TxInterface) error) error {
	tx, err := r.DB.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			log.Errorf("Panic caught in WithTX: %v", string(debug.Stack()))
			_ = tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

type TxInterface interface {
	Commit() error
	Rollback() error
	Client() *ent.Client
}

// IgnoreExisting treats an upsert that did nothing because the row exists as success; ent reports it as sql.ErrNoRows.
func IgnoreExisting(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func (r *BaseRepository) Client(tx TxInterface) *ent.Client {
	if tx != nil {
		return tx.Client()
	}
	return r.DB
}

func LockRowsOnPostgres(table string) func(*entsql.Selector) {
	return func(s *entsql.Selector) {
		if s.Dialect() == dialect.Postgres {
			s.ForUpdate(entsql.WithLockTables(table))
		}
	}
}
