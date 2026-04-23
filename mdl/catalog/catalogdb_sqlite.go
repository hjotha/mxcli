// SPDX-License-Identifier: Apache-2.0

//go:build !js

package catalog

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// SqliteCatalogDB wraps *sql.DB from modernc.org/sqlite.
type SqliteCatalogDB struct {
	db *sql.DB
}

// NewSqliteCatalogDB opens an in-memory SQLite database.
func NewSqliteCatalogDB() (*SqliteCatalogDB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}
	return &SqliteCatalogDB{db: db}, nil
}

// NewSqliteCatalogDBFromFile opens a file-based SQLite database.
func NewSqliteCatalogDBFromFile(path string) (*SqliteCatalogDB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}
	return &SqliteCatalogDB{db: db}, nil
}

func (s *SqliteCatalogDB) Query(query string, args ...any) (*sql.Rows, error) {
	return s.db.Query(query, args...)
}

func (s *SqliteCatalogDB) QueryRow(query string, args ...any) *sql.Row {
	return s.db.QueryRow(query, args...)
}

func (s *SqliteCatalogDB) Exec(query string, args ...any) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func (s *SqliteCatalogDB) Begin() (CatalogTx, error) {
	return s.db.Begin()
}

func (s *SqliteCatalogDB) Close() error {
	return s.db.Close()
}

// RawDB returns the underlying *sql.DB. Used only for SQLite-specific
// operations like VACUUM INTO in SaveToFile.
func (s *SqliteCatalogDB) RawDB() *sql.DB {
	return s.db
}

// WrapSqlDB wraps an existing *sql.DB as a CatalogDB.
// Used by tests that create their own in-memory databases.
func WrapSqlDB(db *sql.DB) *SqliteCatalogDB {
	return &SqliteCatalogDB{db: db}
}
