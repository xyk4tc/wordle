package game

import (
	"testing"
)

func TestNewGame(t *testing.T) {
	// Test valid game creation
	wordList := []string{"APPLE", "BRAIN", "CRANE"}
	game, err := NewGame(6, wordList)

	if err != nil {
		t.Errorf("NewGame() error = %v, want nil", err)
	}

	if game.MaxRounds != 6 {
		t.Errorf("NewGame() MaxRounds = %d, want 6", game.MaxRounds)
	}

	if game.CurrentRound != 0 {
		t.Errorf("NewGame() CurrentRound = %d, want 0", game.CurrentRound)
	}

	if game.Status != InProgress {
		t.Errorf("NewGame() Status = %v, want InProgress", game.Status)
	}

	// Test invalid max rounds
	_, err = NewGame(0, wordList)
	if err == nil {
		t.Error("NewGame(0, wordList) should return error")
	}

	// Test empty word list
	_, err = NewGame(6, []string{})
	if err == nil {
		t.Error("NewGame(6, []) should return error")
	}
}

func TestGameFlow(t *testing.T) {
	// Create a game with known answer
	wordList := []string{"APPLE"}
	game, _ := NewGame(6, wordList)

	// Make a wrong guess
	_, err := game.MakeGuess("BRAIN")
	if err != nil {
		t.Errorf("MakeGuess() error = %v, want nil", err)
	}

	if game.CurrentRound != 1 {
		t.Errorf("After 1 guess, CurrentRound = %d, want 1", game.CurrentRound)
	}

	if game.Status != InProgress {
		t.Errorf("After wrong guess, Status = %v, want InProgress", game.Status)
	}

	// Make correct guess
	_, err = game.MakeGuess("APPLE")
	if err != nil {
		t.Errorf("MakeGuess() error = %v, want nil", err)
	}

	if game.Status != Won {
		t.Errorf("After correct guess, Status = %v, want Won", game.Status)
	}

	// Try to guess after game is won
	_, err = game.MakeGuess("BRAIN")
	if err == nil {
		t.Error("MakeGuess() after game won should return error")
	}
}

func TestGameLoss(t *testing.T) {
	// Create a game with 2 max rounds
	wordList := []string{"APPLE"}
	game, _ := NewGame(2, wordList)

	// Make 2 wrong guesses
	game.MakeGuess("BRAIN")
	game.MakeGuess("CRANE")

	if game.Status != Lost {
		t.Errorf("After max rounds, Status = %v, want Lost", game.Status)
	}

	if !game.IsGameOver() {
		t.Error("IsGameOver() should return true after loss")
	}
}

func TestInvalidGuess(t *testing.T) {
	wordList := []string{"APPLE"}
	game, _ := NewGame(6, wordList)

	// Test invalid guesses
	invalidGuesses := []string{
		"APP",     // too short
		"APPLES",  // too long
		"APP1E",   // contains number
		"APP-E",   // contains special char
		"",        // empty
		"  ABC  ", // too short after trim
	}

	for _, guess := range invalidGuesses {
		_, err := game.MakeGuess(guess)
		if err == nil {
			t.Errorf("MakeGuess(%q) should return error", guess)
		}
	}
}
