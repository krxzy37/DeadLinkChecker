package main

import (
	"DeadLinkChecker/internal/scrapper"
	"DeadLinkChecker/internal/storage"
	"context"

	"path/filepath"
	"sync"

	"encoding/csv"
	"fmt"
	"net/url"
	"os"
	"strconv"

	_ "modernc.org/sqlite"
)

const (
	sqliteStoragePath = "data/sqlite/storage.db"
)

var finalData []scrapper.Page

func main() {

	var mu sync.Mutex
	rateLimiter := scrapper.NewRateLimiter()

	var userURL string
	jobs := make(chan string, 100)
	results := make(chan []string, 100)
	visited := make(map[string]bool)

	s, err := storage.New(sqliteStoragePath)
	if err != nil {
		panic(err)
	}
	defer s.Close()

	if err = s.Init(); err != nil {
		panic(err)
	}

	fmt.Println("Введите ссылку на сайт, который хотите обойти..")
	fmt.Println("Пример \"https://www.youtube.com/\", \"https://www.grailed.com/\"")
	fmt.Println("")

	_, err = fmt.Scan(&userURL)
	if err != nil {
		panic(err)
	}

	visited[userURL] = true

	for w := 1; w <= 30; w++ {
		go worker(&mu, w, jobs, results, s, rateLimiter)
	}

	parsedStart, err := url.Parse(userURL)
	if err != nil {
		panic(err)
	}
	targetHost := parsedStart.Host

	jobs <- userURL

	jobCount := 1

	for jobCount > 0 {
		currentLink := <-results

		jobCount--

		for _, link := range currentLink {

			if !visited[link] {
				visited[link] = true

				parsedLink, err := url.Parse(link)
				if err != nil {
					fmt.Println("Invalid url", link)
					continue
				}
				if parsedLink.Host == targetHost {
					jobCount++

					go func(l string) {
						jobs <- l
					}(link)
				}
			}
		}
	}
	close(jobs)

	err = writeToCsv(s, userURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	finalPages, err := s.GetPages(userURL)
	if err != nil {
		fmt.Printf("GetPages error in main: %v", err)
	}

	fmt.Println("Работа программы завершена.....")
	fmt.Printf("Всего найдено уникальных страниц: %d\n", len(finalPages))

}
func worker(m *sync.Mutex, id int, jobs <-chan string, results chan<- []string, db *storage.Storage, rL *scrapper.RateLimiter) {

	for link := range jobs {

		fmt.Printf("[Worker %d] Сканирую: %s\n", id, link)

		if err := rL.Limiter.Wait(context.Background()); err != nil {
			return
		}
		foundLinks, statusCode, err := scrapper.GetLinks(link)
		if statusCode == 429 {
			foundLinks, statusCode, err = scrapper.TryAgainExp(link, 6, scrapper.GetLinks)
		}

		if err != nil {
			fmt.Printf("[Worker %d] Ошибка на %s: %v\n", id, link, err)
			m.Lock()

			if err = db.Save(scrapper.Page{URL: link, StatusCode: statusCode}); err != nil {
				fmt.Printf("database save error: %v", err)
			}
			m.Unlock()
			results <- foundLinks

			continue
		}
		m.Lock()
		if err = db.Save(scrapper.Page{URL: link, StatusCode: statusCode}); err != nil {
			fmt.Printf("database save error: %v", err)
		}
		m.Unlock()

		results <- foundLinks

	}

}

func writeToCsv(db *storage.Storage, fileName string) error {

	results, err := db.GetPages(fileName)
	if err != nil {
		fmt.Printf("error GetPages: %v", err)
		return nil
	}

	dirPath := "csv/"

	tempLink, err := url.Parse(fileName)
	if err != nil {
		return fmt.Errorf("url parse err: %w", err)
	}
	fileName = tempLink.Host
	fileName = fileName + ".csv"

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("cant create or open csv directory: %w", err)
	}
	fullPath := filepath.Join(dirPath, fileName)

	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		return fmt.Errorf("cant open or create csv file: %w", err)

	}
	defer func() {
		err = file.Close()
		if err != nil {
			return
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headerRow := []string{"Link;", "StatusCode"}
	if err = writer.Write(headerRow); err != nil {
		return fmt.Errorf("cant write header row in a csv: %w", err)
	}

	for _, res := range results {
		row := []string{res.URL + ";", strconv.Itoa(res.StatusCode)}

		err := writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}
