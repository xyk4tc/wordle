# Wordle Game

A command-line implementation of the popular Wordle game in Go.

## Features

### Task 1: Standalone Mode
- Classic Wordle gameplay with 5-letter words
- Configurable maximum number of rounds
- Customizable word list
- Proper Wordle scoring logic:
  - `O` = Hit (correct letter in correct spot)
  - `?` = Present (correct letter in wrong spot)
  - `_` = Miss (letter not in word)

### Task 2: Server/Client Mode
- Server/Client architecture for multiplayer support
- RESTful API with elegant routing (Gin framework)
- Client-side never knows the answer until game over
- Server-side input validation
- Game session management
- Full history tracking

## Tech Stack

- **Language**: Go 1.24+
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin) - High-performance HTTP web framework
- **Configuration**: YAML (gopkg.in/yaml.v3)
- **Architecture**: Clean layered architecture with separation of concerns

## Project Structure

```
wordle/
├── bin/                     # Binary files (generated)
│   ├── wordle               # Standalone game (Task 1)
│   ├── wordle-server        # Server (Task 2)
│   └── wordle-client        # Client (Task 2)
├── cmd/
│   ├── wordle/              # Standalone game entry
│   │   └── main.go
│   ├── wordle-server/       # Server entry (Task 2)
│   │   └── main.go
│   └── wordle-client/       # Client entry (Task 2)
│       └── main.go
├── pkg/
│   ├── api/                 # API protocol (Task 2)
│   │   └── protocol.go
│   ├── cli/                 # CLI interface (Task 1)
│   │   ├── display.go
│   │   ├── input.go
│   │   └── runner.go
│   ├── client/              # Client library (Task 2)
│   │   └── client.go
│   └── server/              # Server library (Task 2)
│       ├── server.go
│       └── session.go
├── internal/
│   ├── game/                # Core game logic (shared)
│   │   ├── game.go
│   │   ├── game_test.go
│   │   ├── word.go
│   │   └── word_test.go
│   └── config/              # Configuration management
│       └── config.go
├── cfg/                     # Configuration directory
│   ├── config.yaml          # Game configuration
│   └── words.txt            # Extended word list
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

## Task 2: Server/Client Mode

### Quick Start

**Terminal 1 - Start Server:**
```bash
make run-server
# Or manually:
./bin/wordle-server -port 8080
```

**Terminal 2 - Start Client:**
```bash
make run-client
# Or manually:
./bin/wordle-client -server http://localhost:8080
```

### Server

The server provides a RESTful API for managing Wordle games:

**Start server:**
```bash
./bin/wordle-server [options]
```

**Options:**
- `-config string`: Path to configuration file (default: "cfg/config.yaml")
- `-words string`: Path to words list file (overrides config word_list)
- `-port string`: Server port (default: "8080")

**Examples:**
```bash
# Start with default config (15 words)
./bin/wordle-server

# Start with extended word list
./bin/wordle-server -words cfg/words.txt

# Start with custom config and port
./bin/wordle-server -config my_config.yaml -port 9090
```

**API Endpoints:**
```
POST /game/new           - Create a new game
POST /game/:id/guess     - Submit a guess
GET  /game/:id/status    - Get game status
```

**Example API Usage:**

```bash
# Create new game (uses server configuration)
curl -X POST http://localhost:8080/game/new \
  -H "Content-Type: application/json" \
  -d '{}'

# Response: {"game_id":"1","max_rounds":6,"message":"Game created successfully"}

# Submit guess
curl -X POST http://localhost:8080/game/1/guess \
  -H "Content-Type: application/json" \
  -d '{"guess": "APPLE"}'

# Response includes results as array of display characters
# Example: {"guess":"APPLE","results":["?","_","_","O","_"],...}

# Get game status
curl http://localhost:8080/game/1/status
```

**API Design Principles:**
- RESTful design with resource-based URLs
- Server controls all game settings (max_rounds, word_list)
- Client cannot override server configuration
- Server returns display-ready format ('O', '?', '_') - no client-side conversion needed
- Ensures consistent game experience for all players
- Uses Gin framework for elegant routing and parameter handling

### Client

The client connects to the server and provides an interactive command-line interface:

**Start client:**
```bash
./bin/wordle-client [options]
```

**Options:**
- `-server string`: Server URL (default: "http://localhost:8080")

**Note:** Game settings (max rounds, word list) are determined by the server configuration only.

**Features:**
- Client never knows the answer until game over
- Server validates all inputs
- Server controls game configuration (security)
- Full game history available
- Same user experience as standalone mode

### Key Differences from Standalone Mode

| Feature | Standalone (Task 1) | Server/Client (Task 2) |
|---------|-------------------|----------------------|
| Architecture | Single binary | Separate server + client |
| Answer visibility | In client memory | Only on server |
| Input validation | Client-side | Server-side |
| Game configuration | Client controls | Server controls |
| Multi-player | No | Yes (via shared server) |
| Network required | No | Yes |

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
