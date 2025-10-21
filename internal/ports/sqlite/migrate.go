package sqlite

import (
	"log"

	"github.com/jmoiron/sqlx"
)

// SQL for table creation.
const schema = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	sender_id INTEGER NOT NULL,
	recipient_id INTEGER NOT NULL,
	type TEXT NOT NULL,
	content TEXT NOT NULL,
	timestamp DATETIME NOT NULL,
	-- Foreign keys (optional for SQLite, but good practice)
	FOREIGN KEY(sender_id) REFERENCES users(id),
	FOREIGN KEY(recipient_id) REFERENCES users(id)
);
`

// Migrate runs all necessary database schema migrations.
func Migrate(db *sqlx.DB) {
	log.Println("Database schema migration started...")
	_, err := db.Exec(schema)
	if err != nil {
		log.Fatalf("Failed to execute migrations: %v", err)
	}
	log.Println(" Database schema migrated successfully (users and messages tables created/exists).")
}

