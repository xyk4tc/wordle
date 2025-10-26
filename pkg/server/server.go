package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/admin/wordle/internal/config"
	"github.com/admin/wordle/internal/game"
	"github.com/admin/wordle/pkg/api"
	"github.com/gin-gonic/gin"
)

// Server represents the Wordle game server
type Server struct {
	sessions    map[string]*GameSession
	roomManager *RoomManager
	config      *config.Config
	mu          sync.RWMutex
	idCounter   int
}

// NewServer creates a new game server
func NewServer(cfg *config.Config) *Server {
	return &Server{
		sessions:    make(map[string]*GameSession),
		roomManager: NewRoomManager(),
		config:      cfg,
	}
}

// HandleNewGame handles the creation of a new game
func (s *Server) HandleNewGame(c *gin.Context) {
	// Server uses its own configuration only
	// Create new game with server config
	g, err := game.NewGame(s.config.MaxRounds, s.config.WordList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: fmt.Sprintf("Failed to create game: %v", err),
		})
		return
	}

	// Generate game ID and create session
	s.mu.Lock()
	s.idCounter++
	gameID := strconv.Itoa(s.idCounter)
	session := NewGameSession(gameID, g)
	s.sessions[gameID] = session
	s.mu.Unlock()

	response := api.NewGameResponse{
		GameID:    gameID,
		MaxRounds: s.config.MaxRounds,
		Message:   "Game created successfully",
	}

	c.JSON(http.StatusCreated, response)
}

// HandleGuess handles a guess submission
func (s *Server) HandleGuess(c *gin.Context) {
	// Extract game ID from URL path parameter
	gameID := c.Param("id")

	s.mu.RLock()
	session, exists := s.sessions[gameID]
	s.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: "Game not found",
		})
		return
	}

	var req api.GuessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	// Validate input
	if !game.ValidateWord(req.Guess) {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Invalid word: must be 5 letters, alphabetic only",
		})
		return
	}

	response, err := session.MakeGuess(req.Guess)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HandleStatus handles game status requests
func (s *Server) HandleStatus(c *gin.Context) {
	// Extract game ID from URL path parameter
	gameID := c.Param("id")

	s.mu.RLock()
	session, exists := s.sessions[gameID]
	s.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: "Game not found",
		})
		return
	}

	status := session.GetStatus()
	c.JSON(http.StatusOK, status)
}

// ============================================
// Multi-player Room API Handlers (Task 4)
// ============================================

// HandleCreateRoom handles room creation
func (s *Server) HandleCreateRoom(c *gin.Context) {
	var req api.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	if req.Nickname == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Nickname is required",
		})
		return
	}

	// Generate player ID
	s.mu.Lock()
	s.idCounter++
	playerID := fmt.Sprintf("player-%d", s.idCounter)
	s.mu.Unlock()

	maxPlayers := req.MaxPlayers
	if maxPlayers == 0 {
		maxPlayers = 4
	}

	room, err := s.roomManager.CreateRoom(playerID, req.Nickname, maxPlayers, s.config.MaxRounds, s.config.WordList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: fmt.Sprintf("Failed to create room: %v", err),
		})
		return
	}

	response := api.CreateRoomResponse{
		RoomID:    room.ID,
		MaxRounds: room.MaxRounds,
		Message:   fmt.Sprintf("Room created! You are the host. Player ID: %s", playerID),
	}

	c.JSON(http.StatusCreated, response)
}

// HandleJoinRoom handles joining a room
func (s *Server) HandleJoinRoom(c *gin.Context) {
	roomID := c.Param("id")

	var req api.JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	if req.Nickname == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Nickname is required",
		})
		return
	}

	room, exists := s.roomManager.GetRoom(roomID)
	if !exists {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: "Room not found",
		})
		return
	}

	// Generate player ID
	s.mu.Lock()
	s.idCounter++
	playerID := fmt.Sprintf("player-%d", s.idCounter)
	s.mu.Unlock()

	err := room.JoinRoom(playerID, req.Nickname)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Get player list
	status := room.GetStatus()

	response := api.JoinRoomResponse{
		RoomID:    roomID,
		MaxRounds: room.MaxRounds,
		Players:   status.Players,
		IsHost:    playerID == room.Host,
		Message:   fmt.Sprintf("Joined room successfully! Player ID: %s", playerID),
	}

	c.JSON(http.StatusOK, response)
}

// HandleLeaveRoom handles leaving a room
func (s *Server) HandleLeaveRoom(c *gin.Context) {
	roomID := c.Param("id")
	playerID := c.Query("player_id")

	if playerID == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Player ID is required",
		})
		return
	}

	room, exists := s.roomManager.GetRoom(roomID)
	if !exists {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: "Room not found",
		})
		return
	}

	err := room.LeaveRoom(playerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Left room successfully",
	})
}

// HandleStartRoom handles starting the game
func (s *Server) HandleStartRoom(c *gin.Context) {
	roomID := c.Param("id")
	playerID := c.Query("player_id")

	if playerID == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Player ID is required",
		})
		return
	}

	room, exists := s.roomManager.GetRoom(roomID)
	if !exists {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: "Room not found",
		})
		return
	}

	err := room.StartGame(playerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Game started!",
	})
}

// HandleRoomGuess handles a guess in multiplayer mode
func (s *Server) HandleRoomGuess(c *gin.Context) {
	roomID := c.Param("id")

	var req api.RoomGuessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Invalid request body",
		})
		return
	}

	// Validate input
	if !game.ValidateWord(req.Guess) {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "Invalid word: must be 5 letters, alphabetic only",
		})
		return
	}

	room, exists := s.roomManager.GetRoom(roomID)
	if !exists {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: "Room not found",
		})
		return
	}

	response, err := room.MakeGuess(req.PlayerID, req.Guess)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HandleRoomProgress handles long polling for room progress
func (s *Server) HandleRoomProgress(c *gin.Context) {
	roomID := c.Param("id")
	versionStr := c.Query("version")

	room, exists := s.roomManager.GetRoom(roomID)
	if !exists {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: "Room not found",
		})
		return
	}

	// Parse version
	lastVersion := 0
	if versionStr != "" {
		if v, err := strconv.Atoi(versionStr); err == nil {
			lastVersion = v
		}
	}

	// Long polling implementation using sync.Cond
	// Check if there's already an update
	room.mu.RLock()
	currentVersion := room.Version
	room.mu.RUnlock()

	if currentVersion > lastVersion {
		progress := room.GetProgress()
		c.JSON(http.StatusOK, progress)
		return
	}

	// Wait for update or timeout using condition variable
	// We use a separate goroutine to wait on the condition variable so we can
	// simultaneously listen for timeout and client disconnect using select
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	done := make(chan struct{})

	go func() {
		room.mu.Lock()
		defer room.mu.Unlock()
		defer close(done) // Always close done to prevent blocking

		// Wait for version change or context cancellation
		// Note: Wait() releases the lock while waiting and reacquires it on wakeup
		for room.Version == lastVersion {
			// Check if context is cancelled before waiting
			select {
			case <-ctx.Done():
				return
			default:
			}

			room.updateCond.Wait() // Releases lock, waits, then reacquires lock
		}
	}()

	// Wait for version change or timeout
	select {
	case <-done:
		// Version changed - return new progress
		progress := room.GetProgress()
		c.JSON(http.StatusOK, progress)

	case <-ctx.Done():
		// Timeout or client disconnected - wake up the waiting goroutine
		room.updateCond.Broadcast()
		<-done // Wait for goroutine to exit cleanly

		// Check if it's a timeout or client disconnect
		if c.Request.Context().Err() != nil {
			// Client disconnected
			return
		}
		// Timeout - return current state
		progress := room.GetProgress()
		c.JSON(http.StatusOK, progress)
	}
}

// HandleRoomStatus handles room status requests
func (s *Server) HandleRoomStatus(c *gin.Context) {
	roomID := c.Param("id")

	room, exists := s.roomManager.GetRoom(roomID)
	if !exists {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: "Room not found",
		})
		return
	}

	status := room.GetStatus()
	c.JSON(http.StatusOK, status)
}

// HandleListRooms handles listing all available rooms
func (s *Server) HandleListRooms(c *gin.Context) {
	rooms := s.roomManager.ListRooms()

	roomList := make([]api.RoomStatusResponse, 0, len(rooms))
	for _, room := range rooms {
		status := room.GetStatus()
		roomList = append(roomList, *status)
	}

	response := api.ListRoomsResponse{
		Rooms: roomList,
	}

	c.JSON(http.StatusOK, response)
}
