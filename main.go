package main

import (
	"strings"
)

func main() {
	cleanInput("hello world")
}

func cleanInput(text string) []string {
	var sanitizedText = strings.Trim(strings.ToLower(text), " ")
	return strings.Split(sanitizedText, " ")
}
