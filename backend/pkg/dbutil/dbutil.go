// Package dbutil provides helpers to ensure a PostgreSQL database, schema, and
// tables exist before the application starts serving requests.
//
// Usage pattern (in each service's main.go):
//
//	if err := dbutil.EnsureDatabase(dsn, "my_db"); err != nil {
//	    log.Fatal(err)
//	}
//	db := <open gorm connection>
//	if err := dbutil.EnsureSchema(db.DB(), "my_schema"); err != nil {
//	    log.Fatal(err)
//	}
//	if err := db.AutoMigrate(&MyModel{}); err != nil {
//	    log.Fatal(err)
//	}
package dbutil

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
)

// EnsureDatabase connects to the postgres maintenance database and creates
// dbName if it does not exist. dsn must be a valid libpq / pgx DSN that
// already points to an existing database (e.g. "postgres").
//
// The function rewrites the DSN to target the "postgres" maintenance DB so
// that a CREATE DATABASE statement can be issued without a "database does not
// exist" error on startup.
func EnsureDatabase(dsn, dbName string) error {
	maintDSN, err := switchDatabase(dsn, "postgres")
	if err != nil {
		return fmt.Errorf("dbutil: parse DSN: %w", err)
	}

	db, err := sql.Open("pgx", maintDSN)
	if err != nil {
		return fmt.Errorf("dbutil: open maintenance DB: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("dbutil: ping maintenance DB: %w", err)
	}

	var exists bool
	row := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`, dbName)
	if err := row.Scan(&exists); err != nil {
		return fmt.Errorf("dbutil: check database existence: %w", err)
	}

	if !exists {
		// Database names cannot be parameterised in CREATE DATABASE.
		_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE %q`, dbName))
		if err != nil {
			return fmt.Errorf("dbutil: create database %q: %w", dbName, err)
		}
	}
	return nil
}

// EnsureSchema creates the named schema inside the already-connected database
// if it does not exist. Pass the *sql.DB obtained from gorm's DB() method.
func EnsureSchema(db *sql.DB, schema string) error {
	if schema == "" || schema == "public" {
		return nil
	}
	var exists bool
	row := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = $1)`, schema)
	if err := row.Scan(&exists); err != nil {
		return fmt.Errorf("dbutil: check schema existence: %w", err)
	}
	if !exists {
		if _, err := db.Exec(fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %q`, schema)); err != nil {
			return fmt.Errorf("dbutil: create schema %q: %w", schema, err)
		}
	}
	return nil
}

// switchDatabase rewrites the database name in a DSN string.
// Supports both keyword=value format and postgres:// URL format.
func switchDatabase(dsn, targetDB string) (string, error) {
	dsn = strings.TrimSpace(dsn)

	// URL format: postgres://user:pass@host:port/dbname?params
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return "", err
		}
		u.Path = "/" + targetDB
		return u.String(), nil
	}

	// Keyword=value format: host=... dbname=... user=...
	parts := strings.Fields(dsn)
	found := false
	for i, part := range parts {
		if strings.HasPrefix(part, "dbname=") {
			parts[i] = "dbname=" + targetDB
			found = true
			break
		}
	}
	if !found {
		parts = append(parts, "dbname="+targetDB)
	}
	return strings.Join(parts, " "), nil
}
