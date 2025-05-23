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
    Password  string
    Timestamp string
    Success   bool
}

// simulateLogin simulates Instagram login attempt (replace with real API in controlled, legal tests)
func simulateLogin(username, password string) bool {
    // For simulation, assume the correct password is "secret123"
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
    db, err := sql.Open("sqlite3", "./password_attempts.db")
    if err != nil {
        return nil, fmt.Errorf("error connecting to database: %v", err)
    }

    // Create table for storing attempts
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS attempts (
            id INTEGER...
