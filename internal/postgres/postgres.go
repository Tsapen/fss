package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	fss "github.com/Tsapen/fss/internal/fss"
)

const (
	constraintViolationCode = "23503"
)

// Config contains settings for db.
type Config struct {
	UserName    string
	Password    string
	Port        string
	VirtualHost string
	HostName    string
}

// DB contains db connection.
type DB struct {
	*sqlx.DB
}

func (c *Config) dbAddr() string {
	return fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		c.UserName,
		c.Password,
		c.HostName,
		c.Port,
		c.VirtualHost,
	)
}

// New creates new storage.
func New(c Config) (*DB, error) {
	dbAddr := c.dbAddr()
	db, err := sqlx.Open("postgres", dbAddr)
	if err != nil {
		return nil, fmt.Errorf("open connection %s: %w", dbAddr, err)
	}

	for i := 0; i < 10; i++ {
		if err = db.Ping(); err == nil {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		return nil, fmt.Errorf("ping with connection %s: %w", dbAddr, err)
	}

	return &DB{
		db,
	}, nil
}

// CreateFile creates file in system.
func (s *DB) CreateFile(ctx context.Context, filename string) (int64, error) {
	query :=
		`INSERT INTO files (name, last_server_id, last_committed_at) 
			VALUES ($1, (SELECT id FROM servers ORDER by id DESC LIMIT 1), CURRENT_TIMESTAMP)
			RETURNING last_server_id
	`
	var lastServerID int64
	err := s.QueryRowContext(ctx, query, filename).Scan(&lastServerID)
	pqErr := new(pq.Error)
	if ok := errors.As(err, &pqErr); ok && pqErr.Code == constraintViolationCode {
		return 0, fss.NewConflictError("insert book: %w", err)
	}
	if err != nil {
		return 0, fss.NewInternalError("insert file: %w", err)
	}

	return lastServerID, nil
}

// File gets a file by name.
func (s *DB) File(ctx context.Context, name string) (*fss.File, error) {
	q := `SELECT name, last_server_id, last_committed_at, fragments FROM files f WHERE name=$1`
	file := new(fss.File)
	err := s.GetContext(ctx, file, q, name)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, fss.NewNotFoundError("file not found: %w", err)

	case err != nil:
		return nil, fss.NewInternalError("select file: %w", err)

	default:
		return file, nil
	}
}

// UpdateFile updates a file.
func (s *DB) UpdateFile(ctx context.Context, f *fss.File) (err error) {
	params := []any{f.LastCommittedAt, f.Fragments, f.Name}
	q := `UPDATE files f SET last_committed_at = $1, fragments = $2 WHERE name = $3`

	result, err := s.DB.ExecContext(ctx, q, params...)
	if err != nil {
		return fss.NewInternalError("update file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fss.NewInternalError("get the number of affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fss.NewNotFoundError("file with name '%s' not found", f.Name)
	}

	return nil
}

// DeleteFile deletes file by name.
func (s *DB) DeleteFile(ctx context.Context, name string) error {
	q := `DELETE FROM files f WHERE f.name = $1`
	result, err := s.ExecContext(ctx, q, name)
	if err != nil {
		return fss.NewInternalError("remove file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fss.NewInternalError("get the number of affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fss.NewNotFoundError("file '%s' not found ", name)
	}

	return nil
}

// Servers gets servers by last server id.
func (s *DB) Servers(ctx context.Context, lastServerID int64) ([]fss.Server, error) {
	q := "SELECT s.id, s.url FROM servers s WHERE s.id <= $1 ORDER BY id"
	rows, err := s.QueryContext(ctx, q, lastServerID)
	if err != nil {
		return nil, fss.NewInternalError("select servers: %w", err)
	}

	var servers []fss.Server
	if err = sqlx.StructScan(rows, &servers); err != nil {
		return nil, fss.NewInternalError("copy data into struct: %w", err)
	}

	return servers, nil
}

// CreateServer creates server in system.
func (s *DB) CreateServer(ctx context.Context, uri string) error {
	query := "INSERT INTO servers (url) VALUES ($1)"
	_, err := s.DB.ExecContext(ctx, query, uri)
	pqErr := new(pq.Error)
	if ok := errors.As(err, &pqErr); ok && pqErr.Code == constraintViolationCode {
		return fss.NewConflictError("insert server: %w", err)
	}

	if err != nil {
		return fss.NewInternalError("insert server: %w", err)
	}

	return nil
}
