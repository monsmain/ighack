package main

import (
    "bufio"
    "database/sql"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

// Attempt struct to store brute force attempt details
type Attempt struct {
    Username  string
    Password  string
    Timestamp string
    Success   bool
}

// simulateInstagramLogin simulates an Instagram login attempt
func simulateInstagramLogin(username, password string) bool {
    // For simulation, assume the correct username and password
    const correctUsername = "testuser"
    const correctPassword = "secret123"
    return username == correctUsername && password == correctPassword
}

// readPasswords reads passwords from password.txt
func readPasswords(filename string) ([]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("error opening %s: %v", filename, err)
    }
    defer file.Close()

    var passwords []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        password := strings.TrimSpace(scanner.Text())
        if password != "" {
            passwords = append(passwords, password)
        }
    }
    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("error reading %s: %v", filename, err)
    }
    return passwords, nil
}

// initDatabase initializes SQLite database
func initDatabase() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", "./instagram_attempts.db")
    if err != nil {
        return nil, fmt.Errorf("error connecting to database: %v", err)
    }

    // Create table for storing attempts
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS attempts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT,
            password TEXT,
            timestamp TEXT,
            success BOOLEAN
        )`)
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("error creating table: %v", err)
    }
    return db, nil
}

// logAttempt logs an attempt to the database
func logAttempt(db *sql.DB, username, password, timestamp string, success bool) error {
    _, err := db.Exec(
        "INSERT INTO attempts (username, password, timestamp, success) VALUES (?, ?, ?, ?)",
        username, password, timestamp, success)
    if err != nil {
        return fmt.Errorf("error logging attempt: %v", err)
    }
    return nil
}

// bruteForceAttack performs the brute force attack
func bruteForceAttack(username string, passwords []string, db *sql.DB, logger *log.Logger) (string, bool) {
    for i, password := range passwords {
        timestamp := time.Now().Format(time.RFC3339)
        fmt.Printf("Attempt %d: Trying password: %s for username: %s\n", i+1, password, username)
        logger.Printf("Attempt %d: Username: %s, Password: %s, Timestamp: %s", i+1, username, password, timestamp)

        // Simulate rate-limiting delay (Instagram typically limits requests)
        time.Sleep(500 * time.Millisecond)

        // Check password
        success := simulateInstagramLogin(username, password)

        // Log to database
        if err := logAttempt(db, username, password, timestamp, success); err != nil {
            log.Printf("Error logging attempt: %v", err)
        }

        if success {
            logger.Printf("Success: Password found: %s for username: %s", password, username)
            return password, true
        }
    }
    logger.Println("Password not found in the provided list")
    return "", false
}

func main() {
    // Prompt for username
    fmt.Print("Enter Instagram username: ")
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    username := strings.TrimSpace(scanner.Text())
    if username == "" {
        fmt.Println("Error: Username cannot be empty")
        return
    }

    // Initialize log file
    logFile, err := os.OpenFile("instagram_brute_force.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatalf("Error opening log file: %v", err)
    }
    defer logFile.Close()
    logger := log.New(logFile, "INSTAGRAM_BRUTE_FORCE: ", log.LstdFlags)

    // Initialize database
    db, err := initDatabase()
    if err != nil {
        log.Fatalf("Error initializing database: %v", err)
    }
    defer db.Close()

    // Read passwords from file
    passwords, err := readPasswords("password.txt")
    if err != nil {
        log.Fatalf("Error reading passwords: %v", err)
    }

    fmt.Printf("Starting Instagram brute force attack for username: %s with %d passwords\n", username, len(passwords))
    startTime := time.Now()

    // Run brute force attack
    foundPassword, success := bruteForceAttack(username, passwords, db, logger)

    if success {
        fmt.Printf("Success: Password found: %s for username: %s\n", foundPassword, username)
    } else {
        fmt.Println("Failed: Password not found in the provided list")
    }

    fmt.Printf("Time taken: %v\n", time.Since(startTime))
}
