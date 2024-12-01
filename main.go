package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var mysqldb *sql.DB

// Initialize the MySQL connection
func initDB() {
	var err error
	//MySQL credentials
	dsn := "root:root1234@tcp(127.0.0.1:3306)/time_logger"
	mysqldb, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error while opening database: ", err)
	}
}

func createTable() {
	var err error
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS time_log (
		id INT AUTO_INCREMENT PRIMARY KEY,
		timestamp VARCHAR(255) NOT NULL
	)`
	_, err = mysqldb.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

}

// Handler function to fetch and log the current time in Toronto
func currentTimeHandler(w http.ResponseWriter, r *http.Request) {
	// Get current time in Toronto
	currentTime := time.Now().In(time.FixedZone("Toronto", -5*60*60)) // UTC-5 for Toronto

	// Log the time into the database
	if _, err := mysqldb.Exec("INSERT INTO time_log (timestamp) VALUES (?)", currentTime); err != nil {
		http.Error(w, "Error logging time", http.StatusInternalServerError)
		return
	}

	// Return the current time as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"current_time": currentTime.Format(time.RFC3339),
	})
}

func getLoggedTimesHandler(w http.ResponseWriter, r *http.Request) {
	// Query logged times from the database
	rows, err := mysqldb.Query("SELECT timestamp FROM time_log")
	if err != nil {
		http.Error(w, "Error fetching logged times", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Collect timestamps into a slice
	var times []string
	for rows.Next() {
		var timestamp string
		if err := rows.Scan(&timestamp); err != nil {
			http.Error(w, "Error reading row", http.StatusInternalServerError)
			return
		}
		times = append(times, timestamp)
	}

	// Return logged times as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(times)
}

func main() {
	// Initialize the database and set up routes
	initDB()
	defer mysqldb.Close()

	err := mysqldb.Ping()

	if err != nil {
		log.Fatalf("Error pinging the database: %v", err)
	}

	fmt.Println("Connected to the database!")

	createTable()

	fmt.Println("Table created")

	http.HandleFunc("/current-time", currentTimeHandler)
	http.HandleFunc("/logged-times", getLoggedTimesHandler)

	// Start the server
	port := "8080"
	fmt.Println("Starting server on port", port)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
