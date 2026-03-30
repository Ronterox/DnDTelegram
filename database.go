package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DBBaseURL = "http://localhost:3001"

type Database struct {
	client  *http.Client
	baseURL string
}

func NewDatabase() *Database {
	return &Database{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: DBBaseURL,
	}
}

func (db *Database) SaveGame(chatID int64, game *Game) error {
	gameData := map[string]interface{}{
		"session_id":     fmt.Sprintf("%d", chatID),
		"current_player": game.CurrentPlayer,
		"players":        game.Players,
		"session_id_dm":  game.SessionID,
		"started":        game.Started,
		"player_index":   game.PlayerIndex,
	}

	jsonData, err := json.Marshal(gameData)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	resp, err := db.client.Post(db.baseURL+"/api/games", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("save error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("save failed: %s", string(body))
	}

	return nil
}

func (db *Database) LoadGame(chatID int64) (*Game, error) {
	sessionID := fmt.Sprintf("%d", chatID)
	resp, err := db.client.Get(db.baseURL + "/api/games/" + sessionID)
	if err != nil {
		return nil, fmt.Errorf("load error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("load failed: status %d", resp.StatusCode)
	}

	var gameData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&gameData); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	game := &Game{
		PlayerIndex: int(gameData["player_index"].(float64)),
		Started:     gameData["started"].(bool),
	}

	if v, ok := gameData["session_id_dm"]; ok && v != nil {
		game.SessionID = v.(string)
	}

	if v, ok := gameData["players"]; ok && v != nil {
		if data, err := json.Marshal(v); err == nil {
			json.Unmarshal(data, &game.Players)
		}
	}

	if v, ok := gameData["current_player"]; ok && v != nil {
		if data, err := json.Marshal(v); err == nil {
			json.Unmarshal(data, &game.CurrentPlayer)
		}
	}

	return game, nil
}

func (db *Database) LoadAllGames() (map[int64]*Game, error) {
	resp, err := db.client.Get(db.baseURL + "/api/games")
	if err != nil {
		return nil, fmt.Errorf("load all error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("load all failed: status %d", resp.StatusCode)
	}

	var gamesData []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&gamesData); err != nil {
		return nil, fmt.Errorf("decode all error: %w", err)
	}

	games := make(map[int64]*Game)
	for _, gameData := range gamesData {
		sessionID, ok := gameData["session_id"].(string)
		if !ok {
			continue
		}

		var chatID int64
		fmt.Sscanf(sessionID, "%d", &chatID)

		game := &Game{
			PlayerIndex: int(gameData["player_index"].(float64)),
			Started:     gameData["started"].(bool),
		}

		if v, ok := gameData["session_id_dm"]; ok && v != nil {
			game.SessionID = v.(string)
		}

		if v, ok := gameData["players"]; ok && v != nil {
			if data, err := json.Marshal(v); err == nil {
				json.Unmarshal(data, &game.Players)
			}
		}

		if v, ok := gameData["current_player"]; ok && v != nil {
			if data, err := json.Marshal(v); err == nil {
				json.Unmarshal(data, &game.CurrentPlayer)
			}
		}

		games[chatID] = game
	}

	return games, nil
}
