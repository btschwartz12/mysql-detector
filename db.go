// db.go
package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "var/client_logs.db")
	if err != nil {
		return nil, err
	}
	createTableSQL := `CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		client_ip TEXT,
		input TEXT,
		response TEXT
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, err
	}
	return db, nil
}

func LogRequestResponse(db *sql.DB, clientIP, input, response string) error {
	insertSQL := `INSERT INTO logs(client_ip, input, response) VALUES (?, ?, ?)`
	_, err := db.Exec(insertSQL, clientIP, input, response)
	return err
}

func GetAllLogs(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT * FROM logs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []string
	for rows.Next() {
		var id int
		var timestamp string
		var clientIP string
		var input string
		var response string
		err = rows.Scan(&id, &timestamp, &clientIP, &input, &response)
		if err != nil {
			return nil, err
		}
		logs = append(logs, timestamp, clientIP, input, response)
	}
	return logs, nil
}