package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/admin/wordle/pkg/api"
	"golang.org/x/sync/errgroup"
)

// RoomApp represents the multiplayer client application
type RoomApp struct {
	client          *RoomClient
	reader          *bufio.Reader
	screen          *ScreenManager
	progressVersion int
	gameStarted     bool
	gameFinished    bool
	isHost          bool
	currentProgress *api.RoomProgressResponse
	mu              sync.RWMutex
	stopProgress    chan struct{}
	// Global input channel - all input reads go through here
	inputChan chan string
	// Game finished notification channel
	gameFinishedChan chan struct{}
}

// NewRoomApp creates a new multiplayer app
func NewRoomApp(serverURL string, input io.Reader) *RoomApp {
	app := &RoomApp{
		client:           NewRoomClient(serverURL),
		reader:           bufio.NewReader(input),
		screen:           NewScreenManager(),
		stopProgress:     make(chan struct{}),
		inputChan:        make(chan string, 1),
		gameFinishedChan: make(chan struct{}, 1),
	}

	// Start global input reading goroutine
	// This goroutine runs for the lifetime of the app
	// All input operations read from inputChan instead of directly from reader
	go func() {
		for {
			input, err := app.reader.ReadString('\n')
			if err != nil {
				// EOF or error, close the channel
				close(app.inputChan)
				return
			}
			input = strings.TrimSpace(input)

			// Send to channel (blocking - this is intentional)
			// Only one place reads at a time, so no contention
			app.inputChan <- input
		}
	}()

	return app
}

// Run starts the multiplayer application
func (a *RoomApp) Run() error {
	fmt.Println("\n=== Multi-Player Wordle ===")
	fmt.Println()

	// Show main menu
	for {
		fmt.Println("Choose an option:")
		fmt.Println("  1. Create new room")
		fmt.Println("  2. Join existing room")
		fmt.Println("  3. List available rooms")
		fmt.Println("  4. Quit")
		fmt.Print("\nEnter choice: ")

		choice := <-a.inputChan

		switch choice {
		case "1":
			if err := a.createRoomFlow(); err != nil {
				return err
			}
			return nil
		case "2":
			if err := a.joinRoomFlow(); err != nil {
				return err
			}
			return nil
		case "3":
			a.listRooms()
		case "4", "quit", "exit":
			fmt.Println("Goodbye!")
			return nil
		default:
			fmt.Println("Invalid choice. Please try again.\n")
		}
	}
}

// createRoomFlow handles creating a new room
func (a *RoomApp) createRoomFlow() error {
	fmt.Print("\nEnter your nickname: ")
	nickname := <-a.inputChan
	if nickname == "" {
		nickname = "Player"
	}

	fmt.Print("Max players (2-8, default 4): ")
	maxPlayersStr := <-a.inputChan
	maxPlayers := 4
	if maxPlayersStr != "" {
		fmt.Sscanf(maxPlayersStr, "%d", &maxPlayers)
	}

	// Create room
	fmt.Println("\nCreating room...")
	resp, err := a.client.CreateRoom(nickname, maxPlayers)
	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}

	fmt.Printf("\n✓ Room created! Room ID: %s\n", resp.RoomID)
	fmt.Printf("You are the host. Waiting for players to join...\n")
	fmt.Printf("Share this room ID with your friends: %s\n\n", resp.RoomID)

	a.isHost = true
	return a.roomLobby()
}

// joinRoomFlow handles joining an existing room
func (a *RoomApp) joinRoomFlow() error {
	var roomID string
	var nickname string

	// Loop until valid room ID is provided or user quits
	for {
		fmt.Print("\nEnter room ID (or 'list' to see available rooms, 'quit' to cancel): ")
		roomID = <-a.inputChan

		if roomID == "" {
			fmt.Println("❌ Room ID cannot be empty")
			continue
		}

		if roomID == "quit" || roomID == "exit" {
			fmt.Println("Cancelled.")
			return nil
		}

		if roomID == "list" || roomID == "ls" {
			a.listRooms()
			continue
		}

		// Validate room ID exists
		if !a.validateRoomExists(roomID) {
			fmt.Printf("\n❌ Room '%s' does not exist!\n", roomID)
			a.listRooms()
			continue
		}

		// Room ID is valid, break the loop
		break
	}

	fmt.Print("Enter your nickname: ")
	nickname = <-a.inputChan
	if nickname == "" {
		nickname = "Player"
	}

	// Join room
	fmt.Println("\nJoining room...")
	resp, err := a.client.JoinRoom(roomID, nickname)
	if err != nil {
		return fmt.Errorf("failed to join room: %w", err)
	}

	fmt.Printf("\n✓ Joined room %s!\n", resp.RoomID)
	fmt.Printf("Players in room: %s\n\n", strings.Join(resp.Players, ", "))

	a.isHost = resp.IsHost
	return a.roomLobby()
}

// validateRoomExists checks if a room with the given ID exists
func (a *RoomApp) validateRoomExists(roomID string) bool {
	resp, err := a.client.ListRooms()
	if err != nil {
		fmt.Printf("Warning: Could not verify room existence: %v\n", err)
		// On error, let the user try anyway (server will validate)
		return true
	}

	for _, room := range resp.Rooms {
		if room.RoomID == roomID {
			return true
		}
	}

	return false
}

// roomLobby handles the waiting room before game starts
func (a *RoomApp) roomLobby() error {
	// Create context for lifecycle management
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create errgroup for status monitoring
	g, ctx := errgroup.WithContext(ctx)

	// Status channel (buffered to prevent blocking)
	statusChan := make(chan *api.RoomStatusResponse, 1)

	// Status monitoring goroutine
	g.Go(func() error {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		// Helper function to send status or handle context cancellation
		sendStatus := func(status *api.RoomStatusResponse) error {
			select {
			case statusChan <- status:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Fetch initial status immediately
		status, err := a.client.GetRoomStatus()
		if err != nil {
			return fmt.Errorf("failed to get room status: %w", err)
		}
		if err := sendStatus(status); err != nil {
			return err
		}

		// Poll periodically
		for {
			select {
			case <-ctx.Done():
				return nil // Clean exit, not an error
			case <-ticker.C:
				status, err := a.client.GetRoomStatus()
				if err != nil {
					return fmt.Errorf("failed to get room status: %w", err)
				}
				if err := sendStatus(status); err != nil {
					return err
				}
			}
		}
	})

	// Show lobby interface
	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║          Waiting Room                 ║")
	fmt.Println("╚═══════════════════════════════════════╝")
	fmt.Println()

	// Track last status update time to throttle UI updates
	lastStatusUpdate := time.Now()

	// Define input prompts based on user role
	var inputPrompt string
	if a.isHost {
		inputPrompt = "⌨️  [Host] Type 's' to start or 'quit' to leave: "
	} else {
		inputPrompt = "⌨️  Type 'quit' to leave (waiting for host to start): "
	}

	// Initial status display
	fmt.Println("📊 Players: (loading...)")
	fmt.Print(inputPrompt)

	// Main event loop
	for {
		select {
		case status := <-statusChan:
			// Check if game has started
			if status.Status == "playing" {
				// Clear lines and show game starting message
				ansiMoveCursorUp := fmt.Sprintf(AnsiCursorUp, 1)
				ansiClearLine := "\r" + AnsiClearLineRight
				output := ansiMoveCursorUp + ansiClearLine + "\n🎮 Game is starting!\n"
				fmt.Print(output)

				// Cancel context and wait for status goroutine to exit
				cancel()
				_ = g.Wait()
				return a.playGame()
			}

			// Throttle UI updates to avoid flickering
			if time.Since(lastStatusUpdate) > 500*time.Millisecond {
				lastStatusUpdate = time.Now()

				// Update player list without disrupting input line
				ansiMoveCursorUp := fmt.Sprintf(AnsiCursorUp, 1)
				ansiClearLine := "\r" + AnsiClearLine

				playerList := strings.Join(status.Players, ", ")
				playerStatusLine := fmt.Sprintf("📊 Players (%d/%d): %s", status.PlayerCount, status.MaxPlayers, playerList)

				// Move up, clear line, print new status, move down, reprint prompt
				output := ansiMoveCursorUp + ansiClearLine + playerStatusLine + "\n" + inputPrompt
				fmt.Print(output)
			}

		case input := <-a.inputChan:
			// Handle user quit command
			if input == "quit" || input == "exit" {
				fmt.Println("\nLeaving room...")
				cancel()
				_ = g.Wait()
				return nil
			}

			// Handle host start game command
			if a.isHost && (input == "start" || input == "s") {
				if err := a.client.StartGame(); err != nil {
					fmt.Printf("\n❌ Error starting game: %v\n", err)
					fmt.Print(inputPrompt)
				} else {
					fmt.Println("\n🚀 Starting game...")
				}
			} else if a.isHost && input != "" {
				// Invalid input for host
				fmt.Println("\n💡 Hint: Type 's' or 'start' to begin")
				fmt.Print(inputPrompt)
			} else if !a.isHost && input != "" {
				// Invalid input for non-host
				fmt.Println("\n💡 Only the host can start the game")
				fmt.Print(inputPrompt)
			}
			// Empty input - just ignore

		case <-ctx.Done():
			// Context cancelled due to error from status goroutine
			if err := g.Wait(); err != nil && err != context.Canceled {
				return fmt.Errorf("background error: %w", err)
			}
			return nil
		}
	}
}

// playGame handles the main game loop with split-screen UI
func (a *RoomApp) playGame() error {
	a.gameStarted = true

	// Get initial progress to determine player count
	progress, err := a.client.GetProgress(0)
	if err != nil {
		return err
	}

	a.mu.Lock()
	a.progressVersion = progress.Version
	a.currentProgress = progress
	a.mu.Unlock()

	// Initialize screen with dynamic layout based on player count
	a.screen.InitScreen(len(progress.Players))
	defer a.screen.CleanupScreen()

	// Initial progress display
	a.screen.UpdateProgress(progress)

	// Add initial log messages (only once)
	myProgress := a.findMyProgress(progress)
	a.screen.AddLogLine("--- Game Started ---")
	a.screen.AddLogLine(fmt.Sprintf("Room: %s | Max Rounds: %d", a.client.GetRoomID(), myProgress.MaxRounds))
	a.screen.AddLogLine("O=Hit | ?=Present | _=Miss")
	a.screen.AddLogLine("Type QUIT to exit")

	// Start progress monitoring in background (non-blocking)
	go a.monitorProgress()

	// Main game loop - handles user input and monitors game end
gameLoop:
	for {
		a.mu.RLock()
		currentProgress := a.currentProgress
		myProgress := a.findMyProgress(currentProgress)
		a.mu.RUnlock()

		// Check if player finished
		if myProgress.Status == "won" || myProgress.Status == "lost" {
			a.screen.AddLogLine("You finished! Waiting for others...")

			// Just wait for game to finish, no input needed
			select {
			case <-a.gameFinishedChan:
				break gameLoop
			case <-time.After(2 * time.Second):
				// Check again after timeout
				continue
			}
		}

		// Player still playing, prompt for input
		a.screen.PromptInput(myProgress.CurrentRound+1, myProgress.MaxRounds)

		// Wait for either user input or game finished
		select {
		case <-a.gameFinishedChan:
			// Game ended while waiting for input
			a.screen.ClearInputLine()
			a.screen.AddLogLine("Game finished!")
			break gameLoop

		case guess := <-a.inputChan:
			guess = strings.ToUpper(guess)
			a.screen.ClearInputLine()

			if guess == "QUIT" || guess == "EXIT" {
				a.screen.AddLogLine("Exiting game...")
				break gameLoop
			}

			// Submit guess
			response, err := a.client.MakeGuess(guess)
			if err != nil {
				a.screen.AddLogLine(fmt.Sprintf("Error: %v", err))
				continue
			}

			// Display result in log
			result := strings.Join(response.Results, "")
			a.screen.AddLogLine(fmt.Sprintf("You: %s (%s)", result, response.Guess))

			if response.GameStatus == "won" {
				a.screen.AddLogLine("🎉 You got it!")
			} else if response.GameStatus == "lost" {
				a.screen.AddLogLine("😔 Out of guesses!")
			}
		}
	}

	// Stop progress monitoring
	close(a.stopProgress)

	// Wait a bit for final updates
	time.Sleep(500 * time.Millisecond)

	// Show final results
	a.mu.RLock()
	finalProgress := a.currentProgress
	a.mu.RUnlock()

	if finalProgress != nil {
		a.showFinalResults(finalProgress)
	}

	// Wait for user to press any key to continue
	a.screen.AddLogLine("")
	a.screen.AddLogLine("Press ENTER to return to menu...")
	<-a.inputChan

	return nil
}

// monitorProgress monitors game progress with long polling (runs in background goroutine)
func (a *RoomApp) monitorProgress() {
	for {
		select {
		case <-a.stopProgress:
			return
		default:
			// Long polling - this will block until update or timeout
			// But it's OK because it's in a separate goroutine
			a.mu.RLock()
			currentVersion := a.progressVersion
			a.mu.RUnlock()

			progress, err := a.client.GetProgress(currentVersion)
			if err != nil {
				// On error, wait a bit before retrying
				time.Sleep(2 * time.Second)
				continue
			}

			// Update received (or timeout with current state)
			if progress.Version > currentVersion {
				// New update available
				a.mu.Lock()
				a.progressVersion = progress.Version
				a.currentProgress = progress
				a.mu.Unlock()

				// Update screen display (safe to do anytime with cursor save/restore)
				a.screen.UpdateProgress(progress)

				// Check if game finished
				if progress.Status == "finished" {
					a.mu.Lock()
					a.gameFinished = true
					a.mu.Unlock()

					// Notify main loop that game is finished
					select {
					case a.gameFinishedChan <- struct{}{}:
					default:
						// Channel already has notification, skip
					}
					return
				}
			} else {
				// Timeout without update, just update current state
				a.mu.Lock()
				a.currentProgress = progress
				a.mu.Unlock()
			}
		}
	}
}

// showFinalResults displays the final game results and rankings
func (a *RoomApp) showFinalResults(progress *api.RoomProgressResponse) {
	fmt.Println("\n═══════════════════════════════════════")
	fmt.Println("           GAME OVER!")
	fmt.Println("═══════════════════════════════════════")

	if progress.Answer != "" {
		fmt.Printf("The answer was: %s\n\n", progress.Answer)
	}

	if len(progress.Ranking) > 0 {
		fmt.Println("🏆 Final Rankings:")
		for i, playerID := range progress.Ranking {
			player := a.findPlayerByID(progress, playerID)
			if player != nil {
				medal := fmt.Sprintf("%d.", i+1)
				if i == 0 {
					medal = "🥇"
				} else if i == 1 {
					medal = "🥈"
				} else if i == 2 {
					medal = "🥉"
				}

				statusIcon := "❌"
				if player.Status == "won" {
					statusIcon = "✓"
				}

				marker := ""
				if player.PlayerID == a.client.GetPlayerID() {
					marker = " ← YOU"
				}

				fmt.Printf("  %s %s %s - %d rounds%s\n",
					medal, statusIcon, player.Nickname, player.CurrentRound, marker)
			}
		}
	}

	fmt.Println("═══════════════════════════════════════")
}

// findMyProgress finds the current player's progress
func (a *RoomApp) findMyProgress(progress *api.RoomProgressResponse) api.PlayerProgress {
	for _, player := range progress.Players {
		if player.PlayerID == a.client.GetPlayerID() {
			return player
		}
	}
	return api.PlayerProgress{}
}

// findPlayerByID finds a player by ID
func (a *RoomApp) findPlayerByID(progress *api.RoomProgressResponse, playerID string) *api.PlayerProgress {
	for _, player := range progress.Players {
		if player.PlayerID == playerID {
			return &player
		}
	}
	return nil
}

// listRooms lists all available rooms
func (a *RoomApp) listRooms() {
	resp, err := a.client.ListRooms()
	if err != nil {
		fmt.Printf("Error listing rooms: %v\n", err)
		return
	}

	if len(resp.Rooms) == 0 {
		fmt.Println("\n❌ No available rooms. Create one!\n")
		return
	}

	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║       Available Rooms                 ║")
	fmt.Println("╠═══════════════════════════════════════╣")

	hasJoinableRooms := false
	for _, room := range resp.Rooms {
		// Only show rooms that are waiting (joinable)
		if room.Status == "waiting" {
			hasJoinableRooms = true
			playersList := ""
			if len(room.Players) > 0 {
				playersList = fmt.Sprintf(" [%s]", strings.Join(room.Players, ", "))
			}
			fmt.Printf("║ ⏳ Room ID: %-8s %d/%d players%s\n",
				room.RoomID, room.PlayerCount, room.MaxPlayers, playersList)
		}
	}

	if !hasJoinableRooms {
		fmt.Println("║ No joinable rooms available           ║")
	}

	fmt.Println("╚═══════════════════════════════════════╝")
	fmt.Println()
}
