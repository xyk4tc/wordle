package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/admin/wordle/internal/game"
	"github.com/admin/wordle/pkg/api"
)

// RoomStatus represents the status of a room
type RoomStatus string

const (
	RoomWaiting  RoomStatus = "waiting"
	RoomPlaying  RoomStatus = "playing"
	RoomFinished RoomStatus = "finished"
)

// PlayerStatus represents the status of a player
type PlayerStatus string

const (
	PlayerWaiting PlayerStatus = "waiting"
	PlayerPlaying PlayerStatus = "playing"
	PlayerWon     PlayerStatus = "won"
	PlayerLost    PlayerStatus = "lost"
)

// Player represents a player in a room
type Player struct {
	ID         string
	Nickname   string
	Status     PlayerStatus
	Game       *game.Game
	History    []api.GuessResponse
	FinishTime int64 // Unix timestamp when won or lost
	mu         sync.RWMutex
}

// Room represents a multiplayer game room
type Room struct {
	ID          string
	Host        string // Player ID of the host
	Answer      string
	MaxRounds   int
	MaxPlayers  int
	Status      RoomStatus
	Players     map[string]*Player // key: playerID
	PlayerOrder []string           // Maintain join order
	Version     int                // For long polling
	updateCond  *sync.Cond         // Condition variable for broadcasting updates
	mu          sync.RWMutex
}

// RoomManager manages all game rooms
type RoomManager struct {
	rooms     map[string]*Room
	idCounter int
	mu        sync.RWMutex
}

// NewRoomManager creates a new room manager
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

// CreateRoom creates a new game room
func (rm *RoomManager) CreateRoom(playerID, nickname string, maxPlayers, maxRounds int, wordList []string) (*Room, error) {
	// Select a random word for the room
	answer := wordList[game.GetRandomInt(len(wordList))]

	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.idCounter++
	roomID := fmt.Sprintf("%d", rm.idCounter)

	if maxPlayers <= 0 || maxPlayers > 8 {
		maxPlayers = 4
	}

	room := &Room{
		ID:          roomID,
		Host:        playerID,
		Answer:      answer,
		MaxRounds:   maxRounds,
		MaxPlayers:  maxPlayers,
		Status:      RoomWaiting,
		Players:     make(map[string]*Player),
		PlayerOrder: make([]string, 0),
		Version:     0,
	}
	// Initialize condition variable for broadcasting updates
	room.updateCond = sync.NewCond(&room.mu)

	// Add host as first player
	player := &Player{
		ID:       playerID,
		Nickname: nickname,
		Status:   PlayerWaiting,
		History:  make([]api.GuessResponse, 0),
	}
	room.Players[playerID] = player
	room.PlayerOrder = append(room.PlayerOrder, playerID)

	rm.rooms[roomID] = room
	return room, nil
}

// GetRoom gets a room by ID
func (rm *RoomManager) GetRoom(roomID string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	room, exists := rm.rooms[roomID]
	return room, exists
}

// ListRooms lists all available rooms (waiting status)
func (rm *RoomManager) ListRooms() []*Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rooms := make([]*Room, 0)
	for _, room := range rm.rooms {
		if room.Status == RoomWaiting {
			rooms = append(rooms, room)
		}
	}
	return rooms
}

// JoinRoom adds a player to a room
func (r *Room) JoinRoom(playerID, nickname string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != RoomWaiting {
		return fmt.Errorf("room is not accepting new players")
	}

	if len(r.Players) >= r.MaxPlayers {
		return fmt.Errorf("room is full")
	}

	if _, exists := r.Players[playerID]; exists {
		return fmt.Errorf("player already in room")
	}

	player := &Player{
		ID:       playerID,
		Nickname: nickname,
		Status:   PlayerWaiting,
		History:  make([]api.GuessResponse, 0),
	}
	r.Players[playerID] = player
	r.PlayerOrder = append(r.PlayerOrder, playerID)

	r.notifyUpdate()
	return nil
}

// LeaveRoom removes a player from a room
func (r *Room) LeaveRoom(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.Players[playerID]; !exists {
		return fmt.Errorf("player not in room")
	}

	delete(r.Players, playerID)

	// Remove from player order
	for i, id := range r.PlayerOrder {
		if id == playerID {
			r.PlayerOrder = append(r.PlayerOrder[:i], r.PlayerOrder[i+1:]...)
			break
		}
	}

	// If host leaves, assign new host or delete room
	if r.Host == playerID {
		if len(r.Players) > 0 {
			r.Host = r.PlayerOrder[0]
		}
	}

	r.notifyUpdate()
	return nil
}

// StartGame starts the game (only host can start)
func (r *Room) StartGame(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if playerID != r.Host {
		return fmt.Errorf("only host can start the game")
	}

	if r.Status != RoomWaiting {
		return fmt.Errorf("game already started")
	}

	if len(r.Players) < 2 {
		return fmt.Errorf("need at least 2 players to start")
	}

	// Initialize game for each player
	for _, player := range r.Players {
		g, err := game.NewGameWithAnswer(r.MaxRounds, r.Answer)
		if err != nil {
			return err
		}
		player.Game = g
		player.Status = PlayerPlaying
	}

	r.Status = RoomPlaying
	r.notifyUpdate()
	return nil
}

// MakeGuess processes a player's guess
func (r *Room) MakeGuess(playerID, guess string) (*api.GuessResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status != RoomPlaying {
		return nil, fmt.Errorf("game not in progress")
	}

	player, exists := r.Players[playerID]
	if !exists {
		return nil, fmt.Errorf("player not found")
	}

	if player.Status != PlayerPlaying {
		return nil, fmt.Errorf("player already finished")
	}

	// Make the guess
	result, err := player.Game.MakeGuess(guess)
	if err != nil {
		return nil, err
	}

	// Convert to API response
	response := &api.GuessResponse{
		Guess:        result.Guess,
		Results:      convertToAPIResults(result),
		GameOver:     player.Game.IsGameOver(),
		CurrentRound: player.Game.CurrentRound,
		MaxRounds:    player.Game.MaxRounds,
	}

	// Check game status
	switch player.Game.GetStatus() {
	case game.Won:
		response.GameStatus = "won"
		player.Status = PlayerWon
		player.FinishTime = time.Now().Unix()
		r.checkGameEnd()
	case game.Lost:
		response.GameStatus = "lost"
		player.Status = PlayerLost
		player.FinishTime = time.Now().Unix()
		r.checkGameEnd()
	default:
		response.GameStatus = "in_progress"
	}

	player.History = append(player.History, *response)
	r.notifyUpdate()

	return response, nil
}

// checkGameEnd checks if game should end (must be called with lock held)
func (r *Room) checkGameEnd() {
	allFinished := true
	hasWinner := false

	for _, player := range r.Players {
		if player.Status == PlayerPlaying {
			allFinished = false
		}
		if player.Status == PlayerWon {
			hasWinner = true
		}
	}

	// End game if: (1) someone won, or (2) all players finished
	if hasWinner || allFinished {
		r.Status = RoomFinished
	}
}

// GetProgress returns the current progress of all players
func (r *Room) GetProgress() *api.RoomProgressResponse {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]api.PlayerProgress, 0, len(r.Players))
	for _, playerID := range r.PlayerOrder {
		player := r.Players[playerID]

		var lastGuess *api.GuessResponse
		if len(player.History) > 0 {
			lastGuess = &player.History[len(player.History)-1]
		}

		currentRound := 0
		if player.Game != nil {
			currentRound = player.Game.CurrentRound
		}

		players = append(players, api.PlayerProgress{
			PlayerID:     player.ID,
			Nickname:     player.Nickname,
			CurrentRound: currentRound,
			MaxRounds:    r.MaxRounds,
			Status:       string(player.Status),
			LastGuess:    lastGuess,
			History:      player.History,
			FinishTime:   player.FinishTime,
		})
	}

	response := &api.RoomProgressResponse{
		RoomID:    r.ID,
		Status:    string(r.Status),
		Players:   players,
		Version:   r.Version,
		Timestamp: time.Now().Unix(),
	}

	if r.Status == RoomFinished {
		response.Answer = r.Answer
		response.Winner, response.Ranking = r.calculateRanking()
	}

	return response
}

// calculateRanking calculates the final ranking (must be called with lock held)
func (r *Room) calculateRanking() (winner string, ranking []string) {
	// Sort players by: 1. Won > Lost, 2. Fewer rounds, 3. Earlier finish time
	type playerRank struct {
		playerID   string
		won        bool
		rounds     int
		finishTime int64
	}

	ranks := make([]playerRank, 0, len(r.Players))
	for _, player := range r.Players {
		rank := playerRank{
			playerID:   player.ID,
			won:        player.Status == PlayerWon,
			rounds:     player.Game.CurrentRound,
			finishTime: player.FinishTime,
		}
		ranks = append(ranks, rank)
	}

	// Bubble sort (simple for small number of players)
	for i := 0; i < len(ranks); i++ {
		for j := i + 1; j < len(ranks); j++ {
			// Compare: won first, then rounds, then time
			swap := false
			if ranks[i].won != ranks[j].won {
				swap = !ranks[i].won // Won is better
			} else if ranks[i].won { // Both won
				if ranks[i].rounds != ranks[j].rounds {
					swap = ranks[i].rounds > ranks[j].rounds // Fewer rounds is better
				} else {
					swap = ranks[i].finishTime > ranks[j].finishTime // Earlier is better
				}
			} else { // Both lost
				if ranks[i].rounds != ranks[j].rounds {
					swap = ranks[i].rounds > ranks[j].rounds // More attempts is better
				} else {
					swap = ranks[i].finishTime > ranks[j].finishTime // Earlier is better
				}
			}

			if swap {
				ranks[i], ranks[j] = ranks[j], ranks[i]
			}
		}
	}

	ranking = make([]string, len(ranks))
	for i, rank := range ranks {
		ranking[i] = rank.playerID
	}

	if len(ranking) > 0 {
		winner = ranking[0]
	}

	return winner, ranking
}

// notifyUpdate increments version and broadcasts to all waiting clients
// Must be called with write lock held
func (r *Room) notifyUpdate() {
	r.Version++
	// Broadcast wakes up all goroutines waiting on the condition variable
	r.updateCond.Broadcast()
}

// GetStatus returns the room status
func (r *Room) GetStatus() *api.RoomStatusResponse {
	r.mu.RLock()
	defer r.mu.RUnlock()

	playerNames := make([]string, 0, len(r.Players))
	for _, playerID := range r.PlayerOrder {
		playerNames = append(playerNames, r.Players[playerID].Nickname)
	}

	return &api.RoomStatusResponse{
		RoomID:      r.ID,
		Status:      string(r.Status),
		PlayerCount: len(r.Players),
		MaxPlayers:  r.MaxPlayers,
		MaxRounds:   r.MaxRounds,
		Players:     playerNames,
		Host:        r.Host,
	}
}
