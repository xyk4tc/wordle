package server

import (
	"sync"

	"github.com/admin/wordle/internal/game"
	"github.com/admin/wordle/pkg/api"
)

// GameSession represents a server-side game session
type GameSession struct {
	ID      string
	Game    *game.Game
	History []api.GuessResponse
	mu      sync.RWMutex
}

// NewGameSession creates a new game session
func NewGameSession(id string, g *game.Game) *GameSession {
	return &GameSession{
		ID:      id,
		Game:    g,
		History: []api.GuessResponse{},
	}
}

// MakeGuess processes a guess and returns the result
func (s *GameSession) MakeGuess(guess string) (*api.GuessResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.Game.MakeGuess(guess)
	if err != nil {
		return nil, err
	}

	// Convert game.GuessResult to api.GuessResponse
	response := &api.GuessResponse{
		Guess:        result.Guess,
		Results:      convertToAPIResults(result),
		GameOver:     s.Game.IsGameOver(),
		CurrentRound: s.Game.CurrentRound,
		MaxRounds:    s.Game.MaxRounds,
	}

	// Set game status
	switch s.Game.GetStatus() {
	case game.Won:
		response.GameStatus = "won"
		response.Answer = s.Game.Answer
		response.Message = "Congratulations! You won!"
	case game.Lost:
		response.GameStatus = "lost"
		response.Answer = s.Game.Answer
		response.Message = "Game over! Better luck next time."
	default:
		response.GameStatus = "in_progress"
	}

	s.History = append(s.History, *response)
	return response, nil
}

// GetStatus returns the current game status
func (s *GameSession) GetStatus() *api.GameStatusResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := &api.GameStatusResponse{
		GameID:       s.ID,
		CurrentRound: s.Game.CurrentRound,
		MaxRounds:    s.Game.MaxRounds,
		History:      s.History,
	}

	switch s.Game.GetStatus() {
	case game.Won:
		status.GameStatus = "won"
		status.Answer = s.Game.Answer
	case game.Lost:
		status.GameStatus = "lost"
		status.Answer = s.Game.Answer
	default:
		status.GameStatus = "in_progress"
	}

	return status
}

// convertToAPIResults converts game letter statuses to API format
func convertToAPIResults(result game.GuessResult) []string {
	results := make([]string, len(result.Statuses))
	for i, status := range result.Statuses {
		switch status {
		case game.Hit:
			results[i] = "O"
		case game.Present:
			results[i] = "?"
		case game.Miss:
			results[i] = "_"
		default:
			results[i] = "?"
		}
	}
	return results
}
