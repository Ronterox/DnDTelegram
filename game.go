package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
)

type Item struct {
	Type       string // Weapon, Armor, Item
	Name       string
	Desc       string
	Weight     float64
	Properties []string
}

type Weapon struct {
	Item
	Damage     string // 1d8
	DamageType string // Slashing
}

type Armor struct {
	Item
	BaseAC      int
	MinStrength int
}

type Character struct {
	Name      string
	Class     string
	Level     int
	Desc      string
	Stats     map[string]int // Strength, Dexterity, Constitution, Intelligence, Wisdom, and Charisma
	Armor     int
	HitPoints int
	pitty     int
	Equipment []any // Weapon, Armor, Item
}

func (c *Character) UnmarshalJSON(data []byte) error {
	// Temporary type to avoid infinite recursion
	type Alias Character
	aux := &struct {
		Equipment []json.RawMessage `json:"equipment"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	for _, raw := range aux.Equipment {
		var typeCheck struct {
			Type string `json:"type"`
		}
		json.Unmarshal(raw, &typeCheck)

		if typeCheck.Type == "Weapon" {
			var w Weapon
			json.Unmarshal(raw, &w)
			c.Equipment = append(c.Equipment, w)
		} else if typeCheck.Type == "Armor" {
			var a Armor
			json.Unmarshal(raw, &a)
			c.Equipment = append(c.Equipment, a)
		} else {
			var i Item
			json.Unmarshal(raw, &i)
			c.Equipment = append(c.Equipment, i)
		}
	}
	return nil
}

func (c *Character) IsAlive() bool {
	return c.HitPoints > 0
}

func (c *Character) String() string {
	return fmt.Sprintf("%s: (%s)", c.Name, c.Desc)
}

func (c *Character) RollModifier(stat string) int {
	value := c.Stats[stat]
	if value < 10 {
		return 0
	}
	return (value - 10) / 2
}

func (c *Character) Roll(dice int) int {
	if c.pitty > 100 {
		c.pitty = 0
		return dice
	}

	roll := roll(dice)
	if roll <= 3 {
		c.pitty += (4 - roll)
	} else {
		c.pitty++
	}

	return roll
}

func roll(dice int) int {
	return rand.IntN(dice) + 1
}

type Player struct {
	Character
	ID    int64
	Ready bool
}

func NewPlayer(userID int64, name string, description string) *Player {
	return &Player{
		ID: userID,
		Character: Character{
			Name:  name,
			Desc:  description,
			Class: "None",
			Level: 3,
			Stats: map[string]int{
				"Strength":     0,
				"Dexterity":    0,
				"Constitution": 0,
				"Intelligence": 0,
				"Wisdom":       0,
				"Charisma":     0,
			},
		},
	}
}

type Game struct {
	Started       bool
	playerIndex   int
	CurrentPlayer *Player
	Players       []*Player
}

func (g *Game) FindPlayer(id int64) *Player {
	for _, player := range g.Players {
		if player.ID == id {
			return player
		}
	}
	return nil
}

func (g *Game) IncrementPlayerIndex() {
	g.playerIndex = (g.playerIndex + 1) % len(g.Players)
}

func (g *Game) SetNextPlayer() bool {
	g.IncrementPlayerIndex()
	fmt.Printf("Looking for index %d", g.playerIndex)
	for i, player := range g.Players {
		if i != g.playerIndex {
			continue
		}

		if player.Character.IsAlive() {
			g.CurrentPlayer = player
			return true
		}

		g.IncrementPlayerIndex()
	}
	return false
}
