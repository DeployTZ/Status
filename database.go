package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var db *sql.DB

const dbFile = "status.db"
const historyLimitDays = 90 // Show history for the last 90 days

type StatusRecord struct {
	Timestamp      time.Time `json:"timestamp"`
	IsUp           bool      `json:"is_up"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	StatusCode     int       `json:"status_code"` // Store status code
}

// InitDB initializes the database connection and creates tables if they don't exist.
func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", dbFile+"?_journal_mode=WAL") // WAL mode is generally better for concurrency
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Create table if not exists
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS status_history (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        is_up BOOLEAN NOT NULL,
        response_time_ms INTEGER NOT NULL,
        status_code INTEGER NOT NULL
    );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating database table: %v", err)
	}

	// Optional: Create index for faster queries
	createIndexSQL := `CREATE INDEX IF NOT EXISTS idx_timestamp ON status_history (timestamp);`
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		log.Printf("Warning: Could not create index: %v", err) // Non-fatal
	}

	log.Println("Database initialized successfully.")
}

// AddStatusRecord saves a new status check result to the database.
func AddStatusRecord(isUp bool, responseTimeMs int64, statusCode int) error {
	insertSQL := `INSERT INTO status_history (is_up, response_time_ms, status_code) VALUES (?, ?, ?)`
	_, err := db.Exec(insertSQL, isUp, responseTimeMs, statusCode)
	if err != nil {
		log.Printf("Error inserting status record: %v", err)
		return err
	}
	return nil
}

// GetRecentStatusHistory retrieves status records within the historyLimitDays.
func GetRecentStatusHistory() ([]StatusRecord, error) {
	records := []StatusRecord{}
	cutoffDate := time.Now().AddDate(0, 0, -historyLimitDays)

	querySQL := `
    SELECT timestamp, is_up, response_time_ms, status_code
    FROM status_history
    WHERE timestamp >= ?
    ORDER BY timestamp ASC` // Get oldest first for timeline rendering

	rows, err := db.Query(querySQL, cutoffDate)
	if err != nil {
		log.Printf("Error querying status history: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rec StatusRecord
		var ts string // Read timestamp as string first
		err := rows.Scan(&ts, &rec.IsUp, &rec.ResponseTimeMs, &rec.StatusCode)
		if err != nil {
			log.Printf("Error scanning status row: %v", err)
			continue // Skip bad rows
		}
		// Parse timestamp, assuming UTC storage (default for CURRENT_TIMESTAMP in SQLite)
		rec.Timestamp, err = time.Parse("2006-01-02 15:04:05", ts) // Adjust format if needed
		if err != nil {
			// Try parsing with timezone offset if stored differently
			rec.Timestamp, err = time.Parse(time.RFC3339, ts)
			if err != nil {
				log.Printf("Error parsing timestamp string '%s': %v", ts, err)
				continue
			}
		}
		records = append(records, rec)
	}

	return records, rows.Err()
}

// GetCurrentStatus retrieves the most recent status record.
func GetCurrentStatus() (*StatusRecord, error) {
	var rec StatusRecord
	var ts string
	querySQL := `
    SELECT timestamp, is_up, response_time_ms, status_code
    FROM status_history
    ORDER BY timestamp DESC
    LIMIT 1`

	row := db.QueryRow(querySQL)
	err := row.Scan(&ts, &rec.IsUp, &rec.ResponseTimeMs, &rec.StatusCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No status yet
		}
		log.Printf("Error querying current status: %v", err)
		return nil, err
	}

	// Parse timestamp
	rec.Timestamp, err = time.Parse("2006-01-02 15:04:05", ts) // Adjust format if needed
	if err != nil {
		rec.Timestamp, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			log.Printf("Error parsing timestamp string '%s': %v", ts, err)
			// Return record with zero time maybe? Or handle differently
			return &rec, nil // Return partially parsed record
		}
	}

	return &rec, nil
}

// CalculateUptime calculates uptime percentage for given periods.
func CalculateUptime(duration time.Duration) (float64, error) {
	cutoff := time.Now().Add(-duration)
	var totalChecks, upChecks int

	querySQL := `SELECT COUNT(*), SUM(CASE WHEN is_up = 1 THEN 1 ELSE 0 END)
                 FROM status_history
                 WHERE timestamp >= ?`

	err := db.QueryRow(querySQL, cutoff).Scan(&totalChecks, &upChecks)
	if err != nil {
		if err == sql.ErrNoRows || totalChecks == 0 {
			return 100.0, nil // Assume 100% if no data
		}
		log.Printf("Error calculating uptime for %v: %v", duration, err)
		return 0, err
	}

	if totalChecks == 0 {
		return 100.0, nil // Avoid division by zero
	}

	return (float64(upChecks) / float64(totalChecks)) * 100.0, nil
}
