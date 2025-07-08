package main

import (
	"fmt"
	"strings"
)

func main() {
	cleanInput("hello world")
}

func cleanInput(text string) []string {
	var sanitizedText = strings.Trim(strings.ToLower(text), " ")
	fmt.Printf("got: %s\n", text)
	fmt.Printf("santized into: %s\n", sanitizedText)
	return strings.Split(sanitizedText, " ")
}
