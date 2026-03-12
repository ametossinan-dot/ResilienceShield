package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(path string) {
	var err error
	DB, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal("Erreur ouverture SQLite:", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS transactions (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		method      TEXT NOT NULL,
		endpoint    TEXT NOT NULL,
		body        TEXT,
		status      TEXT DEFAULT 'pending',
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		synced_at   DATETIME
	);

	CREATE TABLE IF NOT EXISTS sync_logs (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		message     TEXT,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatal("Erreur création tables:", err)
	}

	log.Println("Base SQLite initialisée.")
}
