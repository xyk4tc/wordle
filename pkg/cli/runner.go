package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/admin/wordle/internal/config"
	"github.com/admin/wordle/internal/game"
)

// Runner manages the game execution flow
type Runner struct {
	display    *Display
	input      *InputReader
	configPath string
	wordsPath  string
}

// NewRunner creates a new game runner
func NewRunner(reader io.Reader, configPath string, wordsPath string) *Runner {
	return &Runner{
		display:    NewDisplay(),
		input:      NewInputReader(reader),
		configPath: configPath,
		wordsPath:  wordsPath,
	}
}

// Run starts and manages the game loop
func (r *Runner) Run() error {
	// Show welcome
	r.display.ShowWelcome()

	// Load configuration
	cfg, err := r.loadConfiguration()
	if err != nil {
		r.display.ShowConfigError(err)
		cfg = config.DefaultConfig()
	}

	// Create game
	g, err := game.NewGame(cfg.MaxRounds, cfg.WordList)
	if err != nil {
		return fmt.Errorf("error creating game: %w", err)
	}

	// Show game start info
	r.display.ShowGameStart(cfg.MaxRounds)

	// Run game loop
	r.runGameLoop(g)

	// Show game over
	r.display.ShowGameOver(g.GetStatus(), g.CurrentRound, g.MaxRounds, g.Answer)
	r.display.ShowFinalResults(g.History)

	return nil
}

// runGameLoop executes the main game loop
func (r *Runner) runGameLoop(g *game.Game) {
	for !g.IsGameOver() {
		r.display.ShowPrompt(g.CurrentRound, g.MaxRounds)

		guess, ok := r.input.ReadGuess()
		if !ok {
			break
		}

		// Check for quit command
		if IsQuitCommand(guess) {
			r.display.ShowQuitMessage()
			os.Exit(0)
		}

		// Process guess
		result, err := g.MakeGuess(guess)
		if err != nil {
			r.display.ShowError(err)
			continue
		}

		// Display results
		r.display.ShowGuessResult(result)
		r.display.ShowHistory(g.History)
	}
}

// loadConfiguration tries to load config from file, returns error if not found
func (r *Runner) loadConfiguration() (*config.Config, error) {
	// Try to load from specified config path
	if _, err := os.Stat(r.configPath); err != nil {
		return nil, fmt.Errorf("configuration file not found: %s", r.configPath)
	}

	cfg, err := config.LoadConfig(r.configPath)
	if err != nil {
		return nil, err
	}

	// If words file is specified, load words from file and override config word_list
	if r.wordsPath != "" {
		words, err := config.LoadWordsFromFile(r.wordsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load words file: %w", err)
		}
		cfg.WordList = words
	}

	return cfg, nil
}
