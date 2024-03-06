package infrastracture

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/shoet/webpagesummary/pkg/config"
)

type DBHandler struct {
	db *sqlx.DB
}

type Transactor interface {
	DBReader
	Commit() error
	Rollback() error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type DBReader interface {
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

var _ Transactor = (*sqlx.Tx)(nil)
var _ DBReader = (*sqlx.DB)(nil)

func NewDBHandler(cfg *config.RDBConfig) (*DBHandler, error) {
	db, err := sql.Open("postgres", cfg.RDBDsn)
	if err != nil {
		return nil, fmt.Errorf("failed sql.Open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed db.Ping: %w", err)
	}
	return &DBHandler{
		db: sqlx.NewDb(db, "postgres"),
	}, nil
}

func (d *DBHandler) GetTransaction() (Transactor, error) {
	tx, err := d.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed db.Beginx: %w", err)
	}
	return tx, nil
}
