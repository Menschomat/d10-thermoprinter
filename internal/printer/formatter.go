package printer

import (
	"bytes"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const maxLineLength = 24

// FormatText takes input UTF-8 text, wraps it to 24 characters, aligns it,
// converts it to CP850 encoding, and ensures DOS line endings (CRLF).
func FormatText(input string, align string) ([]byte, error) {
	// 1. Wrap text
	wrapped := WrapText(input, maxLineLength)

	// 2. Ensure CRLF and apply alignment
	// Replace any isolated \n with \r\n, being careful not to replace already existing \r\n
	var crlfBuilder strings.Builder
	lines := strings.Split(strings.ReplaceAll(wrapped, "\r\n", "\n"), "\n")
	for _, line := range lines {
		aligned := alignLine(line, align, maxLineLength)
		crlfBuilder.WriteString(aligned)

		// If line matches maxLineLength, it implicitly auto-wraps on thermal printers.
		// Avoid appending CRLF in that case to prevent printing a blank second line.
		if len([]rune(aligned)) < maxLineLength {
			crlfBuilder.WriteString("\r\n")
		}
	}

	// 3. Convert to CP850
	encoded, err := encodeToCP850(crlfBuilder.String())
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

func alignLine(line string, align string, maxWidth int) string {
	switch align {
	case "left":
		line = strings.TrimLeft(line, " ")
	case "right":
		line = strings.TrimRight(line, " ")
	case "center":
		line = strings.TrimSpace(line)
	default:
		line = strings.TrimLeft(line, " ")
	}

	runes := []rune(line)
	length := len(runes)
	if length == 0 || length >= maxWidth {
		return string(runes)
	}

	spaces := maxWidth - length
	switch align {
	case "right":
		return strings.Repeat(" ", spaces) + string(runes)
	case "center":
		leftSpaces := spaces / 2
		return strings.Repeat(" ", leftSpaces) + string(runes)
	case "left":
		fallthrough
	default:
		return string(runes)
	}
}

// WrapText wraps the given text to a maximum line length.
// It prioritizes breaking at spaces. If a single word is longer than maxLineLength,
// it will be forcefully broken.
func WrapText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		words := strings.Fields(line)
		if len(words) == 0 {
			if i > 0 {
				result.WriteString("\n")
			}
			continue
		}

		wrapWords(&result, words, maxWidth)

		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func wrapWords(result *strings.Builder, words []string, maxWidth int) {
	currentLen := 0
	for j, word := range words {
		currentLen = processWord(result, word, currentLen, maxWidth, j == len(words)-1)
	}
}

func processWord(result *strings.Builder, word string, currentLen, maxWidth int, isLastWord bool) int {
	runes := []rune(word)

	// Handle words longer than the maximum width by force breaking them
	for len(runes) > maxWidth {
		if currentLen > 0 {
			result.WriteString("\n")
			currentLen = 0
		}
		result.WriteString(string(runes[:maxWidth]))
		result.WriteString("\n")
		runes = runes[maxWidth:]
	}

	if len(runes) == 0 {
		return currentLen
	}

	if currentLen+len(runes)+1 > maxWidth && currentLen > 0 {
		result.WriteString("\n")
		currentLen = 0
	}

	if currentLen > 0 {
		result.WriteString(" ")
		currentLen++
	}

	result.WriteString(string(runes))
	currentLen += len(runes)

	// Fast path: if we hit exact width, break line now to avoid a trailing space on next iter
	if currentLen == maxWidth && !isLastWord {
		result.WriteString("\n")
		currentLen = 0
	}

	return currentLen
}

func encodeToCP850(s string) ([]byte, error) {
	encoder := charmap.CodePage850.NewEncoder()
	var buf bytes.Buffer
	writer := transform.NewWriter(&buf, encoder)
	_, err := writer.Write([]byte(s))
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
