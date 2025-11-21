-- Initial Database Schema Migration
-- Story: 1.2 - SQLite Database Setup and Schema
-- Created: 2025-11-21

-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Enable WAL mode (Write-Ahead-Logging) for better concurrency
PRAGMA journal_mode = WAL;

-- Devices table
CREATE TABLE IF NOT EXISTS devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    model TEXT,
    manufacturer TEXT NOT NULL,  -- 'Melag' or 'Getinge'
    ip TEXT NOT NULL,
    serial TEXT,
    type TEXT NOT NULL,  -- 'Steri' or 'RDG'
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
    result TEXT,  -- 'OK' or 'NOK'
    error_code TEXT,
    error_description TEXT,
    phase TEXT,  -- 'Aufheizen', 'Sterilisation', 'Trocknung'
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
    reachable INTEGER NOT NULL,  -- 0 or 1 (boolean)
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
);

-- Indexes for rdg_status table
CREATE INDEX IF NOT EXISTS idx_rdg_status_device_id ON rdg_status(device_id);
CREATE INDEX IF NOT EXISTS idx_rdg_status_timestamp ON rdg_status(timestamp);

-- Audit log table
CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    action TEXT NOT NULL,  -- 'device_added', 'cycle_started', 'cycle_completed', etc.
    entity_type TEXT,  -- 'device', 'cycle', 'rdg_status'
    entity_id INTEGER,
    user TEXT,
    details TEXT,  -- JSON details
    hash TEXT  -- Integrity hash for audit trail verification
);

-- Indexes for audit_log table
CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_log_entity ON audit_log(entity_type, entity_id);

