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

// ============================================
// Multi-player Room API (Task 4)
// ============================================

// CreateRoomRequest represents a request to create a multiplayer room
type CreateRoomRequest struct {
	Nickname   string `json:"nickname"`
	MaxPlayers int    `json:"max_players,omitempty"` // Default: 4
}

// CreateRoomResponse represents the response when creating a room
type CreateRoomResponse struct {
	RoomID    string `json:"room_id"`
	MaxRounds int    `json:"max_rounds"`
	Message   string `json:"message"`
}

// JoinRoomRequest represents a request to join a room
type JoinRoomRequest struct {
	Nickname string `json:"nickname"`
}

// JoinRoomResponse represents the response when joining a room
type JoinRoomResponse struct {
	RoomID    string   `json:"room_id"`
	MaxRounds int      `json:"max_rounds"`
	Players   []string `json:"players"` // List of player nicknames
	IsHost    bool     `json:"is_host"`
	Message   string   `json:"message"`
}

// RoomGuessRequest represents a guess in multiplayer mode
type RoomGuessRequest struct {
	PlayerID string `json:"player_id"`
	Guess    string `json:"guess"`
}

// PlayerProgress represents a player's progress in the room
type PlayerProgress struct {
	PlayerID     string          `json:"player_id"`
	Nickname     string          `json:"nickname"`
	CurrentRound int             `json:"current_round"`
	MaxRounds    int             `json:"max_rounds"`
	Status       string          `json:"status"` // "waiting", "playing", "won", "lost"
	LastGuess    *GuessResponse  `json:"last_guess,omitempty"`
	History      []GuessResponse `json:"history"`
	FinishTime   int64           `json:"finish_time,omitempty"` // Unix timestamp when finished
}

// RoomProgressResponse represents the progress of all players in a room
type RoomProgressResponse struct {
	RoomID    string           `json:"room_id"`
	Status    string           `json:"status"` // "waiting", "playing", "finished"
	Players   []PlayerProgress `json:"players"`
	Winner    string           `json:"winner,omitempty"`  // PlayerID of winner
	Ranking   []string         `json:"ranking,omitempty"` // Sorted PlayerIDs by rank
	Answer    string           `json:"answer,omitempty"`  // Only when game finished
	Version   int              `json:"version"`           // For long polling
	Timestamp int64            `json:"timestamp"`         // Unix timestamp
}

// RoomStatusResponse represents the current room status
type RoomStatusResponse struct {
	RoomID      string   `json:"room_id"`
	Status      string   `json:"status"` // "waiting", "playing", "finished"
	PlayerCount int      `json:"player_count"`
	MaxPlayers  int      `json:"max_players"`
	MaxRounds   int      `json:"max_rounds"`
	Players     []string `json:"players"` // List of player nicknames
	Host        string   `json:"host"`    // Host player ID
}

// ListRoomsResponse represents the list of available rooms
type ListRoomsResponse struct {
	Rooms []RoomStatusResponse `json:"rooms"`
}
