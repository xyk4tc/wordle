package cli

import (
	"fmt"

	"github.com/admin/wordle/internal/game"
)

// Display handles all output formatting and display logic
type Display struct{}

// NewDisplay creates a new Display instance
func NewDisplay() *Display {
	return &Display{}
}

// ShowWelcome displays the welcome message
func (d *Display) ShowWelcome() {
	fmt.Println("Welcome to Wordle!")
	fmt.Println("==================")
}

// ShowGameStart displays game start information
func (d *Display) ShowGameStart(maxRounds int) {
	fmt.Printf("\nGame started! You have %d attempts to guess the 5-letter word.\n", maxRounds)
	fmt.Println("After each guess, you'll see:")
	fmt.Println("  'O' = correct letter in correct spot (Hit)")
	fmt.Println("  '?' = correct letter in wrong spot (Present)")
	fmt.Println("  '_' = letter not in word (Miss)")
	fmt.Println()
}

// ShowPrompt displays the input prompt for current round
func (d *Display) ShowPrompt(currentRound, maxRounds int) {
	fmt.Printf("Attempt %d/%d - Enter your guess: ", currentRound+1, maxRounds)
}

// ShowError displays an error message
func (d *Display) ShowError(err error) {
	fmt.Printf("Error: %v\n", err)
}

// ShowGuessResult displays the result of a single guess
func (d *Display) ShowGuessResult(result game.GuessResult) {
	formatted := game.FormatResult(result)
	fmt.Printf("Result: %s  (%s)\n", formatted, result.Guess)
}

// ShowHistory displays the guess history
func (d *Display) ShowHistory(history []game.GuessResult) {
	if len(history) > 1 {
		fmt.Println("\nYour guesses so far:")
		for i, h := range history {
			fmt.Printf("  %d. %s  %s\n", i+1, h.Guess, game.FormatResult(h))
		}
	}
	fmt.Println()
}

// ShowGameOver displays the game over message
func (d *Display) ShowGameOver(status game.GameStatus, currentRound, maxRounds int, answer string) {
	fmt.Println("==================")
	if status == game.Won {
		fmt.Printf("🎉 Congratulations! You won in %d attempt(s)!\n", currentRound)
	} else {
		fmt.Printf("😔 Game Over! You've used all %d attempts.\n", maxRounds)
		fmt.Printf("The answer was: %s\n", answer)
	}
}

// ShowFinalResults displays the final game results
func (d *Display) ShowFinalResults(history []game.GuessResult) {
	fmt.Println("\nFinal results:")
	for i, h := range history {
		fmt.Printf("  %d. %s  %s\n", i+1, h.Guess, game.FormatResult(h))
	}
}

// ShowConfigError displays configuration error message
func (d *Display) ShowConfigError(err error) {
	fmt.Printf("Error loading configuration: %v\n", err)
	fmt.Println("Using default configuration...")
}

// ShowQuitMessage displays quit message
func (d *Display) ShowQuitMessage() {
	fmt.Println("Thanks for playing!")
}
