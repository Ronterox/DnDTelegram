package main

import (
	"fmt"
	"math/rand/v2"
)

type Item struct {
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

func (c *Character) IsAlive() bool {
	return c.HitPoints > 0
}

func (c *Character) String() string {
	return fmt.Sprintf("%s: (%s)", c.Name, c.Desc)
}

func (c *Character) Roll(dice int) int {
	if c.pitty > 100 {
		c.pitty = 0
		return dice
	}

	roll := rand.IntN(dice) + 1
	if roll <= 3 {
		c.pitty += roll
	}

	return roll
}

type Player struct {
	Character
	ID int64
}

type Game struct {
	Started       bool
	playerIndex   int
	CurrentPlayer *Player
	Players       []Player
}

func (g *Game) FindPlayer(id int64) *Player {
	for _, player := range g.Players {
		if player.ID == id {
			return &player
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
			g.CurrentPlayer = &player
			return true
		}

		g.IncrementPlayerIndex()
	}
	return false
}
