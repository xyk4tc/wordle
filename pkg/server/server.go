package server

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/admin/wordle/internal/config"
	"github.com/admin/wordle/internal/game"
	"github.com/admin/wordle/pkg/api"
	"github.com/gin-gonic/gin"
)

// Server represents the Wordle game server
type Server struct {
	sessions  map[string]*GameSession
	config    *config.Config
	mu        sync.RWMutex
	idCounter int
}

// NewServer creates a new game server
func NewServer(cfg *config.Config) *Server {
	return &Server{
		sessions: make(map[string]*GameSession),
		config:   cfg,
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
