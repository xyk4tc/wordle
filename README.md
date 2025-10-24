# Wordle Game

A command-line implementation of the popular Wordle game in Go.

## Features

- Classic Wordle gameplay with 5-letter words
- Configurable maximum number of rounds
- Customizable word list
- Proper Wordle scoring logic:
  - `O` = Hit (correct letter in correct spot)
  - `?` = Present (correct letter in wrong spot)
  - `_` = Miss (letter not in word)

## Project Structure

```
wordle/
├── bin/                     # Binary files (generated)
│   └── wordle              
├── cmd/
│   └── wordle/
│       └── main.go          # Main entry point (with CLI flags)
├── pkg/
│   └── cli/
│       ├── display.go       # Display/output logic
│       ├── input.go         # Input handling
│       └── runner.go        # Game runner/controller
├── internal/
│   ├── game/
│   │   ├── game.go          # Core game logic
│   │   ├── game_test.go     # Game tests
│   │   ├── word.go          # Word validation and scoring
│   │   └── word_test.go     # Word scoring tests
│   └── config/
│       └── config.go        # Configuration management
├── cfg/                     # Configuration directory
│   ├── config.yaml          # Game configuration (includes word list)
│   └── words.txt            # Word list backup (optional)
├── Makefile                 # Build automation
└── go.mod                   # Go module file
```

## Installation

1. Make sure you have Go 1.21 or later installed
2. Clone or download this repository
3. Install dependencies:

```bash
go mod download
```

## Building

### Using Makefile (Recommended)

```bash
# Build the binary to bin/wordle
make build

# Show all available commands
make help
```

### Manual Build

```bash
# Build to bin directory
go build -o bin/wordle ./cmd/wordle

# Or build to current directory
go build -o wordle ./cmd/wordle
```

## Running

### Using Makefile

```bash
# Build and run
make run
```

### Manual Run

```bash
# Run with default configuration (15 words from config.yaml)
./bin/wordle

# Run with larger word list from file
./bin/wordle -words cfg/words.txt

# Run with custom configuration file
./bin/wordle -config path/to/config.yaml

# Run with both custom config and words file
./bin/wordle -config path/to/config.yaml -words path/to/words.txt

# Show help
./bin/wordle -h

# Or run directly with Go (no build needed)
go run ./cmd/wordle

# Run with custom options using Go
go run ./cmd/wordle -words cfg/words.txt
```

## Configuration

### Command Line Arguments

```bash
-config string
    path to configuration file (default "cfg/config.yaml")
-words string
    path to words list file (overrides config word_list)
```

**Note**: 
- By default, the game uses a small word list (15 words) from `cfg/config.yaml`
- Use `-words cfg/words.txt` to load a larger word list (80+ words) from the file

### Configuration File

Edit `cfg/config.yaml` to customize the game:

```yaml
# Maximum number of rounds before game over
max_rounds: 6

# Word list for the game (5-letter words only)
word_list:
  - "CRANE"
  - "SLATE"
  - "ABOUT"
  - "APPLE"
  # ... more words
```

### Configuration Options

- `max_rounds`: Maximum number of attempts the player has (default: 6)
- `word_list`: List of 5-letter words for the game (default: 15 words)

## Word Lists

### Default Word List (cfg/config.yaml)
The default configuration includes 15 carefully selected 5-letter words for quick games:
- Words are defined inline in the YAML file
- Good for testing and quick games
- Can be customized by editing `cfg/config.yaml`

### Extended Word List (cfg/words.txt)
For a more challenging game, use the extended word list:
- Contains 80+ diverse 5-letter words
- Load with: `./bin/wordle -words cfg/words.txt`
- One word per line format

All words should:
- Be exactly 5 letters long
- Contain only alphabetic characters (A-Z)
- Be case-insensitive

## Gameplay

1. The game randomly selects a 5-letter word from the configured word list
2. You have a maximum number of attempts (default: 6) to guess the word
3. After each guess, you'll see the result:
   - `O` = correct letter in correct position (Hit)
   - `?` = correct letter in wrong position (Present)
   - `_` = letter not in the word (Miss)
4. Win by guessing the word within the allowed attempts
5. Type `quit` or `exit` to exit the game

### Example

```
Welcome to Wordle!
==================

Game started! You have 6 attempts to guess the 5-letter word.
After each guess, you'll see:
  'O' = correct letter in correct spot (Hit)
  '?' = correct letter in wrong spot (Present)
  '_' = letter not in word (Miss)

Attempt 1/6 - Enter your guess: crane
Result: _?___  (CRANE)

Attempt 2/6 - Enter your guess: slime
Result: _?_?O  (SLIME)

Attempt 3/6 - Enter your guess: apple
Result: OOOOO  (APPLE)

==================
🎉 Congratulations! You won in 3 attempt(s)!

Final results:
  1. CRANE  _?___
  2. SLIME  _?_?O
  3. APPLE  OOOOO
```

## Testing

### Using Makefile

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage
```

### Manual Testing

```bash
# Test all packages
go test ./...

# Test with verbose output
go test ./internal/game -v

# Test with coverage
go test ./... -cover
```

## Available Make Commands

```bash
make build          # Build the binary to bin/wordle
make run            # Build and run the game
make test           # Run all tests
make test-coverage  # Run tests with coverage report
make clean          # Remove build artifacts
make deps           # Download and tidy dependencies
make fmt            # Format code
make lint           # Run linter
make help           # Show help message
```

## Implementation Details

### Core Game Logic

The game logic is implemented in the `internal/game` package:

- **Word Validation**: Ensures all words are 5 letters and alphabetic only
- **Scoring Algorithm**: Implements the exact Wordle scoring logic:
  1. First pass: Mark all exact matches (Hit)
  2. Second pass: Mark Present for remaining letters
  3. Correctly handles duplicate letters

### Configuration Management

The `internal/config` package handles:
- Loading configuration from YAML files
- Loading word lists from files
- Providing default configuration

### Extensibility

The modular design allows for easy extensions:
- Add new game modes
- Implement different word lengths
- Add difficulty levels
- Create web or GUI interfaces

## License

This is an educational project for learning purposes.
