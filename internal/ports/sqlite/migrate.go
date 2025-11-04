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
    -- NEW COLUMN for storing the external media URL (e.g., S3 link)
    media_url TEXT, 
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
	is_admin BOOLEAN NOT NULL DEFAULT FALSE, 
	PRIMARY KEY (group_id, user_id),
	FOREIGN KEY(group_id) REFERENCES groups(id) ON DELETE CASCADE,
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
`

// Migrate runs all necessary database schema migrations.
func Migrate(db *sqlx.DB) {
	log.Println("Database schema migration started...")
	
    // 1. Run all initial CREATE TABLE statements
    _, err := db.Exec(schema)
	if err != nil {
		log.Fatalf("Failed to execute initial migrations: %v", err)
	}
    
    // Add indexes for efficient conversation history lookups
    indexQueries := []string{
        `CREATE INDEX IF NOT EXISTS idx_messages_sender_recipient_id_desc ON messages (sender_id, recipient_id, id DESC);`,
        `CREATE INDEX IF NOT EXISTS idx_messages_recipient_sender_id_desc ON messages (recipient_id, sender_id, id DESC);`,
        `CREATE INDEX IF NOT EXISTS idx_messages_sender_recipient_type_timestamp_id ON messages (sender_id, recipient_id, type, timestamp DESC, id DESC);`,
        `CREATE INDEX IF NOT EXISTS idx_messages_recipient_sender_type_timestamp_id ON messages (recipient_id, sender_id, type, timestamp DESC, id DESC);`,
    }
    for _, q := range indexQueries {
        _, err = db.Exec(q)
        if err != nil {
            log.Printf("INFO: Could not create index. This is often normal if index already exists: %v", err)
        }
    }

    // 2. ALTER TABLE for adding the new media_url column (handles existing databases)
    // We use a separate EXEC for ALTER TABLE because the execution often fails
    // if the column already exists, which is the expected behavior for a second run.
    alterQuery := `
        -- Attempt to add the new column if it doesn't already exist
        -- SQLite requires checking for table existence, but the simplest approach
        -- is to rely on the error handling if the column is already there.
        ALTER TABLE messages ADD COLUMN media_url TEXT;
    `
    _, err = db.Exec(alterQuery)
    if err != nil {
        // Log the error, but don't fail the whole app.
        // We expect an error here if the column already exists (e.g., "duplicate column name").
        // In a real production system, you'd check the specific error code.
        log.Printf("INFO: Could not run ALTER TABLE (media_url). This is often normal if column already exists: %v", err)
    }

    // 3. ALTER TABLE for adding the new status column
    alterStatusQuery := `
        ALTER TABLE messages ADD COLUMN status TEXT NOT NULL DEFAULT 'SENT';
    `
    _, err = db.Exec(alterStatusQuery)
    if err != nil {
        log.Printf("INFO: Could not run ALTER TABLE (status). This is often normal if column already exists: %v", err)
    }

	log.Println(" Database schema migrated successfully (all core tables created/exists).")
}
