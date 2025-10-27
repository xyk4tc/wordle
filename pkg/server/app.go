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
	// Set gin to debug mode to see more details
	gin.SetMode(gin.DebugMode)

	// Create router with logger and recovery middleware
	router := gin.New()
	router.Use(gin.Logger())   // Add logger middleware to print requests
	router.Use(gin.Recovery()) // Add recovery middleware to handle panics

	return &App{
		server: NewServer(cfg),
		router: router,
		port:   port,
	}
}

// Start starts the HTTP server
func (a *App) Start() error {
	// Register single-player game routes (Task 2)
	a.router.POST("/game/new", a.server.HandleNewGame)
	a.router.POST("/game/:id/guess", a.server.HandleGuess)
	a.router.GET("/game/:id/status", a.server.HandleStatus)

	// Register multi-player room routes (Task 4)
	a.router.POST("/room/create", a.server.HandleCreateRoom)
	a.router.POST("/room/:id/join", a.server.HandleJoinRoom)
	a.router.POST("/room/:id/leave", a.server.HandleLeaveRoom)
	a.router.POST("/room/:id/start", a.server.HandleStartRoom)
	a.router.POST("/room/:id/guess", a.server.HandleRoomGuess)
	a.router.GET("/room/:id/progress", a.server.HandleRoomProgress)
	a.router.GET("/room/:id/status", a.server.HandleRoomStatus)
	a.router.GET("/room/list", a.server.HandleListRooms)

	// Print startup info
	addr := ":" + a.port
	fmt.Printf("Wordle Server starting on http://localhost%s\n", addr)
	fmt.Println("\n=== Single-Player API (Task 2) ===")
	fmt.Println("  POST /game/new            - Create new game")
	fmt.Println("  POST /game/:id/guess      - Submit a guess")
	fmt.Println("  GET  /game/:id/status     - Get game status")
	fmt.Println("\n=== Multi-Player API (Task 4) ===")
	fmt.Println("  POST /room/create         - Create a room")
	fmt.Println("  POST /room/:id/join       - Join a room")
	fmt.Println("  POST /room/:id/leave      - Leave a room")
	fmt.Println("  POST /room/:id/start      - Start the game (host only)")
	fmt.Println("  POST /room/:id/guess      - Submit a guess")
	fmt.Println("  GET  /room/:id/progress   - Get live progress (long polling)")
	fmt.Println("  GET  /room/:id/status     - Get room status")
	fmt.Println("  GET  /room/list           - List available rooms")
	fmt.Println()

	// Start server
	log.Printf("Server listening on port %s", a.port)
	return a.router.Run(addr)
}
