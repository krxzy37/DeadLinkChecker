package main

import (
	"encoding/csv"
	"fmt"
	"net/url"
	"os"
)

type linkResult struct {
	URL      string
	isBroken string
}

var finalData []linkResult

func main() {

	var userURL string
	var workers int

	jobs := make(chan string, 100)
	results := make(chan []string, 100)
	visited := make(map[string]bool)

	fmt.Println("Введите ссылку на сайт, который хотите обойти..")
	fmt.Println("Пример \"https://www.youtube.com/\", \"https://www.grailed.com/\"")
	fmt.Println("")

	_, err := fmt.Scan(&userURL)
	if err != nil {
		panic(err)
	}

	fmt.Println("Введите скорость обхода сайта (от 3 до 50)")
	_, err = fmt.Scan(&workers)
	if err != nil {
		panic(err)
	}

	visited[userURL] = true

	for w := 1; w <= workers; w++ {
		go worker(w, jobs, results)
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
	err = writeToCsv(finalData)
	if err != nil {
		fmt.Println(err)
		return
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
			finalData = append(finalData, linkResult{URL: link, isBroken: "Error: " + err.Error()})
			results <- nil
			continue
		}
		finalData = append(finalData, linkResult{URL: link, isBroken: "OK"})

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
	writer.Comma = ';'

	for _, res := range results {
		row := []string{res.URL, res.isBroken}

		err := writer.Write(row)
		if err != nil {
			return err
		}
	}

	fmt.Println("Данные успешно записаны в файл...")

	return nil
}
