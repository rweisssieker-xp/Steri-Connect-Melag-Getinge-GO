package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"os"
	"path/filepath"
)

var db *sql.DB

// DB returns the global database connection
func DB() *sql.DB {
	return db
}

// InitializeDatabase creates the database file, enables WAL mode, and runs migrations
func InitializeDatabase(dbPath string) error {
	// Create data directory if it doesn't exist
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	// Open database connection
	var err error
	db, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=1")
	if err != nil {
		return err
	}

	// Enable WAL mode
	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		return err
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return err
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return err
	}

	// Run migrations
	if err := runMigrations(); err != nil {
		return err
	}

	return nil
}

// runMigrations executes the initial schema migration
func runMigrations() error {
	migrationSQL := `
	-- Enable foreign key constraints
	PRAGMA foreign_keys = ON;

	-- Enable WAL mode (Write-Ahead-Logging) for better concurrency
	PRAGMA journal_mode = WAL;

	-- Devices table
	CREATE TABLE IF NOT EXISTS devices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		model TEXT,
		manufacturer TEXT NOT NULL,
		ip TEXT NOT NULL,
		serial TEXT,
		type TEXT NOT NULL,
		location TEXT,
		created DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(ip, manufacturer)
	);

	-- Cycles table
	CREATE TABLE IF NOT EXISTS cycles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id INTEGER NOT NULL,
		program TEXT,
		start_ts DATETIME NOT NULL,
		end_ts DATETIME,
		result TEXT,
		error_code TEXT,
		error_description TEXT,
		phase TEXT,
		temperature REAL,
		pressure REAL,
		progress_percent INTEGER,
		FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
	);

	-- Indexes for cycles table
	CREATE INDEX IF NOT EXISTS idx_cycles_device_id ON cycles(device_id);
	CREATE INDEX IF NOT EXISTS idx_cycles_start_ts ON cycles(start_ts);

	-- RDG Status table (Getinge devices)
	CREATE TABLE IF NOT EXISTS rdg_status (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id INTEGER NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		reachable INTEGER NOT NULL,
		FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
	);

	-- Indexes for rdg_status table
	CREATE INDEX IF NOT EXISTS idx_rdg_status_device_id ON rdg_status(device_id);
	CREATE INDEX IF NOT EXISTS idx_rdg_status_timestamp ON rdg_status(timestamp);

	-- Audit log table
	CREATE TABLE IF NOT EXISTS audit_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		action TEXT NOT NULL,
		entity_type TEXT,
		entity_id INTEGER,
		user TEXT,
		details TEXT,
		hash TEXT
	);

	-- Indexes for audit_log table
	CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp);
	CREATE INDEX IF NOT EXISTS idx_audit_log_entity ON audit_log(entity_type, entity_id);
	`

	_, err := db.Exec(migrationSQL)
	return err
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

