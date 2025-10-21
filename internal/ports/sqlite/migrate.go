package sqlite

import (
	"log"

	"github.com/jmoiron/sqlx"
)

// Migrate runs the initial database schema setup.
// In a real application, you'd use a dedicated migration library.
// For JIT learning, this single file is sufficient.
func Migrate(db *sqlx.DB) {
	// 1. Create the 'users' table
	usersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(usersTableSQL)
	if err != nil {
		log.Fatalf("FATAL: Failed to create users table: %v", err)
	}
	
	// TODO: Message table migration will be added here later
	
	log.Println("✅ Database schema migrated successfully (users table created/exists).")
}
