package game

import (
	"errors"
	"math/rand"
	"strings"
)

// GameStatus represents the current status of the game
type GameStatus int

const (
	// InProgress means the game is still ongoing
	InProgress GameStatus = iota
	// Won means the player has won the game
	Won
	// Lost means the player has lost the game
	Lost
)

// Game represents a Wordle game instance
type Game struct {
	Answer       string
	MaxRounds    int
	WordList     []string
	CurrentRound int
	History      []GuessResult
	Status       GameStatus
}

// NewGame creates a new Wordle game with the given configuration
func NewGame(maxRounds int, wordList []string) (*Game, error) {
	if maxRounds <= 0 {
		return nil, errors.New("max rounds must be positive")
	}
	if len(wordList) == 0 {
		return nil, errors.New("word list cannot be empty")
	}

	// Validate all words in the list
	validWords := []string{}
	for _, word := range wordList {
		word = strings.TrimSpace(word)
		if ValidateWord(word) {
			validWords = append(validWords, strings.ToUpper(word))
		}
	}

	if len(validWords) == 0 {
		return nil, errors.New("no valid words in word list")
	}

	// Select a random word as the answer
	answer := validWords[rand.Intn(len(validWords))]

	return &Game{
		Answer:       answer,
		MaxRounds:    maxRounds,
		WordList:     validWords,
		CurrentRound: 0,
		History:      []GuessResult{},
		Status:       InProgress,
	}, nil
}

// MakeGuess processes a player's guess and updates the game state
func (g *Game) MakeGuess(guess string) (GuessResult, error) {
	if g.Status != InProgress {
		return GuessResult{}, errors.New("game is already over")
	}

	guess = strings.TrimSpace(guess)
	if !ValidateWord(guess) {
		return GuessResult{}, errors.New("invalid word: must be 5 letters, alphabetic only")
	}

	guess = strings.ToUpper(guess)

	// Optional: Check if the guess is in the word list
	// For now, we'll allow any valid 5-letter word

	g.CurrentRound++
	result := EvaluateGuess(guess, g.Answer)
	g.History = append(g.History, result)

	// Check if the player won
	if guess == g.Answer {
		g.Status = Won
		return result, nil
	}

	// Check if the player lost
	if g.CurrentRound >= g.MaxRounds {
		g.Status = Lost
		return result, nil
	}

	return result, nil
}

// IsGameOver checks if the game has ended
func (g *Game) IsGameOver() bool {
	return g.Status != InProgress
}

// GetStatus returns the current game status
func (g *Game) GetStatus() GameStatus {
	return g.Status
}

// GetRemainingRounds returns the number of remaining rounds
func (g *Game) GetRemainingRounds() int {
	return g.MaxRounds - g.CurrentRound
}
