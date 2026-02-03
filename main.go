package main

import (
	"fmt"
)

func main() {

	jobs := make(chan string, 100)
	results := make(chan []string, 100)

	startURL := "https://ezgif.com/"

	visited := make(map[string]bool)

	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	jobs <- startURL

	for currentLinks := range results {

		fmt.Printf("Было получено %d новых ссылок ---\n", len(currentLinks))

		for _, link := range currentLinks {
			if !visited[link] {
				visited[link] = true

				go func(l string) {
					jobs <- l
				}(link)
				fmt.Println("-> Добавлена в очередь:", link)
			}
		}
		fmt.Printf("Всего уникальных ссылок в базе: %d\n", len(visited))
	}
}
func worker(id int, jobs <-chan string, results chan<- []string) {
	for link := range jobs {
		fmt.Printf("[Worker %d] Сканирую: %s\n", id, link)

		foundLinks, err := GetLinks(link)

		if err != nil {
			fmt.Printf("[Worker %d] Ошибка на %s: %v\n", id, link, err)

			continue
		}
		results <- foundLinks
	}

}
