package main

import (
	"bytes"
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
	Race      string
	Class     string
	Level     int
	Desc      string
	Stats     map[string]int    // Strength, Dexterity, Constitution, Intelligence, Wisdom, and Charisma
	Skills    map[string]string // Acrobatics, Animal Handling, Arcana, Athletics, Deception, History, Insight,
	Armor     int
	HitPoints int
	Equipment []any // Weapon, Armor, Item
	pitty     int
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

		switch typeCheck.Type {
		case "Weapon":
			var w Weapon
			json.Unmarshal(raw, &w)
			c.Equipment = append(c.Equipment, w)
		case "Armor":
			var a Armor
			json.Unmarshal(raw, &a)
			c.Equipment = append(c.Equipment, a)
		default:
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

/*
This may be a good way of outputting information about the character

	┃  Let me think about what I know:                                                    Context
	┃  - Ricardo: A mage (but with 17 Strength, which is interesting)                     7,325 tokens
	┃  - Stats: DEX 8, CON 13, INT 9, WIS 15, CHA 8, STR 17                               4% used
*/
func (c *Character) toString() string {
	var output bytes.Buffer

	output.WriteString(fmt.Sprintf("%s: %s\n", c.Name, c.Desc))
	for key, value := range c.Stats {
		output.WriteString(fmt.Sprintf("%s: %d\n", key, value))
	}

	return output.String()
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

type PlayerState int

const (
	StateNone PlayerState = iota
	StateSettingUp
	StatePaused
	StateReady
)

type Player struct {
	Character
	ID    int64
	State PlayerState
}

func NewPlayer(userID int64, name string, description string) *Player {
	/*
		Class,Hit Die,Max Value (Level 1)
		"Wizard, Sorcerer",d6,6
		"Bard, Cleric, Druid, Monk, Rogue, Warlock",d8,8
		"Fighter, Paladin, Ranger",d10,10
		Barbarian,d12,12

		Every level, plus constitution bonus
	*/

	return &Player{
		ID: userID,
		Character: Character{
			Name: name,
			// TODO: AI get race
			// Proficiencies in equipment/tools, Cool native Skills, Languages, weapons armor,
			// Lifespan, size, speed
			// Fixed roll bonuses to certain rolls like Dexterity for Elves +2, Dwarves constitution +2
			Race: "Human",
			Desc: description,
			// TODO: AI get class
			Class: "None",
			// TODO: Roll for hitpoints
			HitPoints: roll(10) + 5,
			Level:     3,
			// TODO: Armor depends on equipment given by AI
			// Rule if is role greater equal than AC then hit
			// 1 always fails
			// NATURAL 20 always hits, even is AC is 30 (19+something fails)
			// Some skills always hit, and some the enemy has to dodge
			// Always rolls 20 + bonuses
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
	CurrentPlayer *Player
	Players       []*Player
	SessionID     string
	Started       bool
	playerIndex   int
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
