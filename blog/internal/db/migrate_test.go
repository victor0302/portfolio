package db

import (
	"testing"
)

func TestApply_FreshDB(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()

	if err := Apply(d); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	var n int
	if err := d.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n == 0 {
		t.Fatal("schema_migrations should have at least one row after Apply on a fresh DB")
	}

	// posts table should exist and be usable
	if _, err := d.Exec(`INSERT INTO posts (title, slug, body) VALUES ('x', 'x', 'x')`); err != nil {
		t.Errorf("insert into posts: %v", err)
	}
}

func TestApply_Idempotent(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()

	if err := Apply(d); err != nil {
		t.Fatalf("first Apply: %v", err)
	}
	var first int
	if err := d.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&first); err != nil {
		t.Fatalf("count after first apply: %v", err)
	}

	if err := Apply(d); err != nil {
		t.Fatalf("second Apply: %v", err)
	}
	var second int
	if err := d.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&second); err != nil {
		t.Fatalf("count after second apply: %v", err)
	}

	if first != second {
		t.Errorf("schema_migrations row count changed on second apply: %d -> %d", first, second)
	}
}

func TestApply_PreExistingPostsTable(t *testing.T) {
	// Simulates a DB that ran the old Apply (before the migration runner)
	// and already has the posts table but no schema_migrations.
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()

	if _, err := d.Exec(`CREATE TABLE posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE,
		body TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		published INTEGER NOT NULL DEFAULT 0
	)`); err != nil {
		t.Fatalf("seed posts: %v", err)
	}

	if err := Apply(d); err != nil {
		t.Fatalf("Apply on pre-existing posts table: %v", err)
	}
	var n int
	if err := d.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n == 0 {
		t.Error("Apply should record migrations even when target objects already exist")
	}
}
