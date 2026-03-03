package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const token = ""
const apiBase = "https://api.telegram.org/bot" + token

type Update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Text string `json:"text"`
		Chat struct {
			ID   int64  `json:"id"`
			Type string `json:"type"` // private, group, supergroup, channel
		} `json:"chat"`
		User struct {
			ID        int64  `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
		} `json:"from"`
	} `json:"message"`
}

type UpdateResult struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Character struct {
	Name string
	Desc string
}

type Player struct {
	Character
}

type Game struct {
	CurrentIndex int
	Players      map[int64]Player
}

func (g *Game) FindPlayer(id int64) *Player {
	for pid, player := range g.Players {
		if pid == id {
			return &player
		}
	}
	return nil
}

func getUpdates(offset int) (UpdateResult, error) {
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=30", apiBase, offset)

	resp, err := http.Get(url)
	if err != nil {
		return UpdateResult{}, err
	}

	var data UpdateResult
	json.NewDecoder(resp.Body).Decode(&data)
	resp.Body.Close()

	if !data.Ok {
		return UpdateResult{}, fmt.Errorf("Update result not ok: %v", data.Result)
	}

	return data, nil
}

func sendText(chatID int64, text string) {
	url := apiBase + "/sendMessage"
	fmt.Printf("Sending text to chat %d: %s\n", chatID, text)
	jsonData, _ := json.Marshal(map[string]any{"chat_id": chatID, "text": text})
	http.Post(url, "application/json", bytes.NewBuffer(jsonData))
}

func main() {
	games := make(map[int64]*Game)
	offset := 0

	fmt.Println("Bot started... Press Ctrl+C to stop.")

	for {
		updates, err := getUpdates(offset)
		if err != nil {
			fmt.Println("Error fetching updates:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, update := range updates.Result {
			message := update.Message
			chatID := message.Chat.ID
			text := message.Text

			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}

			fmt.Printf("Received: %s\n", text)

			if strings.HasPrefix(text, "/join") {
				if games[chatID] == nil {
					games[chatID] = &Game{CurrentIndex: 0, Players: make(map[int64]Player)}
					sendText(chatID, "Campaña grupal iniciada!")
				}

				if player := games[chatID].FindPlayer(message.User.ID); player != nil {
					sendText(chatID, fmt.Sprintf("¡Ya eres un jugador! %+v", player))
					continue
				}

				if description := strings.TrimSpace(strings.Replace(text, "/join", "", 1)); description != "" {
					newPlayer := Player{
						Character: Character{
							Name: message.User.FirstName,
							Desc: description,
						},
					}
					games[chatID].Players[message.User.ID] = newPlayer
					sendText(chatID, fmt.Sprintf("¡Te has unido a la campaña, bienvenido %s!", newPlayer.Name))
				} else {
					sendText(chatID, "Escribe una descripción para tu personaje /join <descripcion>")
				}
			} else if strings.HasPrefix(text, "/whoami") {
				if games[chatID] == nil {
					sendText(chatID, "No hay ninguna partida en curso, empieza una uniendote con /join")
					continue
				}

				if player := games[chatID].FindPlayer(message.User.ID); player != nil {
					sendText(chatID, fmt.Sprintf("Eres %+v", player))
				} else {
					sendText(chatID, "No te has unido a la partida, unete con /join")
				}
			}
		}
	}
}
