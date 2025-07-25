package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
)

func main() {
	// Get connection parameters from environment variables or set defaults
	server := getEnvOrDefault("MSSQL_SERVER", "localhost")
	port := getEnvOrDefault("MSSQL_PORT", "1433")
	user := getEnvOrDefault("MSSQL_USER", "sa")
	password := getEnvOrDefault("MSSQL_PASSWORD", "YourPassword123!")
	database := getEnvOrDefault("MSSQL_DB", "master")

	// Build connection string
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s",
		server, user, password, port, database)

	fmt.Printf("Attempting to connect to MSSQL server: %s:%s\n", server, port)

	// Open connection
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to database: ", err.Error())
	}
	fmt.Println("✅ Connected to MSSQL database!")

	// Test write: create a table and insert a row
	fmt.Println("Creating test table...")
	_, err = db.Exec(`
		IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='test_table' AND xtype='U')
		CREATE TABLE test_table (
			id INT PRIMARY KEY,
			name NVARCHAR(50),
			created_at DATETIME DEFAULT GETDATE()
		)`)
	if err != nil {
		log.Fatal("Error creating table: ", err.Error())
	}
	fmt.Println("✅ Table created successfully!")

	// Insert a test row
	fmt.Println("Inserting test data...")
	_, err = db.Exec(`INSERT INTO test_table (id, name) VALUES (@p1, @p2)`, 1, "Test Name")
	if err != nil {
		// If row already exists, try updating it instead
		_, err = db.Exec(`UPDATE test_table SET name = @p1 WHERE id = @p2`, "Updated Test Name", 1)
		if err != nil {
			log.Fatal("Error inserting/updating row: ", err.Error())
		}
		fmt.Println("✅ Row updated successfully!")
	} else {
		fmt.Println("✅ Row inserted successfully!")
	}

	// Test read: query the data back
	fmt.Println("Reading test data...")
	var id int
	var name string
	var createdAt string
	err = db.QueryRow("SELECT id, name, created_at FROM test_table WHERE id = @p1", 1).Scan(&id, &name, &createdAt)
	if err != nil {
		log.Fatal("Error reading data: ", err.Error())
	}
	fmt.Printf("✅ Read data successfully: ID=%d, Name=%s, Created=%s\n", id, name, createdAt)

	fmt.Println("\n🎉 All database operations completed successfully!")

	// Set up HTTP endpoints
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>MSSQL Test Server</title>
</head>
<body>
    <h1>MSSQL Connection Test Server</h1>
    <p>✅ Database connection successful!</p>
    <p>✅ Table operations completed!</p>
    <p>Server is running and database is accessible.</p>
    <p>Last test data: ID=%d, Name=%s, Created=%s</p>
</body>
</html>`, id, name, createdAt)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","database":"connected","message":"MSSQL test server is running"}`)
	})

	// Start the HTTP server
	fmt.Println("🚀 Starting HTTP server on port 8080...")
	fmt.Println("Visit http://localhost:8080 to see the status")
	fmt.Println("Visit http://localhost:8080/health for health check")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
