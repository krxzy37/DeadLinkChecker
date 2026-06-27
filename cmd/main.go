package main

import (
	"DeadLinkChecker/internal/scrapper"
	"DeadLinkChecker/internal/storage"
	"path/filepath"
	"time"

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

	limit, remaining, resetTime, err := scrapper.CheckRateLimits(userURL)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Лимит запросов таков: %v", limit)

	intRemaining, err := strconv.Atoi(remaining)
	if err != nil {
		println("cant convert atoi: %v", err)
	}
	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		println("cant convert atoi: %v", err)
	}

	rateLimiter := scrapper.NewRateLimiter(intRemaining, intLimit, resetTime)

	visited[userURL] = true

	for w := 1; w <= 50; w++ {
		go worker(w, jobs, results, s, rateLimiter)
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
					fmt.Println("code 404", link)
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

	err = writeToCsv(finalData, userURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Работа программы завершена.....")
	fmt.Printf("Всего найдено уникальных страниц: %d\n", len(visited))

}
func worker(id int, jobs <-chan string, results chan<- []string, db *storage.Storage, rL *scrapper.RateLimiter) {

	remain := rL.Remaining

	sleepTime := time.Until(rL.ResetAt)
	sleepTimeFloat := sleepTime.Seconds()
	sleepTimeSeconds := int(sleepTimeFloat)

	for link := range jobs {
		switch {

		case remain == 0:
			time.Sleep(time.Duration(sleepTimeSeconds))
			time.Sleep(500 * time.Millisecond)
			remain = remain + rL.RateLimit

		default:
			fmt.Printf("[Worker %d] Сканирую: %s\n", id, link)

			rL.M.Lock()
			foundLinks, err := scrapper.GetLinks(link)
			remain--
			rL.M.Unlock()

			if err != nil {
				fmt.Printf("[Worker %d] Ошибка на %s: %v\n", id, link, err)
				finalData = append(finalData, scrapper.Page{URL: link, IsDead: true})
				rL.M.Lock()
				db.Save(scrapper.Page{URL: link, IsDead: true})
				results <- nil
				rL.M.Unlock()
				continue
			}
			rL.M.Lock()
			finalData = append(finalData, scrapper.Page{URL: link, IsDead: false})
			db.Save(scrapper.Page{URL: link, IsDead: false})
			rL.M.Unlock()

			results <- foundLinks
		}
	}

}

func writeToCsv(results []scrapper.Page, fileName string) error {

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

	headerRow := []string{"Link;", "IsDead"}
	if err = writer.Write(headerRow); err != nil {
		return fmt.Errorf("cant write header row in a csv: %w", err)
	}

	for _, res := range results {
		row := []string{res.URL + ";", strconv.FormatBool(res.IsDead)}

		err := writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}
