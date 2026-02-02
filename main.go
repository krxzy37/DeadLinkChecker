package main

import (
	"fmt"
	"os"
)

func main() {
	URL := "https://ezgif.com/"

	fmt.Printf("Сканируется %v\n", URL)

	foundLinks, err := GetLinks(URL)
	if err != nil {
		fmt.Printf("Ошибка программы: %v\n", err)
		os.Exit(1)
	}

	for i, link := range foundLinks {
		i++
		fmt.Printf("%d. %s\n", i, link)
	}
}
