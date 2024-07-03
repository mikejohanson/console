package db

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
	"modernc.org/sqlite"               // sqlite driver
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second

	UniqueViolation = "23505"
)

// SQL -.
type SQL struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Builder    squirrel.StatementBuilderType
	Pool       *sql.DB
	IsEmbedded bool
}

// New -.
func New(url string, opts ...Option) (*SQL, error) {
	db := &SQL{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(db)
	}

	db.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	var err error

	if strings.HasPrefix(url, "postgres://") {
		err = setupHostedDB(db, url)
		if err != nil {
			return nil, err
		}
	} else {
		err = setupEmbeddedDB(db)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func setupEmbeddedDB(db *SQL) error {
	db.IsEmbedded = true

	dirname, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	dbPath := filepath.Join(dirname, "device-management-toolkit", "console.db")

	db.Pool, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	_, err = db.Pool.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Pool.Close()

		return err
	}

	return nil
}

func setupHostedDB(db *SQL, url string) error {
	var err error

	db.Pool, err = sql.Open("pgx", url)
	if err != nil {
		return err
	}

	return nil
}

// Close -.
func (p *SQL) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

func CheckNotUnique(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == UniqueViolation {
			return true
		}
	}

	var sqlErr *sqlite.Error
	if errors.As(err, &sqlErr) {
		if sqlErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE || sqlErr.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY {
			return true
		}
	}

	return false
}

func CheckForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23503" {
			return true
		}
	}

	// SQLite constraint error check
	if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
		return true
	}

	return false
}
