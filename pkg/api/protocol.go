package api

// NewGameRequest represents a request to create a new game
type NewGameRequest struct {
	// No parameters - server uses its own configuration
}

// NewGameResponse represents the response when creating a new game
type NewGameResponse struct {
	GameID    string `json:"game_id"`
	MaxRounds int    `json:"max_rounds"`
	Message   string `json:"message"`
}

// GuessRequest represents a guess submission
type GuessRequest struct {
	Guess string `json:"guess"`
}

// GuessResponse represents the response to a guess
type GuessResponse struct {
	Guess        string   `json:"guess"`
	Results      []string `json:"results"` // Array of "O" (hit), "?" (present), "_" (miss)
	GameOver     bool     `json:"game_over"`
	GameStatus   string   `json:"game_status"` // "in_progress", "won", "lost"
	CurrentRound int      `json:"current_round"`
	MaxRounds    int      `json:"max_rounds"`
	Answer       string   `json:"answer,omitempty"` // Only present when game is over
	Message      string   `json:"message,omitempty"`
}

// GameStatusResponse represents the current game status
type GameStatusResponse struct {
	GameID       string          `json:"game_id"`
	CurrentRound int             `json:"current_round"`
	MaxRounds    int             `json:"max_rounds"`
	GameStatus   string          `json:"game_status"`
	History      []GuessResponse `json:"history"`
	Answer       string          `json:"answer,omitempty"` // Only present when game is over
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
