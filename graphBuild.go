package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func CreateFolder() {
	err := os.MkdirAll("obsidian_graph", 0755)
	if err != nil {
		fmt.Printf("папка не была создана: %v\n", err)
	}
}

func ClearName(url string) string {

	clean := strings.ReplaceAll(url, "https://", "")

	replace := strings.NewReplacer("/", "_", "?", "_", "=", "_", ":", "_")

	return replace.Replace(clean)
}

func ReadFromDb(db *sql.DB) error {

	rows, err := db.Query("SELECT url, connected_links FROM visited_links")
	if err != nil {
		return fmt.Errorf("ошибка запроса к бд: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sourceUrl string
		var linksJson []byte

		err := rows.Scan(&sourceUrl, &linksJson)
		if err != nil {
			continue
		}

		var childLinks []string
		err = json.Unmarshal(linksJson, &childLinks)
		if err != nil {
			fmt.Printf("error unMarshal: %v", err)
		}

		filename := "obsidian_graph/" + ClearName(sourceUrl) + ".md"

		file, err := os.Create(filename)
		if err != nil {
			fmt.Printf("ошибка создания файла: %v\n", err)
			continue
		}

		_, err = file.WriteString("Оригинальная ссылка -> " + sourceUrl + "\n\n")
		if err != nil {
			continue
		}

		for _, link := range childLinks {

			targetName := ClearName(link)

			fmt.Printf("[[%s]]\n", targetName)
		}

	}

	return nil
}
