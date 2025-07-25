package main

import (
	"regexp"
	"strings"
)

func cleanInput(cmdArgs string) []string {
	var sanitizedText = strings.Trim(strings.ToLower(cmdArgs), " ")
	return strings.Split(sanitizedText, " ")
}

func validNoteInputFormat(note string) bool {
	// only letters, numbers, commas, and dashes
	pattern := `^[a-zA-Z0-9,-]+$`
	matched, err := regexp.MatchString(pattern, note)
	if err != nil {
		return false
	}

	return matched
}
