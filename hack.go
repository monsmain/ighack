package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// InstaConfig مدل اطلاعات کمکی
type InstaConfig struct{}

// idbear کوکی رو آنالیز می‌کنه و uid و bearer برمی‌گردونه
func (i *InstaConfig) idbear(cookie string) (string, string, error) {
	// فرض کن mid و bearer رو از کوکی استخراج کنیم (مثال ساده)
	reUID := regexp.MustCompile(`ds_user_id=(\d+);`)
	reBearer := regexp.MustCompile(`sessionid=(\w+);`)

	uidMatch := reUID.FindStringSubmatch(cookie)
	bearerMatch := reBearer.FindStringSubmatch(cookie)

	if len(uidMatch) < 2 || len(bearerMatch) < 2 {
		return "", "", fmt.Errorf("invalid cookie format")
	}

	return uidMatch[1], bearerMatch[1], nil
}

func (i *InstaConfig) pigeon() string {
	return fmt.Sprintf("sess-%d", rand.Intn(1000000))
}

func (i *InstaConfig) clintime() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

func (i *InstaConfig) deviceid() string {
	return "android-1234567890"
}

func (i *InstaConfig) family() string {
	return "family-1234567890"
}

func (i *InstaConfig) androidid() string {
	return "android-0987654321"
}

// مدل پاسخ JSON
type UserInfo struct {
	User struct {
		FullName string `json:"full_name"`
	} `json:"user"`
}

func checkCookie(cookie string) (bool, string, error) {
	// استخراج mid از کوکی
	mid := ""
	midRegex := regexp.MustCompile(`mid=(.*?);`)
	if strings.Contains(cookie, "mid=") {
		match := midRegex.FindStringSubmatch(cookie)
		if len(match) > 1 {
			mid = match[1]
		}
	}

	insta := &InstaConfig{}

	uid, bearer, err := insta.idbear(cookie)
	if err != nil {
		return false, "", fmt.Errorf("idbear error: %v", err)
	}

	client := &http.Client{}

	// ساخت ریکوئست با Context برای Timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://www.instagram.com/accounts/login/%s", uid), nil)
	if err != nil {
		return false, "", fmt.Errorf("request creation error: %v", err)
	}

	// هدرهای لازم
	req.Header.Set("x-ig-app-locale", "in_ID")
	req.Header.Set("x-ig-device-locale", "in_ID")
	req.Header.Set("x-ig-mapped-locale", "id_ID")
	req.Header.Set("x-pigeon-session-id", insta.pigeon())
	req.Header.Set("x-pigeon-rawclienttime", insta.clintime())
	req.Header.Set("x-ig-bandwidth-speed-kbps", fmt.Sprintf("%.3f", rand.Float64()*(10000-1000)+1000))
	req.Header.Set("x-ig-bandwidth-totalbytes-b", fmt.Sprintf("%d", rand.Intn(10000000-1000000)+1000000))
	req.Header.Set("x-ig-bandwidth-totaltime-ms", fmt.Sprintf("%d", rand.Intn(10000-1000)+1000))
	req.Header.Set("x-bloks-version-id", "ee55d61628b17424a72248a17431be7303200a6e7fa08b0de1736f393f1017bd")
	req.Header.Set("x-ig-www-claim", "0")
	req.Header.Set("x-debug-www-claim-source", "handleLogin3")
	req.Header.Set("x-bloks-prism-button-version", "CONTROL")
	req.Header.Set("x-bloks-prism-colors-enabled", "false")
	req.Header.Set("x-bloks-prism-ax-base-colors-enabled", "false")
	req.Header.Set("x-bloks-prism-font-enabled", "false")
	req.Header.Set("x-bloks-is-layout-rtl", "false")
	req.Header.Set("x-ig-device-id", insta.deviceid())
	req.Header.Set("x-ig-family-device-id", insta.family())
	req.Header.Set("x-ig-android-id", insta.androidid())
	req.Header.Set("x-ig-timezone-offset", fmt.Sprintf("%d", -int(time.Now().UTC().Unix())))
	req.Header.Set("x-ig-nav-chain", fmt.Sprintf("MainFeedFragment:feed_timeline:1:cold_start:%d.853::,com.bloks.www.bloks.ig.ndx.ci.entry.screen:com.bloks.www.bloks.ig.ndx.ci.entry.screen:2:button:%d.356::", time.Now().Unix(), time.Now().Unix()))
	req.Header.Set("x-fb-connection-type", "WIFI")
	req.Header.Set("x-ig-connection-type", "WIFI")
	req.Header.Set("x-ig-capabilities", "3brTv10=")
	req.Header.Set("x-ig-app-id", "567067343352427")
	req.Header.Set("priority", "u=3")
	req.Header.Set("user-agent", "Instagram 360.0.0.52.192 Android (28/9; 239dpi; 720x1280; google; G011A; G011A; intel; in_ID; 672535977)")
	req.Header.Set("accept-language", "id-ID, en-US")
	req.Header.Set("authorization", "Bearer IGT:2:"+bearer)
	req.Header.Set("x-mid", mid)
	req.Header.Set("ig-u-ds-user-id", uid)
	req.Header.Set("ig-intended-user-id", uid)
	req.Header.Set("x-fb-http-engine", "Liger")
	req.Header.Set("x-fb-client-ip", "True")
	req.Header.Set("x-fb-server-cluster", "True")
	req.Header.Set("Cookie", cookie)

	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("request error: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Errorf("read body error: %v", err)
	}

	var userInfo UserInfo
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return false, "", fmt.Errorf("json unmarshal error: %v", err)
	}

	fullName := userInfo.User.FullName

	// ساخت پوشه data در صورت نبودن
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		os.Mkdir("data", 0755)
	}

	// ذخیره کوکی در فایل
	err = ioutil.WriteFile("data/login.txt", []byte(cookie), 0644)
	if err != nil {
		return false, "", fmt.Errorf("write file error: %v", err)
	}

	if fullName != "" {
		return true, fullName, nil
	}

	return false, "", nil
}

func main() {
	// مقدار کوکی واقعی رو اینجا بذار
	cookie := "ds_user_id=123456; sessionid=abcdef1234567890; mid=XYZ123;"

	ok, name, err := checkCookie(cookie)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if ok {
		fmt.Println("Login successful! User full name:", name)
	} else {
		fmt.Println("Login failed!")
	}
}
