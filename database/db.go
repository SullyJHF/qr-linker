package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type URL struct {
	ID        int       `json:"id"`
	FullURL   string    `json:"full_url"`
	ShortHash string    `json:"short_hash"`
	CreatedAt time.Time `json:"created_at"`
	Clicks    int       `json:"clicks"`
}

type DB struct {
	conn *sql.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.createTables(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		full_url TEXT NOT NULL,
		short_hash TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		clicks INTEGER DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_short_hash ON urls(short_hash);
	`

	_, err := db.conn.Exec(query)
	if err != nil {
		return err
	}

	log.Println("Database tables created successfully")
	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) CreateURL(fullURL, shortHash string) (*URL, error) {
	query := `
		INSERT INTO urls (full_url, short_hash, created_at, clicks)
		VALUES (?, ?, ?, 0)
	`

	result, err := db.conn.Exec(query, fullURL, shortHash, time.Now())
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &URL{
		ID:        int(id),
		FullURL:   fullURL,
		ShortHash: shortHash,
		CreatedAt: time.Now(),
		Clicks:    0,
	}, nil
}

func (db *DB) GetURLByHash(shortHash string) (*URL, error) {
	query := `
		SELECT id, full_url, short_hash, created_at, clicks
		FROM urls
		WHERE short_hash = ?
	`

	var url URL
	err := db.conn.QueryRow(query, shortHash).Scan(
		&url.ID,
		&url.FullURL,
		&url.ShortHash,
		&url.CreatedAt,
		&url.Clicks,
	)

	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (db *DB) IncrementClicks(shortHash string) error {
	query := `
		UPDATE urls
		SET clicks = clicks + 1
		WHERE short_hash = ?
	`

	_, err := db.conn.Exec(query, shortHash)
	return err
}

func (db *DB) GetAllURLs() ([]URL, error) {
	query := `
		SELECT id, full_url, short_hash, created_at, clicks
		FROM urls
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []URL
	for rows.Next() {
		var url URL
		err := rows.Scan(
			&url.ID,
			&url.FullURL,
			&url.ShortHash,
			&url.CreatedAt,
			&url.Clicks,
		)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	return urls, nil
}

func (db *DB) CheckHashExists(shortHash string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_hash = ?)`
	
	var exists bool
	err := db.conn.QueryRow(query, shortHash).Scan(&exists)
	return exists, err
}