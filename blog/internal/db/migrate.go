package db

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schemaSQL string

// Apply runs the embedded schema against db. Safe to call repeatedly
// because the schema uses CREATE TABLE/INDEX IF NOT EXISTS.
func Apply(db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
