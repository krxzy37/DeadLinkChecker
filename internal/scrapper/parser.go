package scrapper

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/PuerkitoBio/goquery"
)

type RateLimiter struct {
	Limiter *rate.Limiter
}

func NewRateLimiter() *RateLimiter {

	rl := rate.NewLimiter(10, 20)

	return &RateLimiter{
		Limiter: rl,
	}
}

type Page struct {
	URL        string
	StatusCode int
}

func GetLinks(targetURL string) ([]string, int, error) {

	var results []string
	client := &http.Client{
		Timeout: 45 * time.Second,
	}

	base, err := url.Parse(targetURL)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {

		switch {
		case resp.StatusCode == 429:
			return nil, 429, nil
		case resp.StatusCode == 404:
			fmt.Println("Битая ссылка -> ", targetURL)
			return nil, 404, fmt.Errorf("dead link, status code 404: %w", err)
		default:
			fmt.Printf("Unknown error on link -> %v", targetURL)
			return nil, resp.StatusCode, fmt.Errorf("status code %d, error: %w", resp.StatusCode, err)
		}

	}

	contentType := resp.Header.Get("Content-Type")

	if !strings.Contains(contentType, "text/html") {
		return nil, 0, fmt.Errorf("Response dont contains a text or html: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, 0, err
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

	return results, 200, nil
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

	if depth > 7 {
		return true
	}

	return false
}

func TryAgainExp(targetLink string, maxRetries int, o func(string) ([]string, int, error)) ([]string, int, error) {

	delay := 2 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {

		links, statusCode, err := o(targetLink)
		if err == nil {
			return links, statusCode, nil
		}

		if attempt == maxRetries-1 {
			return nil, 504, fmt.Errorf("out of tries: %w", err)
		}

		time.Sleep(delay)
		delay *= 2

	}

	return nil, 404, fmt.Errorf("Out of tries: (%d)", maxRetries)
}
