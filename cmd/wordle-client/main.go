package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/admin/wordle/pkg/cli"
	"github.com/admin/wordle/pkg/client"
)

func main() {
	// Command line flags
	serverURL := flag.String("server", "http://localhost:8080", "server URL (for online modes)")
	mode := flag.String("mode", "", "game mode: offline, single, or multi (if not specified, will prompt)")
	configPath := flag.String("config", "cfg/config.yaml", "path to configuration file (for offline mode)")
	wordsPath := flag.String("words", "", "path to words list file (for offline mode, overrides config)")
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
	case "offline", "standalone", "0":
		// Offline standalone mode (Task 1)
		fmt.Println("\n→ Starting Offline Mode (no server required)...")
		runner := cli.NewRunner(os.Stdin, *configPath, *wordsPath)
		err = runner.Run()
	case "single", "1":
		// Single-player online mode (Task 2)
		fmt.Println("\n→ Starting Online Single-Player Mode...")
		app := client.NewApp(*serverURL, os.Stdin)
		err = app.Run()
	case "multi", "2":
		// Multi-player online mode (Task 4)
		fmt.Println("\n→ Starting Online Multi-Player Mode...")
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
		fmt.Println("  0. Offline      (standalone, no server required)")
		fmt.Println("  1. Single-Player (online, connect to server)")
		fmt.Println("  2. Multi-Player  (online, race against friends)")
		fmt.Print("\nEnter choice (0, 1, or 2): ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "0", "offline", "standalone":
			return "offline"
		case "1", "single":
			return "single"
		case "2", "multi":
			return "multi"
		default:
			fmt.Println("Invalid choice. Please enter 0, 1, or 2.\n")
		}
	}
}
