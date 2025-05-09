package database

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	loggerType = slog.Any("type", types.LogTypeDB)
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type contextKey string

const querierKey = contextKey("querier")

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithQuerier adds a querier to the context
func WithQuerier(ctx context.Context, querier Querier) context.Context {
	return context.WithValue(ctx, querierKey, querier)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// QuerierFromContext returns the querier from the context, defaulting to a defaulted querier if
// not found
func QuerierFromContext(ctx context.Context, defaultQuerier Querier) Querier {
	if querier, ok := ctx.Value(querierKey).(Querier); ok && querier != nil {
		return querier
	}

	return defaultQuerier
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Defines the sql functions
type (
	ExecFn     = func(query string, args ...interface{}) (sql.Result, error)
	QueryFn    = func(query string, args ...interface{}) (*sql.Rows, error)
	QueryRowFn = func(query string, args ...interface{}) *sql.Row
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Database defines the interface for a database
type Database interface {
	Querier
	DB() *sql.DB
	RunInTransaction(context.Context, func(context.Context) error) error
	SetLogger(*slog.Logger)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Querier interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DatabaseManager manages the database connections
type DatabaseManager struct {
	DataDb Database
	LogsDb Database
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DatabaseConfig defines the configuration for a database
type DatabaseConfig struct {
	DataDir    string
	DSN        string
	MigrateDir string
	AppFs      *appfs.AppFs
	InMemory   bool
	Logger     *slog.Logger
}
