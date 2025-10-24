package game

import (
	"testing"
)

func TestValidateWord(t *testing.T) {
	tests := []struct {
		word  string
		valid bool
	}{
		{"APPLE", true},
		{"apple", true},
		{"APPL", false},   // too short
		{"APPLES", false}, // too long
		{"APP1E", false},  // contains number
		{"APP-E", false},  // contains special char
	}

	for _, tt := range tests {
		result := ValidateWord(tt.word)
		if result != tt.valid {
			t.Errorf("ValidateWord(%s) = %v, want %v", tt.word, result, tt.valid)
		}
	}
}

func TestEvaluateGuess(t *testing.T) {
	tests := []struct {
		guess    string
		answer   string
		expected []LetterStatus
	}{
		{
			guess:    "APPLE",
			answer:   "APPLE",
			expected: []LetterStatus{Hit, Hit, Hit, Hit, Hit},
		},
		{
			guess:    "BRAIN",
			answer:   "APPLE",
			expected: []LetterStatus{Miss, Miss, Present, Miss, Miss},
		},
		{
			guess:    "PLEAS",
			answer:   "APPLE",
			expected: []LetterStatus{Present, Present, Present, Present, Miss},
		},
		{
			// Test duplicate letters: SPEED vs ERASE
			// Positions: S P E E D vs E R A S E
			// No exact matches. ERASE has: E(2), R(1), A(1), S(1)
			// S-Present, P-Miss, E-Present (uses 1st E), E-Present (uses 2nd E), D-Miss
			guess:    "SPEED",
			answer:   "ERASE",
			expected: []LetterStatus{Present, Miss, Present, Present, Miss},
		},
		{
			// Test duplicate letters: GEESE vs ERASE
			// Positions: G E E S E vs E R A S E
			// Exact matches: pos 3 (S-S), pos 4 (E-E). Remaining: E(1), R(1), A(1)
			// G-Miss, E-Present (uses remaining E), E-Miss (no E left), S-Hit, E-Hit
			guess:    "GEESE",
			answer:   "ERASE",
			expected: []LetterStatus{Miss, Present, Miss, Hit, Hit},
		},
	}

	for _, tt := range tests {
		result := EvaluateGuess(tt.guess, tt.answer)
		for i, status := range result.Statuses {
			if status != tt.expected[i] {
				t.Errorf("EvaluateGuess(%s, %s)[%d] = %v, want %v",
					tt.guess, tt.answer, i, status, tt.expected[i])
			}
		}
	}
}

func TestFormatResult(t *testing.T) {
	result := GuessResult{
		Guess:    "APPLE",
		Statuses: []LetterStatus{Hit, Present, Miss, Hit, Present},
	}

	formatted := FormatResult(result)
	expected := "O?_O?"

	if formatted != expected {
		t.Errorf("FormatResult() = %s, want %s", formatted, expected)
	}
}
