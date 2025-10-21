package sqlite

import (
	"log"

	"github.com/Emmanuel326/chatserver/internal/config"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite" // SQLite driver
)

// InitDB initializes the database connection using the provided configuration.
func InitDB(cfg *config.Config) *sqlx.DB {
	// Connect using the modernc.org/sqlite driver
	db, err := sqlx.Connect("sqlite", cfg.DB_FILE) // <- driver name must be "sqlite"
	if err != nil {
		log.Fatalf("FATAL: Could not connect to the SQLite database: %v", err)
	}

	// SQLite-specific: enable foreign key constraints for safety
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("FATAL: Could not enable foreign keys for SQLite: %v", err)
	}

	log.Printf("✅ Successfully connected to SQLite file: %s", cfg.DB_FILE)
	return db
}
