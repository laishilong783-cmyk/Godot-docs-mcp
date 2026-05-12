package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database.
type DB struct {
	conn *sql.DB
}

// Open opens or creates the SQLite database.
func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath+"?_fk=on")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

// Close closes the database.
func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS documents (
  id INTEGER PRIMARY KEY,
  version TEXT NOT NULL,
  path TEXT NOT NULL,
  title TEXT,
  section TEXT,
  content TEXT NOT NULL,
  created_at TEXT DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(version, path)
);

CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
  title,
  path,
  content,
  content='documents',
  content_rowid='id'
);

CREATE TABLE IF NOT EXISTS symbols (
  id INTEGER PRIMARY KEY,
  version TEXT NOT NULL,
  kind TEXT NOT NULL,
  class_name TEXT,
  member_name TEXT,
  signature TEXT,
  return_type TEXT,
  description TEXT,
  path TEXT NOT NULL,
  line_start INTEGER,
  line_end INTEGER,
  created_at TEXT DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_symbols_version_kind ON symbols(version, kind);
CREATE INDEX IF NOT EXISTS idx_symbols_class ON symbols(version, class_name);
CREATE INDEX IF NOT EXISTS idx_symbols_member ON symbols(version, member_name);
CREATE INDEX IF NOT EXISTS idx_symbols_class_member ON symbols(version, class_name, member_name);
`
	_, err := db.conn.Exec(schema)
	return err
}

// ClearVersion removes all data for a specific version.
func (db *DB) ClearVersion(version string) error {
	_, err := db.conn.Exec("DELETE FROM documents WHERE version = ?", version)
	if err != nil {
		return err
	}
	_, err = db.conn.Exec("DELETE FROM symbols WHERE version = ?", version)
	return err
}

// InsertDocument inserts a document and returns its ID.
func (db *DB) InsertDocument(version, path, title, section, content string) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO documents (version, path, title, section, content, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(version, path) DO UPDATE SET
		   title=excluded.title, section=excluded.section, content=excluded.content, updated_at=excluded.updated_at`,
		version, path, title, section, content, time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// SyncFTS rebuilds the FTS index for documents.
func (db *DB) SyncFTS() error {
	_, err := db.conn.Exec("INSERT INTO documents_fts(documents_fts) VALUES('rebuild')")
	return err
}

// InsertSymbol inserts an API symbol.
func (db *DB) InsertSymbol(version string, kind, className, memberName, signature, returnType, description, path string, lineStart, lineEnd int) error {
	_, err := db.conn.Exec(
		`INSERT INTO symbols (version, kind, class_name, member_name, signature, return_type, description, path, line_start, line_end, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		version, kind, className, memberName, signature, returnType, description, path, lineStart, lineEnd, time.Now().Format(time.RFC3339),
	)
	return err
}

// Stats returns index statistics.
func (db *DB) Stats(version string) (map[string]int, error) {
	stats := map[string]int{}

	var docCount int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM documents WHERE version = ?", version).Scan(&docCount)
	if err != nil {
		return nil, err
	}
	stats["documents"] = docCount

	rows, err := db.conn.Query("SELECT kind, COUNT(*) FROM symbols WHERE version = ? GROUP BY kind", version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var kind string
		var count int
		if err := rows.Scan(&kind, &count); err != nil {
			return nil, err
		}
		stats[kind] = count
	}

	return stats, nil
}
