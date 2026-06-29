// Package db opens the blog's SQLite database and applies schema
// migrations from numbered .sql files embedded at compile time.
//
// Migration files live in db/migrations/ and are named NNN_description.sql
// where NNN is a zero-padded sort key (001, 002, ...). Apply runs files in
// lexical order, recording each one in schema_migrations so it isn't
// re-applied on the next start. Each migration runs in its own transaction.
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const migrationsDir = "migrations"

// Apply runs every migration not yet recorded in schema_migrations.
// Safe to call repeatedly; idempotent on a clean DB and on one already
// migrated to the current set.
func Apply(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT     PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := loadAppliedVersions(db)
	if err != nil {
		return err
	}

	versions, files, err := listMigrations()
	if err != nil {
		return err
	}

	for _, v := range versions {
		if _, ok := applied[v]; ok {
			continue
		}
		body, err := fs.ReadFile(migrationsFS, migrationsDir+"/"+files[v])
		if err != nil {
			return fmt.Errorf("read migration %s: %w", v, err)
		}
		if err := runMigration(db, v, string(body)); err != nil {
			return err
		}
	}
	return nil
}

func runMigration(db *sql.DB, version, body string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx for %s: %w", version, err)
	}
	if _, err := tx.Exec(body); err != nil {
		tx.Rollback()
		return fmt.Errorf("exec migration %s: %w", version, err)
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
		tx.Rollback()
		return fmt.Errorf("record migration %s: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}
	return nil
}

func loadAppliedVersions(db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()
	applied := map[string]struct{}{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scan version: %w", err)
		}
		applied[v] = struct{}{}
	}
	return applied, rows.Err()
}

func listMigrations() ([]string, map[string]string, error) {
	entries, err := fs.ReadDir(migrationsFS, migrationsDir)
	if err != nil {
		return nil, nil, fmt.Errorf("read migrations dir: %w", err)
	}
	var versions []string
	files := map[string]string{}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		v := strings.TrimSuffix(name, ".sql")
		versions = append(versions, v)
		files[v] = name
	}
	sort.Strings(versions)
	return versions, files, nil
}
