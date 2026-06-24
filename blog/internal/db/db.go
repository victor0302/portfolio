package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Open opens (and pings) a SQLite database at path.
// Use ":memory:" for an in-memory database (useful in tests).
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite at %q: %w", path, err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite at %q: %w", path, err)
	}
	return db, nil
}
