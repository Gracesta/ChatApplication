package main

import (
	"database/sql"
	"log"
	"time"
)

func checkLogin(db *sql.DB, username, password string) (int, error) {
	var user_id int
	err := db.QueryRow("SELECT user_id FROM users WHERE username=? AND password=?", username, password).Scan(&user_id)
	if err != nil {
		return user_id, err
	}
	if user_id != 0 {
		return user_id, nil
	}
	return user_id, nil
}

func (client *Client) getUserByUsername(username string) (int, error) {
	var user_id int
	err := client.db.QueryRow("SELECT user_id FROM users WHERE username = ?", username).Scan(&user_id)
	return user_id, err
}

func (client *Client) insertUser(username string, password string) error {
	// Generate hash from the password
	hashedPassword := password
	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	// if err != nil {
	// 	return err
	// }

	// Insert the new user into the database
	_, err := client.db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

type Chatlog struct {
	Username  string
	Content   string
	Timestamp string
}

func loadChatLogsFromDatabase(client *Client) struct {
	Tile     string
	Chatlogs []Chatlog
} {
	db := client.db
	rows, err := db.Query("SELECT username, message, timestamp FROM users u JOIN chat_logs cl ON u.user_id = cl.user_id")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	data := struct {
		Tile     string
		Chatlogs []Chatlog
	}{}

	for rows.Next() {
		var username, chatlog, timestamp string
		err := rows.Scan(&username, &chatlog, &timestamp)

		// Check for errors when scanning the query results
		if err != nil {
			log.Fatal(err)
		}

		data.Chatlogs = append(data.Chatlogs, Chatlog{Username: username, Content: chatlog, Timestamp: timestamp})
	}
	return data
}

func timesStampMySQLFormat(timestamp string) string {
	t, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		panic(err.Error())
	}

	// Format the time.Time object into MySQL-compatible format
	mysqlTimestamp := t.Format("2006-01-02 15:04:05")
	return mysqlTimestamp
}
