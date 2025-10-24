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

-- NOTE: recipient_id can be a UserID (P2P) or a GroupID (Group Chat).
CREATE TABLE IF NOT EXISTS messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	sender_id INTEGER NOT NULL,
	recipient_id INTEGER NOT NULL,
	type TEXT NOT NULL,
	content TEXT NOT NULL,
	timestamp DATETIME NOT NULL,
	FOREIGN KEY(sender_id) REFERENCES users(id)
);

-- Group Chat Tables --
CREATE TABLE IF NOT EXISTS groups (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	owner_id INTEGER NOT NULL,
	created_at DATETIME NOT NULL,
	FOREIGN KEY(owner_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS group_members (
	group_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	joined_at DATETIME NOT NULL,
	is_admin BOOLEAN NOT NULL DEFAULT FALSE, -- <--- THE FIX
	PRIMARY KEY (group_id, user_id),
	FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
`

// Migrate runs all necessary database schema migrations.
func Migrate(db *sqlx.DB) {
	log.Println("Database schema migration started...")
	_, err := db.Exec(schema)
	if err != nil {
		log.Fatalf("Failed to execute migrations: %v", err)
	}
	log.Println(" Database schema migrated successfully (all core tables created/exists).")
}
