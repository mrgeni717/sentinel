package store

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

type Store struct {
	DB *sql.DB
}

// Open opens (or creates) the SQLite database file at path and applies the
// schema. CREATE TABLE IF NOT EXISTS makes this safe to run on every
// startup - simpler than a full migration tool for a project this size,
// and a deliberate trade-off given the scope.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite handles one writer at a time; keep the pool small and
	// predictable rather than letting database/sql over-parallelize writes.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return &Store{DB: db}, nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}
