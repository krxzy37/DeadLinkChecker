package main

import (
	"database/sql"
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
	clean = strings.ReplaceAll(clean, "http://", "")

	replace := strings.NewReplacer("/", "_", "?", "_", "=", "_", ":", "_")

	return replace.Replace(clean)
}

func ReadFromDb(*sql.DB) {}
