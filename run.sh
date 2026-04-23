#!/bin/bash

set -e

. .env

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Starting all services..."

echo "Starting Redis..."
redis-server --daemonize yes 2>/dev/null || true

echo "Starting D&D OpenCode server (port 3000)..."
(cd dnd && bun run opencode) &
DND_PID=$!

echo "Starting D&D API server (port 3000)..."
(cd dnd && npm start) &
API_PID=$!

echo "Starting Game API server (port 3001)..."
bun run game-api.js &
GAME_PID=$!

sleep 2

echo "Starting Bot..."
go run . &
BOT_PID=$!

echo "Starting web..."
cd SixSevenStory && bun dev &
WEB_PID=$!

echo ""
echo "All services started!"
echo "  - D&D OpenCode: port 3000"
echo "  - D&D API:      port 3000"
echo "  - Game API:     port 3001"
echo "  - Bot:          Telegram"
echo ""
echo "PIDs: D&D=$DND_PID API=$API_PID Game=$GAME_PID Bot=$BOT_PID"
echo "Press Ctrl+C to stop all services"

trap "kill $DND_PID $API_PID $GAME_PID $BOT_PID $WEB_PID 2>/dev/null" EXIT

wait
