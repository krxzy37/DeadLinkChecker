package main

import (
	"fmt"
	"net/http"
	"net/url"
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
				abs := base.ResolveReference(rel)

				if abs.Scheme != "https" && abs.Scheme != "http" {
					return
				}

				results = append(results, abs.String())

			}
		}
	})

	return results, nil
}
