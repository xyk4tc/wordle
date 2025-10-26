package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/admin/wordle/pkg/client"
)

func main() {
	// Command line flags
	serverURL := flag.String("server", "http://localhost:8080", "server URL")
	mode := flag.String("mode", "", "game mode: single or multi (if not specified, will prompt)")
	flag.Parse()

	// Show welcome message
	fmt.Println("╔════════════════════════════════════╗")
	fmt.Println("║     Welcome to Wordle Game!        ║")
	fmt.Println("╚════════════════════════════════════╝")
	fmt.Println()

	// Determine game mode
	gameMode := *mode
	if gameMode == "" {
		gameMode = promptMode()
	}

	var err error
	switch strings.ToLower(gameMode) {
	case "single", "1":
		// Single-player mode (Task 2)
		fmt.Println("\n→ Starting Single-Player Mode...")
		app := client.NewApp(*serverURL, os.Stdin)
		err = app.Run()
	case "multi", "2":
		// Multi-player mode (Task 4)
		fmt.Println("\n→ Starting Multi-Player Mode...")
		app := client.NewRoomApp(*serverURL, os.Stdin)
		err = app.Run()
	default:
		fmt.Fprintf(os.Stderr, "Invalid mode: %s\n", gameMode)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func promptMode() string {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Please select game mode:")
		fmt.Println("  1. Single-Player (race against yourself)")
		fmt.Println("  2. Multi-Player  (race against friends)")
		fmt.Print("\nEnter choice (1 or 2): ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1", "single":
			return "single"
		case "2", "multi":
			return "multi"
		default:
			fmt.Println("Invalid choice. Please enter 1 or 2.\n")
		}
	}
}
