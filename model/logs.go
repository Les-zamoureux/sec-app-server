package model

import (
	"fmt"
	"sec-app-server/db"
	"time"
)

type LogEntry struct {
	ID        int       `json:"id"`
	UserID    string    `json:"user_id"`
	Method    string    `json:"method"`
	URL       string    `json:"url"`
	Timestamp time.Time `json:"timestamp"`
}

func DeleteLogByID(id string) error {
	_, err := db.DB.Exec("DELETE FROM logs WHERE id = $1", id)
	if err != nil {
		fmt.Println("Erreur lors de la suppression du log :", err)
		return err
	}
	return nil
}

func GetAllLogs() ([]LogEntry, error) {
	rows, err := db.DB.Query("SELECT id, user_id, method, url, timestamp FROM logs ORDER BY timestamp DESC LIMIT 100")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry

	for rows.Next() {
		var logEntry LogEntry
		if err := rows.Scan(&logEntry.ID, &logEntry.UserID, &logEntry.Method, &logEntry.URL, &logEntry.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, logEntry)
	}

	return logs, nil
}
