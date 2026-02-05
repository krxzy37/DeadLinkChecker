package main

import (
	"fmt"
	"net/url"
)

func main() {

	jobs := make(chan string, 100)
	results := make(chan []string, 100)
	visited := make(map[string]bool)

	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	startURL := "https://ezgif.com/"
	visited[startURL] = true

	parsedStart, err := url.Parse(startURL)
	if err != nil {
		panic(err)
	}
	targetHost := parsedStart.Host

	jobs <- startURL

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
func worker(id int, jobs <-chan string, results chan<- []string) {
	for link := range jobs {
		fmt.Printf("[Worker %d] Сканирую: %s\n", id, link)

		foundLinks, err := GetLinks(link)

		if err != nil {
			fmt.Printf("[Worker %d] Ошибка на %s: %v\n", id, link, err)
			results <- nil
			continue
		}
		results <- foundLinks
	}

}
