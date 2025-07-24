package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	_ "github.com/denisenkom/go-mssqldb"
)

var (
	connectionLog []string
	logMutex      sync.Mutex
)

func logMessage(msg string) {
	log.Println(msg)
	logMutex.Lock()
	defer logMutex.Unlock()
	connectionLog = append(connectionLog, msg)
}

func connectToMSSQL(server, port, user, password, database string) (*sql.DB, error) {
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s", server, user, password, port, database)
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		logMessage("Failed to create DB handle: " + err.Error())
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		logMessage("Failed to connect to DB: " + err.Error())
		return nil, err
	}

	logMessage("Successfully connected to the SQL Server.")
	return db, nil
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	logMutex.Lock()
	defer logMutex.Unlock()
	w.Header().Set("Content-Type", "text/plain")
	for _, msg := range connectionLog {
		fmt.Fprintln(w, msg)
	}
}

func main() {
	// Change these with actual credentials or use environment variables
	server := os.Getenv("MSSQL_SERVER")   // e.g., "localhost"
	port := os.Getenv("MSSQL_PORT")       // e.g., "1433"
	user := os.Getenv("MSSQL_USER")       // e.g., "sa"
	password := os.Getenv("MSSQL_PASSWORD")
	database := os.Getenv("MSSQL_DB")     // e.g., "master"

	_, err := connectToMSSQL(server, port, user, password, database)
	if err != nil {
		logMessage("Error while connecting to SQL Server: " + err.Error())
	} else {
		logMessage("Connection established successfully.")
	}

	http.HandleFunc("/logs", logHandler)
	logMessage("Starting web server on :8080...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}
