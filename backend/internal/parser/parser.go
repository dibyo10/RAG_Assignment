package parser

import (
	"fmt"
	"strings"
)

func Parse(path, mimeType string) (string, error) {
	switch {
	case mimeType == "application/pdf" || strings.HasSuffix(path, ".pdf"):
		return parsePDF(path)
	default:
		return parseText(path)
	}
}

func DetectMIME(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".pdf"):
		return "application/pdf"
	case strings.HasSuffix(lower, ".txt"):
		return "text/plain"
	case strings.HasSuffix(lower, ".md"):
		return "text/markdown"
	default:
		return fmt.Sprintf("application/octet-stream")
	}
}
