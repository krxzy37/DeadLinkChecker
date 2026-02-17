package main

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetLinks(targetURL string) ([]string, error) {
	var results []string

	base, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(targetURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			return
		}
	}()

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
