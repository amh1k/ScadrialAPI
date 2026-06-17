//go:build e2e

package main

import (
	"database/sql"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var sharedTestDB *sql.DB

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://test:test@localhost/scadrial_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic(err)
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://../../migrations", "postgres", driver)
	if err != nil {
		panic(err)
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		panic(err)
	}

	sharedTestDB = db
	code := m.Run()
	db.Close()
	os.Exit(code)
}

func newTestDB(t *testing.T) *sql.DB {
	t.Cleanup(func() {
		_, err := sharedTestDB.Exec(`TRUNCATE movies, users, tokens, permissions, users_permissions RESTART IDENTITY CASCADE`)
		if err != nil {
			t.Fatalf("failed to reset test db: %v", err)
		}

		_, err = sharedTestDB.Exec(`
			INSERT INTO permissions (code)
			VALUES ('movies:read'), ('movies:write')
		`)
		if err != nil {
			t.Fatalf("failed to reseed permissions: %v", err)
		}
	})
	return sharedTestDB
}
