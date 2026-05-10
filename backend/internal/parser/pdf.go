package parser

import (
	"bytes"
	"fmt"
	"os/exec"
)

func parsePDF(path string) (string, error) {
	// Use pdftotext (poppler) if available
	if _, err := exec.LookPath("pdftotext"); err == nil {
		return pdftotextExtract(path)
	}
	return "", fmt.Errorf("pdftotext not found; install poppler-utils")
}

func pdftotextExtract(path string) (string, error) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd := exec.Command("pdftotext", "-layout", path, "-")
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext: %s: %w", errBuf.String(), err)
	}
	return out.String(), nil
}
