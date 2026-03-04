package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/url"
	"os"
)

type linkResult struct {
	URL      string
	isBroken string
}

func main() {

	var userURL string

	db, err := ConnectDB()
	if err != nil {
		fmt.Println("Ошибка БД: ", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}()

	jobs := make(chan string, 100)
	results := make(chan []string, 100)
	visited := make(map[string]bool)

	fmt.Println("Введите ссылку на сайт, который хотите обойти..")
	fmt.Println("Пример \"https://www.youtube.com/\", \"https://www.grailed.com/\"")
	fmt.Println("")

	_, errr := fmt.Scan(&userURL)
	if errr != nil {
		panic(err)
	}
	visited[userURL] = true

	for w := 1; w <= 10; w++ {
		go worker(w, jobs, results, db)
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
					fmt.Println("Битая ссылка", link)
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

	//err = ReadFromDb(db)
	//if err != nil {  fmt.Printf("ошибка создания графа для обсидиана: %v\n", err) }

	fmt.Println("Работа программы завершена.....")
	fmt.Printf("Всего найдено уникальных страниц: %d\n", len(visited))

}
func worker(id int, jobs <-chan string, results chan<- []string, db *sql.DB) {
	for link := range jobs {
		fmt.Printf("[Worker %d] Сканирую: %s\n", id, link)

		foundLinks, err := GetLinks(link)

		if err != nil {
			fmt.Printf("[Worker %d] Ошибка на %s: %v\n", id, link, err)
			_ = SaveResult(db, link, true, err.Error(), nil)
			results <- nil
			continue
		}
		_ = SaveResult(db, link, false, "none", foundLinks)
		results <- foundLinks
	}

}

func writeToCsv(results []linkResult) error {

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
		row := []string{res.URL, res.isBroken}

		err := writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}
