package database

import (
	"os"
	"testing"
)

func TestInitializeDatabase(t *testing.T) {
	// Use in-memory database for testing
	testDBPath := ":memory:"
	
	err := InitializeDatabase(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer Close()

	// Verify WAL mode is enabled
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to check journal mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("Expected WAL mode, got: %s", journalMode)
	}

	// Verify foreign keys are enabled
	var fkEnabled int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		t.Fatalf("Failed to check foreign keys: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("Expected foreign keys enabled (1), got: %d", fkEnabled)
	}

	// Verify tables exist
	tables := []string{"devices", "cycles", "rdg_status", "audit_log"}
	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		err := db.QueryRow(query, table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("Table %s does not exist", table)
		}
	}

	// Verify indexes exist
	indexes := []string{
		"idx_cycles_device_id",
		"idx_cycles_start_ts",
		"idx_rdg_status_device_id",
		"idx_rdg_status_timestamp",
		"idx_audit_log_timestamp",
		"idx_audit_log_entity",
	}
	for _, index := range indexes {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?"
		err := db.QueryRow(query, index).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check index %s: %v", index, err)
		}
		if count != 1 {
			t.Errorf("Index %s does not exist", index)
		}
	}
}

func TestInitializeDatabaseCreatesFile(t *testing.T) {
	// Test with actual file
	testDBPath := "./test_data/test.db"
	defer os.RemoveAll("./test_data")

	err := InitializeDatabase(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer Close()

	// Verify file was created
	if _, err := os.Stat(testDBPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", testDBPath)
	}
}

func TestInitializeDatabaseIdempotent(t *testing.T) {
	// Test that running initialization twice doesn't fail
	testDBPath := ":memory:"
	
	err := InitializeDatabase(testDBPath)
	if err != nil {
		t.Fatalf("First initialization failed: %v", err)
	}

	// Run again
	err = InitializeDatabase(testDBPath)
	if err != nil {
		t.Fatalf("Second initialization failed: %v", err)
	}
	defer Close()

	// Verify tables still exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='devices'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to verify table: %v", err)
	}
	if count != 1 {
		t.Errorf("Table devices does not exist after second initialization")
	}
}

