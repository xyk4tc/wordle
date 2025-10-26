package client

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/admin/wordle/pkg/api"
	"github.com/mattn/go-runewidth"
)

// ScreenManager manages split-screen display for multiplayer game
type ScreenManager struct {
	mu sync.Mutex // Protects concurrent access to screen state

	// Dynamic layout
	numPlayers    int // Current number of players
	progressStart int // Line number where progress starts
	progressEnd   int // Line number where progress ends
	logStart      int // Line number where log starts
	logEnd        int // Line number where log ends
	inputLine     int // Line number for input
	inputCol      int // Column position for input cursor (to restore after updates)

	// Config
	maxLogLines int      // Maximum log lines to keep
	logBuffer   []string // Rolling log buffer
}

// NewScreenManager creates a new screen manager
func NewScreenManager() *ScreenManager {
	return &ScreenManager{
		maxLogLines: 10,
		logBuffer:   make([]string, 0),
		inputCol:    1, // Default to column 1 (will be updated by PromptInput)
	}
}

// padOrTruncate pads or truncates a string to exactly targetWidth display columns
// This properly handles emojis and CJK characters that occupy 2 columns
func padOrTruncate(s string, targetWidth int) string {
	currentWidth := runewidth.StringWidth(s)

	if currentWidth == targetWidth {
		return s
	}

	if currentWidth > targetWidth {
		// Need to truncate - use runewidth.Truncate for accurate width-based truncation
		return runewidth.Truncate(s, targetWidth, "")
	}

	// Need to pad - add spaces to reach target width
	padding := targetWidth - currentWidth
	return s + strings.Repeat(" ", padding)
}

// centerText centers text within targetWidth display columns
func centerText(s string, targetWidth int) string {
	currentWidth := runewidth.StringWidth(s)

	if currentWidth >= targetWidth {
		return runewidth.Truncate(s, targetWidth, "")
	}

	// Calculate padding on both sides
	totalPadding := targetWidth - currentWidth
	leftPadding := totalPadding / 2
	rightPadding := totalPadding - leftPadding

	return strings.Repeat(" ", leftPadding) + s + strings.Repeat(" ", rightPadding)
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

	// Content width = 60 (total) - 2 (borders) = 58 columns
	const contentWidth = 58
	const totalWidth = 60

	// Line 1: Top border - generate to ensure correct width
	topBorder := "╔" + strings.Repeat("═", totalWidth-2) + "╗"
	fmt.Println(topBorder)
	line++

	// Line 2: Title (centered, considering emoji width)
	titleText := centerText("🏆 Live Progress", contentWidth)
	fmt.Printf("║%s║\n", titleText)
	line++

	// Line 3: Separator - generate to ensure correct width
	separator := "╠" + strings.Repeat("═", totalWidth-2) + "╣"
	fmt.Println(separator)
	line++

	// Lines 4+: Progress section (one line per player)
	sm.progressStart = line
	emptyLine := strings.Repeat(" ", contentWidth)
	for i := 0; i < sm.numPlayers; i++ {
		fmt.Printf("║%s║\n", emptyLine)
		line++
	}
	sm.progressEnd = line - 1

	// Next: Separator
	fmt.Println(separator)
	line++

	// Next: Log header (left-aligned with 2 spaces padding)
	logHeaderText := padOrTruncate("  Game Log:", contentWidth)
	fmt.Printf("║%s║\n", logHeaderText)
	line++

	// Next: Separator
	fmt.Println(separator)
	line++

	// Next lines: Log content
	sm.logStart = line
	for i := 0; i < sm.maxLogLines; i++ {
		fmt.Printf("║%s║\n", emptyLine)
		line++
	}
	sm.logEnd = line - 1

	// Next: Separator
	fmt.Println(separator)
	line++

	// Next: Input Area header (left-aligned with 2 spaces padding)
	inputHeaderText := padOrTruncate("  Input Area:", contentWidth)
	fmt.Printf("║%s║\n", inputHeaderText)
	line++

	// Next: Actual input line (where user types)
	sm.inputLine = line
	fmt.Printf("║%s║\n", emptyLine)
	line++

	// Last: Bottom border - generate to ensure correct width
	bottomBorder := "╚" + strings.Repeat("═", totalWidth-2) + "╝"
	fmt.Println(bottomBorder)
	line++

	// Force flush all output to ensure everything is visible
	os.Stdout.Sync()
}

// UpdateProgress updates the progress section (top area)
func (sm *ScreenManager) UpdateProgress(progress *api.RoomProgressResponse) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if player count changed - need full redraw
	if len(progress.Players) != sm.numPlayers {
		sm.fullRedrawLocked(progress)
		return
	}

	// Build output buffer
	output := ""

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

		// Pad nickname to exactly 10 display columns for alignment
		paddedNickname := padOrTruncate(player.Nickname, 10)

		info := fmt.Sprintf("%s %s: Round %d/%d %s",
			statusIcon, paddedNickname, player.CurrentRound, player.MaxRounds, lastResult)

		// Pad to 58 display columns (60 - 2 borders)
		// Format: "║{58 cols}║" - consistent with all other lines
		info = padOrTruncate(info, 58)

		output += fmt.Sprintf("║%s║", info)
	}

	// Move cursor back to input position (line and column)
	moveCursorToInput := fmt.Sprintf(AnsiCursorPos, sm.inputLine, sm.inputCol)
	output += moveCursorToInput

	// Force flush output
	fmt.Print(output)
	os.Stdout.Sync()
}

// FullRedraw redraws the entire screen (when layout changes)
func (sm *ScreenManager) FullRedraw(progress *api.RoomProgressResponse) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.fullRedrawLocked(progress)
}

// fullRedrawLocked is the internal implementation without locking
func (sm *ScreenManager) fullRedrawLocked(progress *api.RoomProgressResponse) {
	// Note: InitScreen doesn't need lock protection as it only writes to terminal
	sm.InitScreen(len(progress.Players))

	// Build output for progress update (similar to UpdateProgress but without lock)
	output := ""
	for i, player := range progress.Players {
		lineNum := sm.progressStart + i
		moveCursor := fmt.Sprintf(AnsiCursorPos, lineNum, 1)
		output += moveCursor
		output += AnsiClearLine

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

		// Pad nickname to exactly 10 display columns for alignment
		paddedNickname := padOrTruncate(player.Nickname, 10)

		info := fmt.Sprintf("%s %s: Round %d/%d %s",
			statusIcon, paddedNickname, player.CurrentRound, player.MaxRounds, lastResult)

		// Pad to 58 display columns (60 - 2 borders)
		// Format: "║{58 cols}║" - consistent with all other lines
		info = padOrTruncate(info, 58)

		output += fmt.Sprintf("║%s║", info)
	}

	moveCursorToInput := fmt.Sprintf(AnsiCursorPos, sm.inputLine, sm.inputCol)
	output += moveCursorToInput

	fmt.Print(output)
	os.Stdout.Sync()

	// Restore all log lines
	sm.redrawAllLogsLocked()
}

// redrawAllLogsLocked redraws all log lines (must be called with lock held)
func (sm *ScreenManager) redrawAllLogsLocked() {
	output := ""

	// Draw all log lines
	for i := 0; i < len(sm.logBuffer); i++ {
		lineNum := sm.logStart + i
		moveCursor := fmt.Sprintf(AnsiCursorPos, lineNum, 1)
		output += moveCursor
		output += AnsiClearLine

		logLine := sm.logBuffer[i]
		// Pad to 58 display columns (60 - 2 borders)
		// Format: "║{58 cols}║" - consistent with all other lines
		logLine = padOrTruncate(logLine, 58)
		output += fmt.Sprintf("║%s║", logLine)
	}

	// Move cursor back to input position (line and column)
	moveCursorToInput := fmt.Sprintf(AnsiCursorPos, sm.inputLine, sm.inputCol)
	output += moveCursorToInput

	fmt.Print(output)
	os.Stdout.Sync()
}

// AddLogLine adds a line to the game log (middle area)
func (sm *ScreenManager) AddLogLine(line string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Add to buffer
	sm.logBuffer = append(sm.logBuffer, line)
	if len(sm.logBuffer) > sm.maxLogLines {
		sm.logBuffer = sm.logBuffer[1:] // Remove oldest
	}

	// Redraw all logs to show the new one
	sm.redrawAllLogsLocked()
}

// PromptInput shows the input prompt at the bottom
func (sm *ScreenManager) PromptInput(round, maxRounds int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Build the prompt text
	promptText := fmt.Sprintf("Round %d/%d - Enter your guess: ", round, maxRounds)

	// Pad to fill the line (60 - 2 borders = 58 display columns)
	const contentWidth = 58
	paddedPrompt := "  " + promptText
	paddedPrompt = padOrTruncate(paddedPrompt, contentWidth)

	// Move cursor to input line and draw the full bordered line
	moveCursor := fmt.Sprintf(AnsiCursorPos, sm.inputLine, 1)
	output := moveCursor
	output += AnsiClearLine
	output += fmt.Sprintf("║%s║", paddedPrompt)

	// Move cursor to the input position (after the prompt text)
	// Position = 1 (border) + 2 (spaces) + display width of promptText
	sm.inputCol = 3 + runewidth.StringWidth(promptText)
	moveCursorToInput := fmt.Sprintf(AnsiCursorPos, sm.inputLine, sm.inputCol)
	output += moveCursorToInput

	// Show cursor for user input
	output += AnsiShowCursor

	// Print and flush
	fmt.Print(output)
	os.Stdout.Sync()
}

// ClearInputLine clears the input line (removes user's typed input)
func (sm *ScreenManager) ClearInputLine() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Move to input line and clear it, then redraw the empty bordered line
	moveCursor := fmt.Sprintf(AnsiCursorPos, sm.inputLine, 1)
	output := moveCursor
	output += AnsiClearLine
	// 60 chars total - 2 for borders = 58 chars between borders
	output += "║" + strings.Repeat(" ", 58) + "║"

	// Reset inputCol to line start since input area is now clear
	sm.inputCol = 1

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
