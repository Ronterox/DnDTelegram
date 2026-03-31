package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type WebItem struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type WebStats struct {
	Health       int `json:"health"`
	Mana         int `json:"mana"`
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

type WebPlayer struct {
	Name  string    `json:"name"`
	Info  string    `json:"info"`
	Role  string    `json:"role"`
	Race  string    `json:"race"`
	Items []WebItem `json:"items"`
	Stats WebStats  `json:"stats"`
}

type HistoryEvent struct {
	Action string `json:"action"`
	Map    string `json:"map,omitempty"`
	Damage *int   `json:"damage,omitempty"`
}

type WebGameState struct {
	History []HistoryEvent `json:"history"`
	Players []WebPlayer    `json:"players"`
}

func (g *Game) ToWebFormat() WebGameState {
	players := make([]WebPlayer, len(g.Players))
	for i, p := range g.Players {
		items := make([]WebItem, len(p.Equipment))
		for j, item := range p.Equipment {
			switch it := item.(type) {
			case Weapon:
				items[j] = WebItem{Name: it.Name, Type: "Weapon"}
			case Armor:
				items[j] = WebItem{Name: it.Name, Type: "Armor"}
			case Item:
				items[j] = WebItem{Name: it.Name, Type: it.Type}
			}
		}

		players[i] = WebPlayer{
			Name:  p.Name,
			Info:  p.Desc,
			Role:  p.Class,
			Race:  p.Race,
			Items: items,
			Stats: WebStats{
				Health:       p.HitPoints,
				Mana:         0,
				Strength:     p.Stats["Strength"],
				Dexterity:    p.Stats["Dexterity"],
				Intelligence: p.Stats["Intelligence"],
				Wisdom:       p.Stats["Wisdom"],
				Charisma:     p.Stats["Charisma"],
			},
		}
	}

	return WebGameState{
		History: []HistoryEvent{},
		Players: players,
	}
}

const WebDataPath = "SixSevenStory/public"

func (db *Database) GetWebDataPath(chatID int64) string {
	return fmt.Sprintf("%s/data_%d.json", WebDataPath, chatID)
}

func (db *Database) GetWebURL(chatID int64) string {
	return fmt.Sprintf("http://localhost:5173/?id=%d", chatID)
}

func (db *Database) ExportToWeb(chatID int64, action string, damage int) error {
	game := games[chatID]
	if game == nil {
		return fmt.Errorf("no game found for chatID %d", chatID)
	}

	webState := game.ToWebFormat()

	dataPath := db.GetWebDataPath(chatID)
	var existingEvents []HistoryEvent
	if data, err := os.ReadFile(dataPath); err == nil {
		var existing WebGameState
		if json.Unmarshal(data, &existing) == nil {
			existingEvents = existing.History
		}
	}

	event := HistoryEvent{Action: action}
	if damage > 0 {
		event.Damage = &damage
	}
	existingEvents = append(existingEvents, event)
	webState.History = existingEvents

	jsonData, err := json.MarshalIndent(webState, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	absPath, err := filepath.Abs(dataPath)
	if err != nil {
		return fmt.Errorf("abs path error: %w", err)
	}

	if err := os.WriteFile(absPath, jsonData, 0644); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	fmt.Printf("Exported game state to %s\n", absPath)
	return nil
}

func (db *Database) ExportFullToWeb(chatID int64) error {
	game := games[chatID]
	if game == nil {
		return fmt.Errorf("no game found for chatID %d", chatID)
	}

	webState := game.ToWebFormat()

	dataPath := db.GetWebDataPath(chatID)
	var existingEvents []HistoryEvent
	if data, err := os.ReadFile(dataPath); err == nil {
		var existing WebGameState
		if json.Unmarshal(data, &existing) == nil {
			existingEvents = existing.History
		}
	}

	webState.History = existingEvents

	jsonData, err := json.MarshalIndent(webState, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	absPath, err := filepath.Abs(dataPath)
	if err != nil {
		return fmt.Errorf("abs path error: %w", err)
	}

	if err := os.WriteFile(absPath, jsonData, 0644); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	fmt.Printf("Exported full game state to %s\n", absPath)
	return nil
}
