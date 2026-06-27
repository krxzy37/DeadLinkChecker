package scrapper

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type RateLimiter struct {
	M         sync.Mutex
	RateLimit int
	Remaining int
	ResetAt   time.Time
}

func NewRateLimiter(remaining int, rateLimit int, resetAt time.Time) *RateLimiter {

	return &RateLimiter{
		RateLimit: rateLimit,
		Remaining: remaining,
		ResetAt:   resetAt,
	}
}

type Page struct {
	URL    string
	IsDead bool
}

func GetLinks(targetURL string) ([]string, error) {

	var results []string
	client := &http.Client{}

	base, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Битая ссылка -> ", targetURL)
		return nil, fmt.Errorf("status code %d, error", resp.StatusCode)

	}

	contentType := resp.Header.Get("Content-Type")

	if !strings.Contains(contentType, "text/html") {
		return nil, nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {

			rel, err := url.Parse(href)
			if err == nil {

				rel.RawQuery = ""
				rel.Fragment = ""
				abs := base.ResolveReference(rel)
				ext := strings.ToLower(path.Ext(abs.Path)) // abs.Path берет путь из url, Path.ext находит последнюю точку strings.ToLower переносит все в нижний регистр

				if abs.Scheme != "https" && abs.Scheme != "http" {
					return
				}
				if isTooDeep(abs.String()) {
					return
				}
				if isBadLink(abs.String()) {
					return
				}
				if shouldSkipExtension(ext) {
					return
				}

				results = append(results, abs.String())

			}
		}
	})

	return results, nil
}

func shouldSkipExtension(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".ico", ".bmp", ".tiff", ".mng", ".jxl", ".heic", ".avif":
		return true
	case ".mp4", ".mp3", ".avi", ".mkv", ".mov", ".wav", ".flac", ".webm", ".ogg", ".m4a":
		return true
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".csv", ".rtf":
		return true
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".exe", ".dmg", ".iso", ".apk", ".bin":
		return true
	case ".css", ".js", ".woff", ".woff2", ".ttf", ".eot":
		return true
	}
	return false
}

func isBadLink(link string) bool {

	segments := strings.Split(link, "/")
	seen := make(map[string]bool)

	for _, segment := range segments {
		if segment == "" {
			continue
		}
		segment = strings.ToLower(segment)

		if seen[segment] {
			return true
		}
		seen[segment] = true
	}
	return false
}

func isTooDeep(link string) bool {

	segments := strings.Split(link, "/")

	depth := len(segments) - 1

	if depth > 6 {
		return true
	}

	return false
}

func CheckRateLimits(link string) (string, string, time.Time, error) {

	var rateLimit string
	var rateRemaining string
	var rateReset string

	resp, err := http.Get(link)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("cant use GET request: %w", err)
	}
	defer resp.Body.Close()

	possibleLimitKeys := []string{
		"X-RateLimit-Limit",
		"X-Rate-Limit-Limit",
		"RateLimit-Limit",
	}

	for _, key := range possibleLimitKeys {
		rateLimit = resp.Header.Get(key)

		if rateLimit != "" {
			continue
		}
		return "", "", time.Time{}, fmt.Errorf("header is empty or undefined: %w", err)
	}

	possibleRemaningKeys := []string{
		"X-RateLimit-Remaining",
		"X-Rate-Limit-Remaining",
		"RateLimit-Remaining",
	}

	for _, key := range possibleRemaningKeys {
		rateRemaining = resp.Header.Get(key)

		if rateLimit != "" {
			continue
		}
		return "", "", time.Time{}, fmt.Errorf("header is empty or undefined: %w", err)
	}

	possibleResetKeys := []string{
		"X-RateLimit-Reset",
		"X-Rate-Limit-Reset",
		"RateLimit-Reset",
	}

	for _, key := range possibleResetKeys {
		rateReset = resp.Header.Get(key)

		if rateLimit != "" {
			continue
		}
		return "", "", time.Time{}, fmt.Errorf("header is empty or undefined: %w", err)
	}

	resetSec, err := strconv.ParseInt(rateReset, 10, 64)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("cant parse header ratelimit reset: %w", err)
	}

	resetTime := time.Unix(resetSec, 0)

	return rateLimit, rateRemaining, resetTime, nil
}
