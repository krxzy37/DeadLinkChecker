package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
)

var count = 0

func chek(err error) {
	if err != nil {
		fmt.Println(err)
		fmt.Println("ERRRRRROOOOOR")

		os.Exit(1)
	}
}

func main() {
	URL := "https://ezgif.com/"

	base, err := url.Parse(URL)
	chek(err)

	resp, err := http.Get(URL)
	chek(err)
	defer func() {
		err := resp.Body.Close()
		chek(err)
	}()

	if resp.StatusCode > 400 {
		fmt.Println("StatusCode:", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	chek(err)

	parse := doc.Find("a").Each(func(index int, item *goquery.Selection) {

		link, exist := item.Attr("href")
		if exist {
			refURL, err := url.Parse(link)
			if err != nil {
				fmt.Printf("Ссылка %v кривая, скип", link)
				return
			}
			fullLink := base.ResolveReference(refURL)
			count++
			fmt.Printf("Ссылка %v, ее индекс %v\n", fullLink, count)

		}
	})
	//chek(err)
	parse.Size()
}
