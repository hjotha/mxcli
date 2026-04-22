// SPDX-License-Identifier: Apache-2.0

package mpr

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func newTestWriterV1(t *testing.T, unitSchema string) (*Writer, *sql.DB) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.mpr")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(unitSchema); err != nil {
		t.Fatalf("failed to create Unit table: %v", err)
	}

	reader := &Reader{
		db:      db,
		version: MPRVersionV1,
	}
	return &Writer{reader: reader}, db
}

func TestInsertUnitV1_PopulatesContentsHash(t *testing.T) {
	writer, db := newTestWriterV1(t, `
		CREATE TABLE Unit (
			UnitID BLOB PRIMARY KEY NOT NULL,
			ContainerID BLOB,
			ContainmentName TEXT,
			TreeConflict LONG,
			ContentsHash TEXT,
			ContentsConflicts TEXT,
			Contents BLOB
		)
	`)

	unitID := "11111111-1111-1111-1111-111111111111"
	containerID := "22222222-2222-2222-2222-222222222222"
	contents := []byte("new microflow bytes")
	if err := writer.insertUnit(unitID, containerID, "Documents", "Microflows$Microflow", contents); err != nil {
		t.Fatalf("insertUnit failed: %v", err)
	}

	var gotHash string
	var gotContents []byte
	err := db.QueryRow(`SELECT ContentsHash, Contents FROM Unit WHERE UnitID = ?`, uuidToBlob(unitID)).Scan(&gotHash, &gotContents)
	if err != nil {
		t.Fatalf("failed to read inserted row: %v", err)
	}

	if gotHash == "" {
		t.Fatal("insertUnit wrote empty ContentsHash")
	}
	if want := contentHashBase64(contents); gotHash != want {
		t.Fatalf("ContentsHash = %q, want %q", gotHash, want)
	}
	if string(gotContents) != string(contents) {
		t.Fatalf("Contents = %q, want %q", string(gotContents), string(contents))
	}
}

func TestUpdateUnitV1_UpdatesContentsHash(t *testing.T) {
	writer, db := newTestWriterV1(t, `
		CREATE TABLE Unit (
			UnitID BLOB PRIMARY KEY NOT NULL,
			ContainerID BLOB,
			ContainmentName TEXT,
			TreeConflict LONG,
			ContentsHash TEXT,
			ContentsConflicts TEXT,
			Contents BLOB
		)
	`)

	unitID := "33333333-3333-3333-3333-333333333333"
	containerID := "44444444-4444-4444-4444-444444444444"
	oldContents := []byte("old bytes")
	newContents := []byte("updated bytes")
	if _, err := db.Exec(`
		INSERT INTO Unit (UnitID, ContainerID, ContainmentName, TreeConflict, ContentsHash, ContentsConflicts, Contents)
		VALUES (?, ?, 'Documents', 0, ?, '', ?)
	`, uuidToBlob(unitID), uuidToBlob(containerID), contentHashBase64(oldContents), oldContents); err != nil {
		t.Fatalf("failed to seed row: %v", err)
	}

	if err := writer.updateUnit(unitID, newContents); err != nil {
		t.Fatalf("updateUnit failed: %v", err)
	}

	var gotHash string
	var gotContents []byte
	err := db.QueryRow(`SELECT ContentsHash, Contents FROM Unit WHERE UnitID = ?`, uuidToBlob(unitID)).Scan(&gotHash, &gotContents)
	if err != nil {
		t.Fatalf("failed to read updated row: %v", err)
	}

	if gotHash == "" {
		t.Fatal("updateUnit wrote empty ContentsHash")
	}
	if want := contentHashBase64(newContents); gotHash != want {
		t.Fatalf("ContentsHash = %q, want %q", gotHash, want)
	}
	if string(gotContents) != string(newContents) {
		t.Fatalf("Contents = %q, want %q", string(gotContents), string(newContents))
	}
}

func TestUnitV1_OldSchemaWithoutContentsHashStillWorks(t *testing.T) {
	writer, db := newTestWriterV1(t, `
		CREATE TABLE Unit (
			UnitID BLOB PRIMARY KEY NOT NULL,
			ContainerID BLOB,
			ContainmentName TEXT,
			Type TEXT,
			Contents BLOB
		)
	`)

	unitID := "55555555-5555-5555-5555-555555555555"
	containerID := "66666666-6666-6666-6666-666666666666"
	initialContents := []byte("initial bytes")
	updatedContents := []byte("updated old schema bytes")

	if err := writer.insertUnit(unitID, containerID, "Documents", "Microflows$Microflow", initialContents); err != nil {
		t.Fatalf("insertUnit failed on old schema: %v", err)
	}
	if err := writer.updateUnit(unitID, updatedContents); err != nil {
		t.Fatalf("updateUnit failed on old schema: %v", err)
	}

	var gotType string
	var gotContents []byte
	err := db.QueryRow(`SELECT Type, Contents FROM Unit WHERE UnitID = ?`, uuidToBlob(unitID)).Scan(&gotType, &gotContents)
	if err != nil {
		t.Fatalf("failed to read old-schema row: %v", err)
	}

	if gotType != "Microflows$Microflow" {
		t.Fatalf("Type = %q, want %q", gotType, "Microflows$Microflow")
	}
	if string(gotContents) != string(updatedContents) {
		t.Fatalf("Contents = %q, want %q", string(gotContents), string(updatedContents))
	}
}
