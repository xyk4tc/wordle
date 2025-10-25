package server

import (
	"fmt"
	"log"

	"github.com/admin/wordle/internal/config"
	"github.com/gin-gonic/gin"
)

// App represents the server application
type App struct {
	server *Server
	router *gin.Engine
	port   string
}

// NewApp creates a new server application
func NewApp(cfg *config.Config, port string) *App {
	// Set gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())

	return &App{
		server: NewServer(cfg),
		router: router,
		port:   port,
	}
}

// Start starts the HTTP server
func (a *App) Start() error {
	// Register routes
	a.router.POST("/game/new", a.server.HandleNewGame)
	a.router.POST("/game/:id/guess", a.server.HandleGuess)
	a.router.GET("/game/:id/status", a.server.HandleStatus)

	// Print startup info
	addr := ":" + a.port
	fmt.Printf("Wordle Server starting on http://localhost%s\n", addr)
	fmt.Println("API Endpoints:")
	fmt.Println("  POST /game/new          - Create new game")
	fmt.Println("  POST /game/:id/guess    - Submit a guess")
	fmt.Println("  GET  /game/:id/status   - Get game status")
	fmt.Println()

	// Start server
	log.Printf("Server listening on port %s", a.port)
	return a.router.Run(addr)
}
