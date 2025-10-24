package game

import (
	"strings"
	"unicode"
)

// LetterStatus represents the status of a letter in the guess
type LetterStatus int

const (
	// Miss means the letter is not in the answer
	Miss LetterStatus = iota
	// Present means the letter is in the answer but wrong spot
	Present
	// Hit means the letter is in the correct spot
	Hit
)

// GuessResult contains the result of a single guess
type GuessResult struct {
	Guess    string
	Statuses []LetterStatus
}

// ValidateWord checks if a word is valid (5 letters, alphabetic only)
func ValidateWord(word string) bool {
	if len(word) != 5 {
		return false
	}
	for _, ch := range word {
		if !unicode.IsLetter(ch) {
			return false
		}
	}
	return true
}

// EvaluateGuess compares the guess with the answer and returns the result
// This implements the exact Wordle scoring logic:
// 1. First pass: mark all exact matches (Hit)
// 2. Second pass: mark Present for remaining letters that exist in answer
// 3. Handle duplicate letters correctly
func EvaluateGuess(guess, answer string) GuessResult {
	guess = strings.ToUpper(guess)
	answer = strings.ToUpper(answer)

	result := GuessResult{
		Guess:    guess,
		Statuses: make([]LetterStatus, 5),
	}

	// Count available letters in answer (excluding exact matches)
	answerLetterCount := make(map[rune]int)
	for _, ch := range answer {
		answerLetterCount[ch]++
	}

	// First pass: identify all exact matches (Hit)
	for i := 0; i < 5; i++ {
		if guess[i] == answer[i] {
			result.Statuses[i] = Hit
			answerLetterCount[rune(guess[i])]--
		}
	}

	// Second pass: identify Present letters
	for i := 0; i < 5; i++ {
		if result.Statuses[i] == Hit {
			continue
		}

		ch := rune(guess[i])
		if count, exists := answerLetterCount[ch]; exists && count > 0 {
			result.Statuses[i] = Present
			answerLetterCount[ch]--
		} else {
			result.Statuses[i] = Miss
		}
	}

	return result
}

// FormatResult converts GuessResult to output string with O, ?, _
func FormatResult(result GuessResult) string {
	var sb strings.Builder
	for _, status := range result.Statuses {
		switch status {
		case Hit:
			sb.WriteString("O")
		case Present:
			sb.WriteString("?")
		case Miss:
			sb.WriteString("_")
		}
	}
	return sb.String()
}
