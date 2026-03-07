package printer

import (
	"bytes"
	"strings"
	"testing"
)

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			maxWidth: 24,
			expected: "",
		},
		{
			name:     "Short string without wrap",
			input:    "Hello world",
			maxWidth: 24,
			expected: "Hello world",
		},
		{
			name:     "Exact max width string",
			input:    "123456789012345678901234",
			maxWidth: 24,
			expected: "123456789012345678901234",
		},
		{
			name:     "Simple wrap string",
			input:    "This string is exactly twenty four chars long not.",
			maxWidth: 24,
			expected: "This string is exactly\ntwenty four chars long\nnot.",
		},
		{
			name:     "Long string with long words",
			input:    "Supercalifragilisticexpialidocious is a very long word",
			maxWidth: 24,
			expected: "Supercalifragilisticexpi\nalidocious is a very\nlong word",
		},
		{
			name:     "Existing newlines preserved",
			input:    "Line 1\nLine 2\nLine 3 is long",
			maxWidth: 10,
			expected: "Line 1\nLine 2\nLine 3 is\nlong",
		},
		{
			name:     "Zero or negative max width",
			input:    "Hello world this should not wrap",
			maxWidth: 0,
			expected: "Hello world this should not wrap",
		},
		{
			name:     "Long word after some text",
			input:    "abc 1234567890123456789012345",
			maxWidth: 24,
			expected: "abc\n123456789012345678901234\n5",
		},
		{
			name:     "Blank lines kept",
			input:    "   \n   \n",
			maxWidth: 24,
			expected: "\n\n",
		},
		{
			name:     "Multiple spaces compressed",
			input:    "word1   word2      word3",
			maxWidth: 10,
			expected: "word1\nword2\nword3", // The words should ideally be space separated if they fit, but here word1 + " " + word2 = 11 > 10. Wait, "word1"(5) + " "(1) + "word2"(5) = 11.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.input, tt.maxWidth)
			if result != tt.expected {
				t.Errorf("WrapText() =\n%q\nwant:\n%q", result, tt.expected)
			}
		})
	}
}

func TestFormatText(t *testing.T) {
	input := "abcdefghijklmnopqrstuvwxyzäöüßABCDEFGHIJKLMNOPQRSTUVWXY"

	// 1. Expected CP850 values
	// ä -> 0x84
	// ö -> 0x94
	// ü -> 0x81
	// ß -> 0xE1

	// Expected wrapper result for maxLineLength 24:
	// "abcdefghijklmnopqrstuvwx" (24)
	// "yzäöüßABCDEFGHIJKLMNOPQR" (24)
	// "STUVWXY"                  (7)

	expectedString := "abcdefghijklmnopqrstuvwx\r\nyz\x84\x94\x81\xe1ABCDEFGHIJKLMNOPQR\r\nSTUVWXY\r\n"

	result, err := FormatText(input)
	if err != nil {
		t.Fatalf("FormatText() unexpected error: %v", err)
	}

	if !bytes.Equal(result, []byte(expectedString)) {
		t.Errorf("FormatText() =\n%x\nwant:\n%x", result, []byte(expectedString))

		// Easy printable strings for failing debug
		t.Logf("Result string: %q", string(result))
		t.Logf("Expected string: %q", string([]byte(expectedString)))
	}

	// Check for CRLF endings
	if !strings.Contains(string(result), "\r\n") {
		t.Errorf("FormatText() did not replace \\n with \\r\\n")
	}
}
