package main

import "fmt"

func main() {
	fmt.Println("test")
}

func cleanInput(text string) []string {
	// split user input into words on based whitespaces
	// lowercase everything
	// trim leading or trailing whitespaces
	// hello WORLD -> ["hello", "world"]
	var temp = []string{text}
	return temp
}
