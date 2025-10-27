# Wordle Game

A feature-rich command-line implementation of the popular Wordle game in Go, supporting offline, single-player online, and competitive multi-player modes.

---

## Table of Contents

- [Installation & Building](#installation--building)
- [Quick Start](#quick-start)
- [Features](#features)
- [Architecture Overview](#architecture-overview)
- [Configuration](#configuration) ⚙️
- [Usage Guide](#usage-guide)
  - [Mode 0: Offline Mode](#mode-0-offline-mode)
  - [Mode 1: Single-Player Online](#mode-1-single-player-online)
  - [Mode 2: Multi-Player Racing](#mode-2-multi-player-racing)
- [Design & Architecture](#design--architecture)
  - [Unified Client Architecture](#unified-client-architecture)
  - [Multi-Player System Design](#multi-player-system-design)
  - [Real-Time Updates: Long Polling](#real-time-updates-long-polling)
  - [UI/UX: Terminal Design](#uiux-terminal-design)
  - [Design Tradeoffs](#design-tradeoffs)
- [Implementation Details](#implementation-details)
- [Future Enhancements](#future-enhancements)
- [Development](#development)

---

## Installation & Building

### Prerequisites

- Go 1.24 or later
- Terminal with ANSI support (most modern terminals)

### Building

```bash
# Clone the repository (if you haven't already)
git clone https://github.com/xyk4tc/wordle.git
cd wordle

# Install dependencies
go mod download

# Build all binaries
make build

# Or build individually
make build-server          # Server only
make build-client          # Client only

# Manual build
go build -o bin/wordle-server ./cmd/wordle-server
go build -o bin/wordle-client ./cmd/wordle-client
```

### Running Tests

```bash
make test                  # Run all tests
make test-coverage         # With coverage report
go test ./...              # Manual
```

---

## Quick Start

```bash
# Start server (for online modes)
./bin/wordle-server

# In a new terminal window
# Start the client (choose mode in interactive prompt)
./bin/wordle-client

# Or specify mode directly
./bin/wordle-client -mode offline          # Play offline
./bin/wordle-client -mode single           # Online single-player
./bin/wordle-client -mode multi            # Online multi-player
```

> 💡 **Tip**: For customization options (word lists, ports, etc.), see [Configuration](#configuration)

---

## Features

### 🎮 Three Game Modes in One Binary

**Mode 0: Offline** 🏠
- No server required - play anywhere, anytime
- Customizable word lists and difficulty
- Perfect for practice and learning
- Zero network latency

**Mode 1: Single-Player Online** 🎯
- Server-authoritative gameplay (anti-cheat)
- Client never knows the answer
- RESTful API design
- Full game history tracking

**Mode 2: Multi-Player Racing** 🏆
- Competitive race: 2-8 players, same word
- **Real-time updates** via long polling
- Room-based lobbies
- Live rankings and leaderboards
- Host-controlled game start

### 🎨 Professional Terminal UI

- **Alternate screen buffer** (like vim/less)
- **ANSI escape codes** for rich formatting
- **Unicode-aware** (perfect alignment with emojis/CJK)
- **Real-time updates** without flicker
- Clean exit - preserves terminal history

### 🛡️ Robust Architecture

- Clean layered design
- Concurrent-safe (goroutines + mutexes)
- Graceful shutdown (context cancellation)
- No goroutine leaks
- Production-ready error handling

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    wordle-client Binary                     │
│                  (Unified Multi-Mode Client)                │
├─────────────────────────────────────────────────────────────┤
│                    Mode Selection Layer                     │
└───────────────┬─────────────────┬─────────────┬─────────────┘
                │                 │             │
        ┌───────▼───────┐ ┌───────▼──────┐ ┌───▼──────┐
        │  Mode 0       │ │  Mode 1      │ │ Mode 2   │
        │  Offline      │ │  Single-     │ │ Multi-   │
        │  (pkg/cli)    │ │  Player      │ │ Player   │
        └───────────────┘ └──────┬───────┘ └────┬─────┘
                                 │               │
                          ┌──────▼───────────────▼──────┐
                          │    wordle-server            │
                          │  (RESTful API + Rooms)      │
                          └─────────────────────────────┘
```

### Tech Stack

- **Language**: Go 1.24+
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin) (HTTP routing)
- **Concurrency**: `sync.Cond`, `context.Context`, `errgroup`
- **Terminal UI**: ANSI escape codes, [go-runewidth](https://github.com/mattn/go-runewidth)
- **Configuration**: YAML

### Project Structure

```
wordle/
├── bin/                     # Compiled binaries
│   ├── wordle-server        # Server (29MB)
│   └── wordle-client        # Unified client (9.3MB)
├── cmd/                     # Entry points
│   ├── wordle-server/main.go
│   └── wordle-client/main.go
├── pkg/                     # Public libraries
│   ├── api/                 # API protocol definitions
│   ├── cli/                 # Offline mode (display, input, runner)
│   ├── client/              # Online client (HTTP, rooms, screen manager)
│   └── server/              # Server logic (API, rooms, sessions)
├── internal/                # Private libraries
│   ├── game/                # Core Wordle logic (shared)
│   └── config/              # Configuration loader
├── cfg/                     # Configuration files
│   ├── config.yaml          # Default config (15 words)
│   └── words.txt            # Extended word list (80+ words)
├── Makefile                 # Build automation
└── go.mod                   # Go dependencies
```

---

## Configuration

### Command Line Flags

**wordle-client**:
```bash
-mode string      # offline, single, multi (default: prompt)
-server string    # Server URL for online modes (default: http://localhost:8080)
-config string    # Config file for offline mode (default: cfg/config.yaml)
-words string     # Word list file for offline mode (overrides config)
```

**wordle-server**:
```bash
-config string    # Config file (default: cfg/config.yaml)
-words string     # Word list file (overrides config)
-port string      # Server port (default: 8080)
```

### Configuration File

`cfg/config.yaml`:
```yaml
max_rounds: 6

word_list:
  - "CRANE"
  - "SLATE"
  - "ABOUT"
  - "APPLE"
  # ... more words
```

### Word Lists

- **Default**: 15 words in `cfg/config.yaml` (quick games)
- **Extended**: 80+ words in `cfg/words.txt` (more variety)

```bash
# Use extended word list
./bin/wordle-client -mode offline -words cfg/words.txt
./bin/wordle-server -words cfg/words.txt
```

---

## Usage Guide

### Mode Selection

When you run `./bin/wordle-client` without flags:

```
╔════════════════════════════════════╗
║     Welcome to Wordle Game!        ║
╚════════════════════════════════════╝

Please select game mode:
  0. Offline      (standalone, no server required)
  1. Single-Player (online, connect to server)
  2. Multi-Player  (online, race against friends)

Enter choice (0, 1, or 2):
```

### Mode 0: Offline Mode

**Use Case**: Practice, traveling without internet, local play

```bash
# Start offline mode
./bin/wordle-client -mode offline

# Use custom word list
./bin/wordle-client -mode offline -words cfg/words.txt

# Use custom config
./bin/wordle-client -mode offline -config my_config.yaml
```

**Gameplay**:
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
```

**Features**:
- ✅ No network required
- ✅ Instant startup
- ✅ Customizable word lists
- ✅ Local configuration

---

### Mode 1: Single-Player Online

**Use Case**: Secure online play, server-managed games

**Step 1: Start Server**
```bash
# Terminal 1 - Start server
./bin/wordle-server

# With custom config
./bin/wordle-server -words cfg/words.txt -port 8080
```

**Step 2: Start Client**
```bash
# Terminal 2 - Start client
./bin/wordle-client -mode single

# Or specify server
./bin/wordle-client -mode single -server http://localhost:8080
```

**Benefits**:
- ✅ **Anti-cheat**: Client never knows the answer
- ✅ **Server validation**: All guesses validated server-side
- ✅ **Persistent**: Games survive client restart
- ✅ **API-driven**: Can build other clients (web, mobile)

**API Flow**:
```
Client                          Server
  │                               │
  ├─── POST /game/new ──────────→ │  Create game
  │ ← {game_id, max_rounds} ─────┤
  │                               │
  ├─── POST /game/:id/guess ────→ │  Submit guess
  │    {guess: "APPLE"}           │
  │ ← {results: ["O","?",..]} ───┤  Evaluated results
  │                               │
  ├─── GET /game/:id/status ────→ │  Check status
  │ ← {status, history, ...} ────┤
```

---

### Mode 2: Multi-Player Racing

**Use Case**: Competitive racing with friends

#### Game Flow

```
Create/Join Room → Wait in Lobby → Race to Guess → View Rankings
```

**Terminal 1 - Player 1 (Host)**
```bash
./bin/wordle-client -mode multi

# 1. Create Room
Choose: 1 (Create room)
Nickname: Alice
Max players: 4
Room ID: abc123

# 2. Wait for others to join
Players: Alice (YOU, HOST)
Type 'start' to begin...

# 3. Start game
> start

# 4. Race to guess!
[Live progress shows other players]
```

**Terminal 2 - Player 2**
```bash
./bin/wordle-client -mode multi

# 1. Join Room
Choose: 2 (Join room)
Room ID: abc123
Nickname: Bob

# 2. Wait for host to start
Waiting for host...

# 3. Race!
[Game starts when host types 'start']
```

#### Live Progress Display

```
╔══════════════════════════════════════════════════════════╗
║                      🏆 Live Progress                    ║
╠══════════════════════════════════════════════════════════╣
║ 🎮 Alice: Round 2/6 🟩⬜⬜🟨⬜                            ║
║ 🎮 Bob: Round 3/6 ⬜🟨⬜⬜⬜                              ║
║ ✅ Charlie: Round 4/6 (WON!)                             ║
╠══════════════════════════════════════════════════════════╣
║  Game Log:                                               ║
║  [10:23:45] Alice guessed "CRANE"                        ║
║  [10:23:52] Bob guessed "SLATE"                          ║
║  [10:24:01] Charlie guessed "SMILE" - WINNER! 🎉         ║
╠══════════════════════════════════════════════════════════╣
║  Input Area:                                             ║
║  Round 3/6 - Enter your guess: _                         ║
╚══════════════════════════════════════════════════════════╝
```

#### Final Results

```
╔══════════════════════════════════════════════════════════╗
║                      🎮 GAME OVER! 🎮                    ║
╠══════════════════════════════════════════════════════════╣
║                   The Answer was: SMILE                  ║
╠══════════════════════════════════════════════════════════╣
║                     🏆 Final Rankings                    ║
╠══════════════════════════════════════════════════════════╣
║  🥇 ✅ Charlie      4 rounds                              ║
║  🥈 ✅ Alice        5 rounds <- YOU                       ║
║  🥉 ❌ Bob          6 rounds                              ║
╠══════════════════════════════════════════════════════════╣
║              Press ENTER to return to menu               ║
╚══════════════════════════════════════════════════════════╝
```

#### Features

- ✅ **2-8 players** per room
- ✅ **Real-time updates** (millisecond latency)
- ✅ **Host controls** game start
- ✅ **Live rankings** during gameplay
- ✅ **Room browsing** (list available rooms)
- ✅ **Professional UI** with alternate screen buffer

---

## Design & Architecture

### Unified Client Architecture

**Philosophy**: One binary, multiple modes, zero recompilation

```go
// cmd/wordle-client/main.go
func main() {
    mode := promptMode() // or from -mode flag
    
    switch mode {
    case "offline":
        runner := cli.NewRunner(os.Stdin, configPath, wordsPath)
        runner.Run()
    case "single":
        app := client.NewApp(serverURL, os.Stdin)
        app.Run()
    case "multi":
        app := client.NewRoomApp(serverURL, os.Stdin)
        app.Run()
    }
}
```

**Benefits**:
- ✅ Single download for users
- ✅ Consistent UI across modes
- ✅ Shared core game logic
- ✅ Easy mode switching

---

### Multi-Player System Design

#### Room State Machine

```
┌─────────┐
│  INIT   │ Room created by host
└────┬────┘
     │
     ▼
┌─────────┐
│ WAITING │ Players can join, host manages
└────┬────┘
     │ Host starts game
     ▼
┌─────────┐
│ PLAYING │ All players racing simultaneously
└────┬────┘
     │ Winner emerges or all finish
     ▼
┌─────────┐
│FINISHED │ Show rankings and results
└─────────┘
```

#### Room Data Structure

```go
// pkg/server/room.go
type Room struct {
    RoomID      string
    Status      string              // "waiting" | "playing" | "finished"
    MaxPlayers  int
    Players     map[string]*Player  // playerID → Player
    HostID      string              // First player is host
    Game        *game.Game          // Shared game instance
    Version     int                 // For long polling
    updateCond  *sync.Cond          // Broadcast updates
    mu          sync.RWMutex        // Thread-safe access
}
```

**Key Design Decisions**:
1. **Host-Controlled Start**: Prevents accidental game starts
2. **Version-Based Updates**: Efficient incremental polling
3. **Condition Variable Broadcasting**: Notify all waiting clients simultaneously
4. **RWMutex**: Balance read-heavy workload with write protection

---

### Real-Time Updates: Long Polling

**Challenge**: Push real-time updates to multiple clients without WebSockets?

**Solution**: HTTP Long Polling + `sync.Cond`

#### How It Works

```
Client A                    Server                      Client B
   │                          │                            │
   │  GET /room/1/progress    │                            │
   │  ?version=5              │                            │
   ├─────────────────────────→│                            │
   │      (HOLD 30s)          │                            │
   │                          │  ← POST /room/1/guess ────┤
   │                          │     Client B guesses       │
   │                          │                            │
   │                          ├─ room.Version++           │
   │                          ├─ updateCond.Broadcast()   │
   │                          │                            │
   │ ← {version: 6, ...} ─────┤    Wake up ALL waiters    │
   │     (Immediate return)   │                            │
   │                          ├───────────────────────────→│
   │                          │    Return new state        │
```

#### Server Implementation

```go
// pkg/server/server.go
func (s *Server) HandleRoomProgress(c *gin.Context) {
    lastVersion := c.Query("version")
    
    // If already updated, return immediately
    if room.Version > lastVersion {
        c.JSON(200, room.GetProgress())
        return
    }
    
    // Wait for update or timeout (30s)
    ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
    defer cancel()
    
    done := make(chan struct{})
    go func() {
        room.mu.Lock()
        defer room.mu.Unlock()
        defer close(done)
        
        // Wait on condition variable
        for room.Version == lastVersion {
            select {
            case <-ctx.Done():
                return
            default:
            }
            room.updateCond.Wait() // Releases lock, waits, reacquires
        }
    }()
    
    select {
    case <-done:
        c.JSON(200, room.GetProgress())
    case <-ctx.Done():
        room.updateCond.Broadcast() // Wake up goroutine
        <-done
        c.JSON(200, room.GetProgress())
    }
}
```

**Benefits**:
- ✅ **Near-instant updates** (millisecond latency)
- ✅ **No polling spam** (clients wait up to 30s)
- ✅ **HTTP-based** (works through any proxy/firewall)
- ✅ **Broadcasts to all** (via `sync.Cond`)
- ✅ **Auto-reconnect** on timeout

---

### UI/UX: Terminal Design

#### Alternate Screen Buffer

```go
// Enter alternate screen (like vim)
fmt.Print("\033[?1049h\033[H")

// Game plays in isolated screen...

// Exit alternate screen - restore terminal
fmt.Print("\033[?1049l")
```

**Benefits**:
- ✅ Preserves user's terminal history
- ✅ Clean exit - no pollution
- ✅ Professional application feel

#### Real-Time Updates Without Flicker

**Challenge**: Update progress while user is typing

**Solution**: Precise cursor positioning + Unicode-aware padding

```go
// pkg/client/screen_manager.go
type ScreenManager struct {
    logBuffer   []string          // Circular buffer
    numPlayers  int               // Layout tracking
    inputLine   int               // Exact line for input
    inputCol    int               // Exact column for cursor
    mu          sync.Mutex        // Thread-safe
}

func (sm *ScreenManager) UpdateProgress(progress *api.RoomProgressResponse) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    // Save cursor position
    savedLine := sm.inputLine
    savedCol := sm.inputCol
    
    // Update player progress area (at top)
    sm.updatePlayerProgress(progress)
    
    // Restore cursor to exact input position
    fmt.Printf("\033[%d;%dH", savedLine, savedCol)
    // User's typing is never interrupted!
}
```

#### Unicode Width Handling

**Problem**: Emojis (🎮) = 2 display columns, CJK (中) = 2 columns, but Go `len()` = bytes

**Solution**: `github.com/mattn/go-runewidth`

```go
import "github.com/mattn/go-runewidth"

// Perfect alignment
func padOrTruncate(text string, targetWidth int) string {
    currentWidth := runewidth.StringWidth(text)  // Display width, not bytes!
    if currentWidth == targetWidth {
        return text
    }
    if currentWidth > targetWidth {
        return runewidth.Truncate(text, targetWidth, "")
    }
    return text + strings.Repeat(" ", targetWidth-currentWidth)
}

// Example
text := "🎮 Alice中文"
width := runewidth.StringWidth(text)  // = 2+1+5+4 = 12 columns
padded := padOrTruncate(text, 20)     // Adds 8 spaces
```

**Result**: Perfect border alignment across all terminals, with any language!

---

### Design Tradeoffs

#### 1. Long Polling vs. WebSockets

**Chose: Long Polling**

| Aspect | Long Polling | WebSockets |
|--------|--------------|------------|
| Latency | ~100ms (acceptable) | ~10ms (better) |
| Implementation | Simple HTTP | Complex protocol |
| Proxy/Firewall | ✅ Works everywhere | ❌ Often blocked |
| Load Balancing | ✅ Standard HTTP LB | ❌ Sticky sessions required |
| Debugging | ✅ Standard HTTP tools | ❌ Special tools needed |
| Code Complexity | ✅ ~100 LOC | ❌ ~500 LOC |

**Verdict**: For a word game with ~1 guess/5 seconds, 100ms latency is imperceptible. Long polling wins on simplicity and compatibility.

---

#### 2. ANSI Terminal UI vs. TUI Libraries

**Chose: ANSI Escape Codes**

| Aspect | ANSI Codes | Bubble Tea / tview |
|--------|------------|-------------------|
| Dependencies | ✅ Zero | ❌ External |
| Binary Size | ✅ 9MB | ❌ 15-20MB |
| Learning Curve | Medium | High |
| Flexibility | ✅ Full control | Limited to library features |
| Compatibility | ✅ All ANSI terminals | ✅ All ANSI terminals |
| Event Handling | Manual | ✅ Built-in |

**Verdict**: For this project's UI needs (simple progress display + input), ANSI codes provide sufficient functionality without external dependencies. However, for future enhancements (see below), TUI libraries become attractive.

**When to reconsider**: If adding complex features like:
- Scrollable game history
- Interactive menus with arrow keys
- Multi-panel layouts
- Mouse support

---

#### 3. REST API vs. gRPC

**Chose: REST (HTTP + JSON)**

| Aspect | REST | gRPC |
|--------|------|------|
| Ease of Use | ✅ `curl` testing | ❌ Need special tools |
| Browser Support | ✅ Direct from JS | ❌ Requires proxy |
| Debugging | ✅ Human-readable | ❌ Binary protocol |
| Performance | Acceptable | ✅ Faster |
| Schema | ❌ Manual validation | ✅ Protobuf |

**Verdict**: For a game with low QPS (<100), REST's simplicity and debuggability outweigh gRPC's performance benefits.

---

#### 4. Global Input Goroutine vs. Per-Function

**Chose: Global Input Goroutine**

**Problem**: Multiple parts of the app (menu, lobby, game) need to read stdin, but `ReadString('\n')` blocks forever.

**Solution Evolution**:
1. ❌ **Per-function goroutine**: Goroutine leaks when context cancelled but stdin still blocking
2. ❌ **On-demand reading**: Complex channel signaling, hard to reason about
3. ✅ **Single global goroutine**: Started at app init, reads stdin forever, sends to channel

```go
// pkg/client/room_app.go
type RoomApp struct {
    inputChan chan string  // Shared by all app parts
}

func NewRoomApp(...) *RoomApp {
    app := &RoomApp{inputChan: make(chan string, 1)}
    
    // Single global goroutine for entire app lifecycle
    go func() {
        reader := bufio.NewReader(os.Stdin)
        for {
            input, _ := reader.ReadString('\n')
            app.inputChan <- input
        }
    }()
    
    return app
}

// All parts of app just read from channel
func (a *RoomApp) roomLobby() {
    select {
    case input := <-a.inputChan:
        // Handle input
    case <-ctx.Done():
        // Clean exit
    }
}
```

**Benefits**:
- ✅ No goroutine leaks
- ✅ Simple mental model
- ✅ Works with context cancellation
- ❌ Goroutine never exits (acceptable for client app)

---

## Implementation Details

### Core Game Logic

`internal/game/game.go`:
- **Word Validation**: Ensures 5 letters, alphabetic only
- **Scoring Algorithm**: Exact Wordle logic
  1. First pass: Mark exact matches (Hit = 'O')
  2. Second pass: Mark Present for remaining letters ('?')
  3. Correctly handles duplicate letters

### Concurrency Safety

- **Global Input**: Single goroutine, no leaks
- **Screen Manager**: `sync.Mutex` for concurrent updates
- **Room State**: `sync.RWMutex` + `sync.Cond` for broadcasts
- **Context Cancellation**: Graceful shutdown across all goroutines

### API Endpoints

**Single-Player**:
```
POST /game/new           - Create game
POST /game/:id/guess     - Submit guess
GET  /game/:id/status    - Get game state
```

**Multi-Player**:
```
POST   /room/create         - Create room
POST   /room/:id/join       - Join room
POST   /room/:id/start      - Start game (host only)
POST   /room/:id/guess      - Submit guess
GET    /room/:id/progress   - Get live progress (long polling)
GET    /room/list           - List available rooms
```

---

## Future Enhancements

### 1. 🎨 Upgrade to Professional TUI Library

**Recommendation**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)

**Why**:
- **Event-driven architecture**: Better than manual ANSI codes
- **Rich components**: Spinners, progress bars, tables, panes
- **Mouse support**: Click to select rooms
- **Responsive layouts**: Auto-resize on terminal size change
- **Active development**: 20K+ GitHub stars, frequent updates

**Example Enhancement**:
```
Current (ANSI):                  With Bubble Tea:
  Static text display              ✓ Animated spinners
  Manual cursor tracking           ✓ Automatic focus management
  No mouse support                 ✓ Click to join rooms
  Fixed layout                     ✓ Responsive multi-pane layout
  Basic colors                     ✓ Rich color schemes
```

**Implementation Effort**: ~1-2 weeks
- Rewrite screen_manager.go: 3 days
- Add interactive room browser: 2 days
- Animated progress indicators: 1 day
- Testing: 2 days

**Benefit**: More polished, professional feel, easier to add features

---

### 2. 🌐 Web-Based Client (Browser)

**Recommendation**: Build a web client with modern JS framework

**Technology Stack**:
- **Frontend**: React/Vue/Svelte
- **Communication**: WebSocket or SSE (Server-Sent Events)
- **Styling**: Tailwind CSS for Wordle-like tiles

**Benefits**:
- ✅ **Cross-platform**: Windows, Mac, Linux, mobile
- ✅ **Rich UI**: Smooth animations, gradients, shadows
- ✅ **Shareable**: Just send a URL
- ✅ **Analytics**: Track user behavior
- ✅ **Social**: Easy Facebook/Twitter sharing

**Example Features**:
```
┌────────────────────────────────────────┐
│  🎮 Wordle Online - Room: ABC123       │
├────────────────────────────────────────┤
│  👤 Alice (YOU)    Round 2/6   🟩⬜🟨   │
│  👤 Bob            Round 1/6   ⬜⬜⬜   │
│  👤 Charlie        Round 3/6   🟩🟩🟩 ✓ │
├────────────────────────────────────────┤
│  [C][R][A][N][E]  ← Your last guess    │
│  [_][_][_][_][_]  ← Type here          │
├────────────────────────────────────────┤
│  🏆 Leaderboard    📊 Stats    ⚙️      │
└────────────────────────────────────────┘
```

**Implementation Effort**: ~2-3 weeks
- Backend: Upgrade server to WebSocket: 2 days
- Frontend: React app with Wordle UI: 5 days
- Real-time updates: 2 days
- Responsive design: 2 days
- Testing: 2 days

**ROI**: Significantly larger user base (no terminal needed)

---

### 3. 📊 Persistent Statistics & Leaderboard

**Features**:
- Local stats database (`~/.wordle/stats.db`)
- Global leaderboard (server-side)
- Win streaks, guess distribution
- Achievement system

**Tech Stack**:
- **Local**: SQLite or JSON file
- **Server**: PostgreSQL or MySQL
- **API**: REST endpoints for stats

**UI Enhancement**:
```
╔════════════════════════════════════════╗
║     Welcome back, Alice! 👋            ║
╠════════════════════════════════════════╣
║  🏆 Global Rank: #1,234 (↑ 42)         ║
║  📊 Win Rate: 78.5%                    ║
║  🔥 Current Streak: 12 games           ║
║  ⭐ Level 23 - 2,340 XP                ║
╠════════════════════════════════════════╣
║  🆕 New Achievement: Hot Streak!       ║
╠════════════════════════════════════════╣
║  1. Play Game                          ║
║  2. View Statistics 📈                 ║
║  3. Leaderboard 🏆                     ║
║  4. Achievements 🎖️                    ║
╚════════════════════════════════════════╝
```

**Implementation Effort**: ~2-3 weeks

**Benefit**: Increases retention through progression system

---
### Priority Recommendation

**Immediate (High ROI, Low Effort)**:
1. **Daily Challenge** (1 week) - Highest viral potential
2. **Tile Animations** (3-5 days) - Quick visual polish

**Medium-Term (High Impact)**:
3. **Bubble Tea UI** (1-2 weeks) - Better UX foundation
4. **Statistics & Leaderboard** (2-3 weeks) - Retention driver

**Long-Term (Strategic)**:
5. **Web Client** (2-3 weeks) - Expand user base
6. **AI Solver** (1-2 weeks) - Educational value

---

## Development

### Available Make Commands

```bash
# Building
make build              # Build all binaries
make build-server       # Server only
make build-client       # Client only

# Running
make run                # Client (interactive mode selection)
make run-offline        # Client in offline mode
make run-server         # Server

# Testing
make test               # Run all tests
make test-coverage      # With coverage report

# Maintenance
make clean              # Remove build artifacts
make deps               # Download dependencies
make fmt                # Format code
make lint               # Run linter
make help               # Show all commands
```

### Code Structure

- `cmd/`: Entry points
- `pkg/`: Public libraries (importable by others)
- `internal/`: Private libraries (internal use only)
- `cfg/`: Configuration files
- `bin/`: Compiled binaries (generated)

### Testing

```bash
# Run all tests
go test ./...

# Test specific package
go test ./internal/game -v

# With coverage
go test ./... -cover
```

### Contributing

1. Fork the repository
2. Create feature branch
3. Write tests for new features
4. Ensure `make test` passes
5. Format code with `make fmt`
6. Submit pull request

---

## Summary

This Wordle implementation demonstrates:
- ✅ **Clean architecture**: Modular, testable, maintainable
- ✅ **Go best practices**: Concurrency, error handling, packaging
- ✅ **Production-ready**: Graceful shutdown, no leaks, thread-safe
- ✅ **User experience**: Professional terminal UI, real-time updates
- ✅ **Scalability**: Multiple modes, extensible design

**Key Innovations**:
1. **Unified client** supporting 3 modes without recompilation
2. **Long polling with sync.Cond** for efficient real-time updates
3. **Unicode-aware terminal UI** with perfect alignment
4. **Global input goroutine** pattern for clean stdin handling
