package main

type Character struct {
	Name string
	Desc string
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

func (g *Game) SetNextPlayer() {
	g.playerIndex = (g.playerIndex + 1) % len(g.Players)
	for i, player := range g.Players {
		if i == g.playerIndex {
			g.CurrentPlayer = &player
			return
		}
	}
}
