package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/admin/wordle/pkg/api"
)

// ScreenManager manages split-screen display for multiplayer game
type ScreenManager struct {
	// Dynamic layout
	numPlayers    int // Current number of players
	progressStart int // Line number where progress starts
	progressEnd   int // Line number where progress ends
	logStart      int // Line number where log starts
	logEnd        int // Line number where log ends
	inputLine     int // Line number for input

	// Config
	maxLogLines int      // Maximum log lines to keep
	logBuffer   []string // Rolling log buffer
}

// NewScreenManager creates a new screen manager
func NewScreenManager() *ScreenManager {
	return &ScreenManager{
		maxLogLines: 10,
		logBuffer:   make([]string, 0),
	}
}

// ANSI Escape Codes
const (
	// Cursor control (templates with %d for parameters)
	AnsiCursorUp      = "\033[%dA"    // Move cursor up N lines
	AnsiCursorDown    = "\033[%dB"    // Move cursor down N lines
	AnsiCursorForward = "\033[%dC"    // Move cursor forward N columns
	AnsiCursorBack    = "\033[%dD"    // Move cursor back N columns
	AnsiCursorPos     = "\033[%d;%dH" // Move cursor to row;col
	AnsiCursorHome    = "\033[H"      // Move cursor to home (1,1)
	AnsiCursorSave    = "\033[s"      // Save cursor position
	AnsiCursorRestore = "\033[u"      // Restore cursor position

	// Screen control
	AnsiClearScreen      = "\033[2J" // Clear entire screen
	AnsiClearLine        = "\033[2K" // Clear entire line
	AnsiClearLineRight   = "\033[K"  // Clear from cursor to end of line
	AnsiClearScreenBelow = "\033[J"  // Clear from cursor to end of screen

	// Alternate Screen Buffer (like vi/less/top)
	AnsiEnterAltScreen = "\033[?1049h" // Switch to alternate screen buffer
	AnsiExitAltScreen  = "\033[?1049l" // Return to main screen buffer

	// Cursor visibility
	AnsiHideCursor = "\033[?25l" // Hide cursor
	AnsiShowCursor = "\033[?25h" // Show cursor

	// Colors
	AnsiColorReset  = "\033[0m"
	AnsiColorBold   = "\033[1m"
	AnsiColorGreen  = "\033[32m"
	AnsiColorYellow = "\033[33m"
	AnsiColorBlue   = "\033[34m"
	AnsiColorCyan   = "\033[36m"
)

// InitScreen initializes the split-screen layout
// It calculates layout while drawing to ensure they stay in sync
func (sm *ScreenManager) InitScreen(numPlayers int) {
	sm.numPlayers = numPlayers

	// Enter alternate screen buffer (like vi/less)
	// This preserves the user's terminal history
	output := AnsiEnterAltScreen

	// Hide cursor for cleaner display during updates
	output += AnsiHideCursor

	// Clear screen and move cursor to top
	output += AnsiClearScreen
	output += AnsiCursorHome

	fmt.Print(output)

	// Track line number while drawing (1-indexed for ANSI cursor positioning)
	line := 1

	// Line 1: Top border
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	line++

	// Line 2: Title
	fmt.Println("║                    🏆 Live Progress                        ║")
	line++

	// Line 3: Separator
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	line++

	// Lines 4+: Progress section (one line per player)
	sm.progressStart = line
	for i := 0; i < sm.numPlayers; i++ {
		fmt.Println("║                                                            ║")
		line++
	}
	sm.progressEnd = line - 1

	// Next: Separator
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	line++

	// Next: Log header
	fmt.Println("║  Game Log:                                                 ║")
	line++

	// Next: Separator
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	line++

	// Next lines: Log content
	sm.logStart = line
	for i := 0; i < sm.maxLogLines; i++ {
		fmt.Println("║                                                            ║")
		line++
	}
	sm.logEnd = line - 1

	// Next: Separator
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	line++

	// Next: Input Area header
	fmt.Println("║  Input Area:                                               ║")
	line++

	// Next: Actual input line (where user types)
	sm.inputLine = line
	fmt.Println("║                                                            ║")
	line++

	// Last: Bottom border
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	line++

	// Force flush all output to ensure everything is visible
	os.Stdout.Sync()
}

// UpdateProgress updates the progress section (top area)
func (sm *ScreenManager) UpdateProgress(progress *api.RoomProgressResponse) {
	// Check if player count changed - need full redraw
	if len(progress.Players) != sm.numPlayers {
		sm.FullRedraw(progress)
		return
	}

	// Build output buffer
	output := AnsiCursorSave

	// Update each player line
	for i, player := range progress.Players {
		// Move to this player's line
		lineNum := sm.progressStart + i
		moveCursor := fmt.Sprintf(AnsiCursorPos, lineNum, 1)
		output += moveCursor

		// Clear line
		output += AnsiClearLine

		// Draw content
		statusIcon := "🎮"
		if player.Status == "won" {
			statusIcon = "🏆"
		} else if player.Status == "lost" {
			statusIcon = "❌"
		}

		lastResult := ""
		if player.LastGuess != nil {
			lastResult = strings.Join(player.LastGuess.Results, "")
		}

		info := fmt.Sprintf("%s %-10s: Round %d/%d %s",
			statusIcon, player.Nickname, player.CurrentRound, player.MaxRounds, lastResult)

		// Pad to 56 chars (60 - 4 for borders)
		if len(info) > 56 {
			info = info[:56]
		} else {
			info += strings.Repeat(" ", 56-len(info))
		}

		output += fmt.Sprintf("║ %s ║", info)
	}

	// Restore cursor position
	output += AnsiCursorRestore

	// Force flush output
	fmt.Print(output)
	os.Stdout.Sync()
}

// FullRedraw redraws the entire screen (when layout changes)
func (sm *ScreenManager) FullRedraw(progress *api.RoomProgressResponse) {
	// Redraw everything (InitScreen will recalculate layout)
	sm.InitScreen(len(progress.Players))
	sm.UpdateProgress(progress)

	// Restore all log lines at once
	sm.redrawAllLogs()
}

// redrawAllLogs redraws all log lines in the buffer
func (sm *ScreenManager) redrawAllLogs() {
	output := AnsiCursorSave

	// Draw all log lines
	for i := 0; i < len(sm.logBuffer); i++ {
		lineNum := sm.logStart + i
		moveCursor := fmt.Sprintf(AnsiCursorPos, lineNum, 1)
		output += moveCursor
		output += AnsiClearLine

		logLine := sm.logBuffer[i]
		if len(logLine) > 56 {
			logLine = logLine[:56]
		} else {
			logLine += strings.Repeat(" ", 56-len(logLine))
		}
		output += fmt.Sprintf("║  %s ║", logLine)
	}

	// Restore cursor
	output += AnsiCursorRestore

	fmt.Print(output)
	os.Stdout.Sync()
}

// AddLogLine adds a line to the game log (middle area)
func (sm *ScreenManager) AddLogLine(line string) {
	// Add to buffer
	sm.logBuffer = append(sm.logBuffer, line)
	if len(sm.logBuffer) > sm.maxLogLines {
		sm.logBuffer = sm.logBuffer[1:] // Remove oldest
	}

	// Redraw all logs to show the new one
	sm.redrawAllLogs()
}

// PromptInput shows the input prompt at the bottom
func (sm *ScreenManager) PromptInput(round, maxRounds int) {
	// Build the prompt text
	promptText := fmt.Sprintf("Round %d/%d - Enter your guess: ", round, maxRounds)

	// Pad to fill the line (60 - 4 for borders = 56 chars content, -2 for leading spaces)
	contentWidth := 56
	paddedPrompt := "  " + promptText
	if len(paddedPrompt) < contentWidth {
		paddedPrompt += strings.Repeat(" ", contentWidth-len(paddedPrompt))
	} else if len(paddedPrompt) > contentWidth {
		paddedPrompt = paddedPrompt[:contentWidth]
	}

	// Move cursor to input line and draw the full bordered line
	moveCursor := fmt.Sprintf(AnsiCursorPos, sm.inputLine, 1)
	output := moveCursor
	output += AnsiClearLine
	output += fmt.Sprintf("║%s║", paddedPrompt)

	// Move cursor to the input position (after the prompt text)
	// Position = 1 (border) + 2 (spaces) + len(promptText)
	inputCol := 3 + len(promptText)
	moveCursorToInput := fmt.Sprintf(AnsiCursorPos, sm.inputLine, inputCol)
	output += moveCursorToInput

	// Show cursor for user input
	output += AnsiShowCursor

	// Print and flush
	fmt.Print(output)
	os.Stdout.Sync()
}

// ClearInputLine clears the input line (removes user's typed input)
func (sm *ScreenManager) ClearInputLine() {
	// Move to input line and clear it, then redraw the empty bordered line
	moveCursor := fmt.Sprintf(AnsiCursorPos, sm.inputLine, 1)
	output := moveCursor
	output += AnsiClearLine
	// 60 chars total - 2 for borders = 58 chars between borders
	output += "║" + strings.Repeat(" ", 58) + "║"

	fmt.Print(output)
	os.Stdout.Sync()
}

// CleanupScreen restores normal terminal
func (sm *ScreenManager) CleanupScreen() {
	// Build output buffer
	output := AnsiShowCursor
	output += AnsiExitAltScreen

	fmt.Print(output)
}
