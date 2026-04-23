// SPDX-License-Identifier: Apache-2.0

//go:build !js

package catalog

import (
	"database/sql"
	"testing"
)

// TestCatalogDBContract verifies that SqliteCatalogDB satisfies the CatalogDB
// interface contract: basic CRUD, transactions, and cleanup.
func TestCatalogDBContract(t *testing.T) {
	db, err := NewSqliteCatalogDB()
	if err != nil {
		t.Fatalf("NewSqliteCatalogDB: %v", err)
	}
	defer db.Close()

	// Exec — create a table.
	_, err = db.Exec("CREATE TABLE test_contract (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Exec CREATE TABLE: %v", err)
	}

	// Exec — insert a row.
	res, err := db.Exec("INSERT INTO test_contract (id, name) VALUES (?, ?)", 1, "alpha")
	if err != nil {
		t.Fatalf("Exec INSERT: %v", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		t.Fatalf("RowsAffected: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 row affected, got %d", n)
	}

	// QueryRow — read it back.
	var name string
	row := db.QueryRow("SELECT name FROM test_contract WHERE id = ?", 1)
	if err := row.Scan(&name); err != nil {
		t.Fatalf("QueryRow Scan: %v", err)
	}
	if name != "alpha" {
		t.Errorf("expected alpha, got %q", name)
	}

	// Query — list rows.
	rows, err := db.Query("SELECT id, name FROM test_contract ORDER BY id")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var id int
		var n string
		if err := rows.Scan(&id, &n); err != nil {
			t.Fatalf("rows.Scan: %v", err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestCatalogTxContract(t *testing.T) {
	db, err := NewSqliteCatalogDB()
	if err != nil {
		t.Fatalf("NewSqliteCatalogDB: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test_tx (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatalf("Exec CREATE TABLE: %v", err)
	}

	// Begin + Prepare + Exec + Commit.
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	stmt, err := tx.Prepare("INSERT INTO test_tx (id, val) VALUES (?, ?)")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	for i := 1; i <= 3; i++ {
		if _, err := stmt.Exec(i, "row"); err != nil {
			t.Fatalf("stmt.Exec(%d): %v", i, err)
		}
	}
	stmt.Close()
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	// Verify rows persisted.
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM test_tx").Scan(&count); err != nil {
		t.Fatalf("QueryRow after commit: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 rows after commit, got %d", count)
	}

	// Begin + Rollback — rows should not persist.
	tx2, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin (rollback): %v", err)
	}
	if _, err := tx2.Exec("INSERT INTO test_tx (id, val) VALUES (4, 'gone')"); err != nil {
		t.Fatalf("Exec in rollback tx: %v", err)
	}
	if err := tx2.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	if err := db.QueryRow("SELECT COUNT(*) FROM test_tx").Scan(&count); err != nil {
		t.Fatalf("QueryRow after rollback: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 rows after rollback, got %d", count)
	}
}

// TestInterfaceSatisfaction verifies compile-time interface compliance.
func TestInterfaceSatisfaction(t *testing.T) {
	var _ CatalogDB = (*SqliteCatalogDB)(nil)
	var _ CatalogTx = (*sql.Tx)(nil)
}
