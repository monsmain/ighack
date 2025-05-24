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
	"golang.org/x/net/proxy"
)

const (
	API_URL      = "https://i.instagram.com/api/v1/"
	USER_AGENT   = "Instagram 76.0.0.15.395 Android (24/7.0; 640dpi; 1440x2560; samsung; SM-G930F; herolte; samsungexynos8890; en_US; 138226743)"
	TIMEOUT      = 10 * time.Second
	CURRENT_TIME = "2025-05-23 23:34:49"
	CURRENT_USER = "monsmain"
	// Tor Browser پورت پیش‌فرض برای
	TOR_PROXY    = "socks5://127.0.0.1:9150"  // توجه: پورت برای Tor Browser 9150 است
)

// ... بقیه کد مثل قبل ...
