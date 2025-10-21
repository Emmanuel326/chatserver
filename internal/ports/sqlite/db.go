package sqlite

import (
	"log"

	"github.com/Emmanuel326/chatserver/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// InitDB initializes the database connection using the provided configuration.
func InitDB(cfg *config.Config) *sqlx.DB {
	// sqlx.Connect is a helper that calls sql.Open() and then db.Ping()
	// The driver is "sqlite3" and the DSN is the file path.
	db, err := sqlx.Connect("sqlite3", cfg.DB_FILE)
	if err != nil {
		log.Fatalf("FATAL: Could not connect to the SQLite database: %v", err)
	}

	// SQLite-specific: enable foreign key constraints for safety
	// Must be run every time a connection is opened.
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("FATAL: Could not enable foreign keys for SQLite: %v", err)
	}

	log.Printf("✅ Successfully connected to SQLite file: %s", cfg.DB_FILE)
	return db
}
