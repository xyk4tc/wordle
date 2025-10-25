package client

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/admin/wordle/pkg/api"
)

// App represents the client application
type App struct {
	client *Client
	reader *bufio.Scanner
}

// NewApp creates a new client application
func NewApp(serverURL string, reader io.Reader) *App {
	return &App{
		client: NewClient(serverURL),
		reader: bufio.NewScanner(reader),
	}
}

// Run starts the client application
func (a *App) Run() error {
	a.showWelcome()

	// Create new game on server
	fmt.Println("Connecting to server and creating new game...")
	gameResp, err := a.client.NewGame()
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}

	a.showGameInfo(gameResp)

	// Game loop
	currentRound := 0
	for {
		fmt.Printf("Attempt %d/%d - Enter your guess: ", currentRound+1, gameResp.MaxRounds)

		if !a.reader.Scan() {
			break
		}

		guess := strings.TrimSpace(a.reader.Text())

		// Check for quit command
		if strings.ToLower(guess) == "quit" || strings.ToLower(guess) == "exit" {
			fmt.Println("Thanks for playing!")
			break
		}

		// Send guess to server
		response, err := a.client.MakeGuess(guess)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Display result
		a.displayResult(response)
		currentRound = response.CurrentRound

		// Check if game is over
		if response.GameOver {
			a.showGameOver(response)
			break
		}
	}

	return nil
}

// showWelcome displays the welcome message
func (a *App) showWelcome() {
	fmt.Println("Welcome to Wordle! (Client Mode)")
	fmt.Println("=================================")
}

// showGameInfo displays game information
func (a *App) showGameInfo(gameResp *api.NewGameResponse) {
	fmt.Printf("\n%s\n", gameResp.Message)
	fmt.Printf("Game ID: %s\n", gameResp.GameID)
	fmt.Printf("You have %d attempts to guess the 5-letter word.\n", gameResp.MaxRounds)
	fmt.Println("\nAfter each guess, you'll see:")
	fmt.Println("  'O' = correct letter in correct spot (Hit)")
	fmt.Println("  '?' = correct letter in wrong spot (Present)")
	fmt.Println("  '_' = letter not in word (Miss)")
	fmt.Println()
}

// displayResult displays the result of a guess
func (a *App) displayResult(response *api.GuessResponse) {
	// Results is already an array of display characters
	result := strings.Join(response.Results, "")
	fmt.Printf("Result: %s  (%s)\n", result, response.Guess)

	if response.Message != "" {
		fmt.Println(response.Message)
	}
	fmt.Println()
}

// showGameOver displays game over information
func (a *App) showGameOver(response *api.GuessResponse) {
	fmt.Println("\n==================")
	if response.GameStatus == "won" {
		fmt.Printf("🎉 %s\n", response.Message)
		fmt.Printf("You guessed it in %d attempt(s)!\n", response.CurrentRound)
	} else if response.GameStatus == "lost" {
		fmt.Printf("😔 %s\n", response.Message)
	}
	if response.Answer != "" {
		fmt.Printf("The answer was: %s\n", response.Answer)
	}

	// Display final history
	status, err := a.client.GetStatus()
	if err == nil {
		fmt.Println("\nFinal results:")
		for i, h := range status.History {
			result := strings.Join(h.Results, "")
			fmt.Printf("  %d. %s  %s\n", i+1, h.Guess, result)
		}
	}
}
