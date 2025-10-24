package cli

import (
	"bufio"
	"io"
	"strings"
)

// InputReader handles user input
type InputReader struct {
	scanner *bufio.Scanner
}

// NewInputReader creates a new InputReader
func NewInputReader(reader io.Reader) *InputReader {
	return &InputReader{
		scanner: bufio.NewScanner(reader),
	}
}

// ReadGuess reads a guess from user input
func (r *InputReader) ReadGuess() (string, bool) {
	if !r.scanner.Scan() {
		return "", false
	}
	return strings.TrimSpace(r.scanner.Text()), true
}

// IsQuitCommand checks if the input is a quit command
func IsQuitCommand(input string) bool {
	lower := strings.ToLower(input)
	return lower == "quit" || lower == "exit"
}
