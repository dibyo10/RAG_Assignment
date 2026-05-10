package parser

import (
	"os"
	"strings"
	"unicode/utf8"
)

func parseText(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	text := string(data)
	if !utf8.ValidString(text) {
		// Replace invalid bytes with replacement character
		text = strings.ToValidUTF8(text, "")
	}
	return text, nil
}
