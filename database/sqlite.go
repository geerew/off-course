package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/geerew/off-course/migrations"
	"github.com/geerew/off-course/utils/security"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewSQLiteManager returns a DatabaseManager
func NewSQLiteManager(config *DatabaseManagerConfig) (*DatabaseManager, error) {
	manager := &DatabaseManager{}

	// When testing, pick a unique name
	dsnName := "data.db"
	if config.Testing {
		dsnName = fmt.Sprintf("data_memdb_%s", security.PseudorandomString(8))
	}

	// Data DB (writer)
	writeCfg := &databaseConfig{
		DataDir:    config.DataDir,
		DSN:        dsnName,
		MigrateDir: "data",
		AppFs:      config.AppFs,
		Testing:    config.Testing,
		Logger:     config.Logger,
		Mode:       "rwc",
	}

	writeDb, err := newSqliteDb(writeCfg)
	if err != nil {
		return nil, err
	}

	writeDb.DB().SetMaxOpenConns(1)
	writeDb.DB().SetMaxIdleConns(1)

	// Data DB (reader)
	readCfg := &databaseConfig{
		DataDir:    config.DataDir,
		DSN:        dsnName,
		MigrateDir: "",
		AppFs:      config.AppFs,
		Testing:    config.Testing,
		Logger:     config.Logger,
		Mode:       "ro",
	}

	readDb, err := newSqliteDb(readCfg)
	if err != nil {
		return nil, err
	}

	readDb.DB().SetMaxOpenConns(10)
	readDb.DB().SetMaxIdleConns(5)

	manager.DataDb = &compositeDb{
		read:  readDb,
		write: writeDb,
	}

	// Log DB
	dsnName = "logs.db"
	if config.Testing {
		dsnName = fmt.Sprintf("logs_memdb_%s", security.PseudorandomString(8))
	}

	logsCfg := &databaseConfig{
		DataDir:    config.DataDir,
		DSN:        dsnName,
		MigrateDir: "logs",
		AppFs:      config.AppFs,
		Testing:    config.Testing,
		Logger:     nil,
		Mode:       "rwc",
	}

	logsDb, err := newSqliteDb(logsCfg)
	if err != nil {
		return nil, err
	}

	manager.LogsDb = logsDb

	return manager, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// CompositeDb - Read/write pools
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// compositeDb is a composite database that uses two sqlite databases for read and write
type compositeDb struct {
	read  *sqliteDb
	write *sqliteDb
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Query executes a query that returns rows, typically a SELECT statement (read pool)
//
// It implements the Database interface
func (c *compositeDb) Query(query string, args ...any) (*sql.Rows, error) {
	return c.read.Query(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// QueryRow executes a query that is expected to return at most one row (read pool)
//
// It implements the Database interface
func (c *compositeDb) QueryRow(query string, args ...any) *sql.Row {
	return c.read.QueryRow(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Exec executes a non-query SQL statement against the write pool, with automatic retry logic
// to handle SQLite lock contention. If the operation returns a “database is locked” or “table is locked”
// error, it will wait for an exponentially increasing backoff interval (up to defaultMaxLockRetries times)
// before retrying. Non-lock errors are returned immediately. If all retries fail, the final error is returned
// wrapped with the retry count
//
// It implements the Database interface
func (c *compositeDb) Exec(query string, args ...any) (sql.Result, error) {
	var (
		res sql.Result
		err error
	)

	for attempt := 0; attempt <= defaultMaxLockRetries; attempt++ {
		res, err = c.write.Exec(query, args...)
		if err == nil {
			// if attempt > 0 {
			// 	fmt.Printf("[db] Exec succeeded after %d retries\n", attempt)
			// }

			return res, nil
		}

		// Bail on a non-lock error
		if !isLockError(err) {
			return res, err
		}

		delay := getRetryInterval(attempt)
		// fmt.Printf("[db] Lock error on attempt %d: %v; retrying in %v\n", attempt, err, delay)

		// On the last attempt, stop retrying
		if attempt == defaultMaxLockRetries {
			break
		}

		time.Sleep(delay)
	}

	// fmt.Printf("[db] Exec failed after %d retries: %v\n", defaultMaxLockRetries, err)
	return res, fmt.Errorf("%w after %d retries", err, defaultMaxLockRetries)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RunInTransaction runs a function in a transaction (write pool)
//
// It implements the Database interface
func (c *compositeDb) RunInTransaction(ctx context.Context, fn func(context.Context) error) error {
	return c.write.RunInTransaction(ctx, fn)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DB returns the underlying sql.DB for the write pool
//
// It implements the Database interface
func (c *compositeDb) DB() *sql.DB {
	return c.write.DB()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetLogger sets the logger for both read and write databases
//
// It implements the Database interface
func (c *compositeDb) SetLogger(l *slog.Logger) {
	c.read.SetLogger(l)
	c.write.SetLogger(l)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// sqliteTx - Transaction wrapper
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// sqliteDb is a sqlite-specific transaction wrapper
type sqliteTx struct {
	*sql.Tx
	db *sqliteDb
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Exec executes a query within a transaction without returning any rows
func (tx *sqliteTx) Exec(query string, args ...any) (sql.Result, error) {
	tx.db.log(query, args...)
	return tx.Tx.Exec(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Query executes a query within a transaction that returns rows, typically a SELECT statement
func (tx *sqliteTx) Query(query string, args ...any) (*sql.Rows, error) {
	tx.db.log(query, args...)
	return tx.Tx.Query(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// QueryRow executes a query within a transaction that is expected to return at most one row
func (tx *sqliteTx) QueryRow(query string, args ...any) *sql.Row {
	tx.db.log(query, args...)
	return tx.Tx.QueryRow(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// sqliteDb - Single sqlite database
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// sqliteDb defines a sqlite database
type sqliteDb struct {
	conn   *sql.DB
	config *databaseConfig
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// newSqliteDb creates a new sqliteDb
func newSqliteDb(config *databaseConfig) (*sqliteDb, error) {
	sqliteDb := &sqliteDb{
		config: config,
	}

	if err := sqliteDb.bootstrap(); err != nil {
		return nil, err
	}

	if err := sqliteDb.migrate(); err != nil {
		return nil, err
	}

	return sqliteDb, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DB returns the underlying sql.DB
//
// It implements the Database interface
func (db *sqliteDb) DB() *sql.DB {
	return db.conn
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Query executes a query that returns rows, typically a SELECT statement
//
// It implements the Database interface
func (db *sqliteDb) Query(query string, args ...any) (*sql.Rows, error) {
	db.log(query, args...)
	return db.conn.Query(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// QueryRow executes a query that is expected to return at most one row
//
// It implements the Database interface
func (db *sqliteDb) QueryRow(query string, args ...any) *sql.Row {
	db.log(query, args...)
	return db.conn.QueryRow(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Exec executes a query without returning any rows
//
// It implements the Database interface
func (db *sqliteDb) Exec(query string, args ...any) (sql.Result, error) {
	db.log(query, args...)
	return db.conn.Exec(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RunInTransaction runs a function in a transaction
//
// It implements the Database interface
func (db *sqliteDb) RunInTransaction(ctx context.Context, txFunc func(context.Context) error) (err error) {
	// Check if there's an existing querier in the context
	existingQuerier := QuerierFromContext(ctx, nil)
	if existingQuerier != nil {
		return txFunc(ctx)
	}

	slqTx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	tx := &sqliteTx{
		Tx: slqTx,
		db: db,
	}

	// Set the querier in the context to use the transaction
	txCtx := WithQuerier(ctx, tx)

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	return txFunc(txCtx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetLogger sets the logger for the database
//
// It implements the Database interface
func (db *sqliteDb) SetLogger(l *slog.Logger) {
	db.config.Logger = l
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// bootstrap initializes the sqlite database connection
func (db *sqliteDb) bootstrap() error {
	if err := db.config.AppFs.Fs.MkdirAll(db.config.DataDir, os.ModePerm); err != nil {
		return err
	}

	pragmaParts := []string{
		"cache=shared",
		"_busy_timeout=10000",
		"_journal_mode=WAL",
		"_journal_size_limit=200000000",
		"_synchronous=NORMAL",
		"_foreign_keys=1",
		"_cache_size=-16000",
	}

	if db.config.Mode != "" {
		pragmaParts = append([]string{fmt.Sprintf("mode=%s", db.config.Mode)}, pragmaParts...)
	}

	pragma := strings.Join(pragmaParts, "&")

	dsn := fmt.Sprintf("file:%s?%s", filepath.Join(db.config.DataDir, db.config.DSN), pragma)
	if db.config.Testing {
		dsn += "&mode=memory"
	}

	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}

	conn.SetMaxIdleConns(1)
	conn.SetMaxOpenConns(1)

	db.conn = conn

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// migrate runs the goose migrations
func (db *sqliteDb) migrate() error {
	if db.config.MigrateDir == "" {
		return nil
	}

	goose.SetLogger(goose.NopLogger())
	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Up(db.conn, db.config.MigrateDir); err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// log logs the query and arguments to the logger
func (db *sqliteDb) log(query string, args ...any) {
	if db.config.Logger != nil {
		attrs := make([]any, 0, len(args))
		attrs = append(attrs, loggerType)

		for i, arg := range args {
			attrs = append(attrs, slog.Any(fmt.Sprintf("arg %d", i+1), arg))
		}

		db.config.Logger.Debug(
			query,
			attrs...,
		)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// retry intervals for SQLite lock errors
var defaultRetryIntervals = []time.Duration{
	50 * time.Millisecond,
	100 * time.Millisecond,
	150 * time.Millisecond,
	200 * time.Millisecond,
	300 * time.Millisecond,
	400 * time.Millisecond,
	500 * time.Millisecond,
	700 * time.Millisecond,
	1000 * time.Millisecond,
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// how many times we’ll retry before giving up
const defaultMaxLockRetries = 9

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isLockError returns true for any SQLite “locked” error.
func isLockError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "database is locked") ||
		strings.Contains(s, "table is locked")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// getRetryInterval picks a delay for the Nth retry
func getRetryInterval(attempt int) time.Duration {
	if attempt < 0 || attempt >= len(defaultRetryIntervals) {
		return defaultRetryIntervals[len(defaultRetryIntervals)-1]
	}
	return defaultRetryIntervals[attempt]
}
