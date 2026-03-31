# Telegram Ultimate Redirectioner

A Telegram bot-powered D&D 5th Edition game system with AI Dungeon Master, real-time web viewer, and session-based game state management.

## System Description and Theme

This project is a **tabletop RPG game platform** that combines:
- **Telegram Bot Interface** - Players interact with the game through a Telegram bot
- **AI-Powered Dungeon Master** - OpenCode AI generates dynamic narratives and manages game logic
- **Web Viewer** - Real-time dashboard (SixSevenStory) for viewing game state
- **Persistent Game State** - SQLite database for saving campaigns

### Theme: Medieval Fantasy

The system is designed for **Spanish-language D&D 5e campaigns** with a medieval fantasy aesthetic:
- Narrative-driven gameplay with atmospheric storytelling
- Character creation through interactive Telegram conversation
- Turn-based combat with stat rolls and skill checks

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Telegram Users                                  │
│                  (Interact via buttons & commands)                           │
└────────────────────────────────┬────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Go Telegram Bot (main.go)                          │
│  • Handles commands (/start, /join, /ready, /roll, /pause)                  │
│  • Manages game state (players, turns, inventory)                           │
│  • Interfaces with D&D API and Database                                     │
└────────────────────────────────┬────────────────────────────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  D&D API (dnd)  │    │  SixSevenStory  │    │    Database     │
│  (Bun/Express)  │    │   (React/Vite)  │    │   (SQLite via   │
│                 │    │                 │    │   REST API)     │
│ /api/chat       │    │ Real-time game  │    │                 │
│ /api/init       │    │ state viewer    │    │ Game state      │
│ /api/health     │    │                 │    │ persistence     │
└────────┬────────┘    └────────┬────────┘    └────────┬────────┘
         │                       │                       │
         ▼                       │                       │
┌─────────────────┐              │                       │
│  OpenCode AI    │              │                       │
│  (big-pickle)   │              │                       │
│                 │              │                       │
│ Dungeon Master  │              │                       │
│ Narrative Gen   │              │                       │
└─────────────────┘              │                       │
                                 ▼                       │
                    ┌─────────────────────────┐          │
                    │  JSON Export (web_      │          │
                    │  export.go)             │          │
                    │                         │          │
                    │  /public/data_{chatId}. │          │
                    │  json                    │          │
                    └─────────────────────────┘          │
                                 │                       │
                                 └───────────┬───────────┘
                                             ▼
                              ┌─────────────────────────┐
                              │   Game State JSON      │
                              │   (auto-refresh 5s)     │
                              └─────────────────────────┘
```

### Components

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Telegram Bot** | Go | Main game controller, message handling |
| **D&D API** | Bun + Express | OpenCode AI integration for DM narrative |
| **SixSevenStory** | React 19 + Vite | Web dashboard for game viewing |
| **Database** | REST API (SQLite) | Persistent game state storage |
| **Web Export** | Go | JSON file generation for viewer |

---

## Installation and Configuration

### Prerequisites

- **Go 1.21+** - For the Telegram bot
- **Bun** - For D&D API server
- **Node.js 18+** - For SixSevenStory (optional, uses Vite)
- **SQLite** - For game state storage

### 1. Telegram Bot Setup

1. Create a bot via [@BotFather](https://t.me/BotFather) on Telegram
2. Copy the bot token

### 2. Database API Setup (dnd/)

```bash
cd dnd
bun install
```

### 3. SixSevenStory Setup (optional)

```bash
cd SixSevenStory
bun install
```

---

## Required Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `TOKEN` | Telegram Bot API token | **Yes** |

### Additional Configuration

The system uses hardcoded defaults:

| Setting | Default Value |
|---------|---------------|
| D&D API Base URL | `http://localhost:3000` |
| Database API Base URL | `http://localhost:3001` |
| OpenCode Server | `http://localhost:4096` |
| Web Export Path | `./SixSevenStory/public/` |

---

## REST API Documentation

### D&D API (Port 3000)

#### `POST /api/init`

Initialize a new Dungeon Master session.

**Request:**
```json
{}
```

**Response:**
```json
{
  "success": true,
  "sessionId": "abc123..."
}
```

---

#### `POST /api/chat`

Send a message to the Dungeon Master.

**Request:**
```json
{
  "message": "I want to attack the goblin",
  "sessionId": "abc123..."
}
```

**Response:**
```json
{
  "response": {
    "narrative": "You swing your sword at the goblin..."
  },
  "sessionId": "abc123",
  "type": "structured"
}
```

---

#### `GET /api/session/:id/messages`

Get message history for a session.

**Response:**
```json
{
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "..."},
    {"role": "assistant", "content": "..."}
  ]
}
```

---

#### `GET /api/health`

Check API and OpenCode server health.

**Response:**
```json
{
  "status": "ok",
  "opencode": "connected"
}
```

---

## Usage Examples

### Starting a New Campaign

1. **First player** sends `/join <character description>` to the bot
2. Bot creates a new game session and DM session
3. **Each player** joins with `/join <description>`
4. **Each player** runs `/ready` to finalize character (roll stats)
5. **First player** runs `/start` to begin the campaign

### During Gameplay

| Command | Action |
|---------|--------|
| `/roll` | Roll all dice (D4, D6, D8, D10, D12, D20) |
| `/pause` | Pause your turn (skip actions) |
| `/whoami` | View your character info |
| `/help` | Show available commands |

### Button Controls

Inline buttons appear with each DM message:
- **Inventario** - View your inventory
- **Stats** - View character stats
- **Skills** - View character skills
- **Roll** - Roll dice for current turn
- **Pass** - Pass your turn to next player
- **Pause** - Pause/unpause your turn

### Web Viewer

After joining, players receive a link to view the game:
```
🎮 Ver tu partida: http://localhost:5173/?id=123456789
```

The viewer shows:
- Player cards with stats and HP/Mana
- Inventory per player
- Game history timeline
- ASCII map (if applicable)

---

## Running the System

### Terminal 1: D&D API Server
```bash
cd dnd
bun run opencode  # Start OpenCode server (port 4096)
bun start        # Start D&D API (port 3000)
```

### Terminal 2: SixSevenStory (Optional)
```bash
cd SixSevenStory
bun run dev  # Start web server (port 5173)
```

### Terminal 3: Telegram Bot
```bash
export TOKEN="your_bot_token_here"
go run main.go
```

### Quick Start with run.sh

For convenience, use the `run.sh` script to start all services at once:

```bash
chmod +x run.sh
./run.sh
```

**Environment Variables:**
| Variable | Description |
|----------|-------------|
| `TOKEN` | Telegram Bot API token (required) |

The script starts all services in the correct order:
- Redis (optional)
- D&D OpenCode server (port 4096)
- D&D API server (port 3000)
- Game API server (port 3001)
- Telegram Bot
- SixSevenStory web (port 5173)

Press `Ctrl+C` to stop all services.

### Docker Deployment

#### Single Container (Dockerfile)

Build and run everything in one container:

```bash
# Build the image
docker build -t tg-dnd .

# Run the container
docker run -d \
  -e TOKEN="your_bot_token_here" \
  -p 3000:3000 \
  -p 3001:3001 \
  -p 4096:4096 \
  -p 5173:5173 \
  --name tg-dnd tg-dnd
```

#### Multi-Container (Docker Compose)

For better service isolation, use docker-compose:

```bash
# Build and start all services
TOKEN="your_bot_token_here" docker-compose up --build

# Run in detached mode
TOKEN="your_bot_token_here" docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

**Docker Services:**
| Service | Port | Description |
|---------|------|-------------|
| `redis` | 6379 | Game state storage |
| `dnd-opencode` | 4096 | OpenCode AI server |
| `dnd-api` | 3000 | D&D API (chat, init) |
| `game-api` | 3001 | Game state API |
| `sixsevenstory` | 5173 | Web viewer |
| `bot` | - | Telegram bot |

**Docker Environment Variables:**
| Variable | Description |
|----------|-------------|
| `TOKEN` | Telegram Bot API token (required) |
| `REDIS_HOST` | Redis host (default: redis in compose) |

---

## Data Flow Example

```
User: "I attack the orc"
   │
   ▼
Telegram Bot receives message
   │
   ▼
Checks if game started, player is current turn
   │
   ▼
Sends to D&D API: POST /api/chat
   │
   ▼
D&D API forwards to OpenCode AI (big-pickle model)
   │
   ▼
DM generates narrative response
   │
   ▼
Telegram Bot sends response + buttons to user
   │
   ▼
Web Export saves game state to JSON
   │
   ▼
SixSevenStory viewer auto-refreshes (5s polling)
```