package main

import (
	"DeadLinkChecker/internal/scrapper"
	"DeadLinkChecker/internal/storage"

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

/*
type LinkResult struct {
	URL      string
	IsBroken string
}
*/

var finalData []scrapper.Page

func main() {

	var userURL string
	var workers int
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

	fmt.Println("Введите скорость обхода сайта (от 3 до 10)")
	_, err = fmt.Scan(&workers)
	if err != nil {
		panic(err)
	}

	visited[userURL] = true

	for w := 1; w <= workers; w++ {
		go worker(w, jobs, results, s)
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

	err = writeToCsv(finalData)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Работа программы завершена.....")
	fmt.Printf("Всего найдено уникальных страниц: %d\n", len(visited))

}
func worker(id int, jobs <-chan string, results chan<- []string, db *storage.Storage) {
	for link := range jobs {
		fmt.Printf("[Worker %d] Сканирую: %s\n", id, link)

		foundLinks, err := scrapper.GetLinks(link)

		if err != nil {
			fmt.Printf("[Worker %d] Ошибка на %s: %v\n", id, link, err)
			finalData = append(finalData, scrapper.Page{URL: link, IsDead: true})
			db.Save(scrapper.Page{URL: link, IsDead: true})
			results <- nil
			continue
		}
		finalData = append(finalData, scrapper.Page{URL: link, IsDead: false})
		db.Save(scrapper.Page{URL: link, IsDead: false})

		results <- foundLinks
	}

}

func writeToCsv(results []scrapper.Page) error {

	fileName := "data.csv"

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println("ошибка записи csv файла" + err.Error())
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			return
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, res := range results {
		row := []string{res.URL, strconv.FormatBool(res.IsDead)}

		err := writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}
