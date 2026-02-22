package main

import (
	"database/sql"
	"fmt"
	"net/url"
)

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
