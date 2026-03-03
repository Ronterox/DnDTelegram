package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	api := NewAPI(os.Getenv("TOKEN"))

	if len(api.token) != 46 {
		fmt.Println("Invalid token. Set TOKEN environment variable.")
		return
	}

	games := make(map[int64]*Game)
	offset := 0

	fmt.Printf("Bot started with token ending in %s... Press Ctrl+C to stop.", api.token[len(api.token)-8:])

	for {
		updates, err := api.getUpdates(offset)
		if err != nil {
			fmt.Println("Error fetching updates:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, update := range updates.Result {
			message := update.Message
			chatID := message.Chat.ID
			text := message.Text
			game := games[chatID]

			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}

			fmt.Printf("Received: %s\n", text)

			if strings.HasPrefix(text, "/start") {
				if game == nil {
					api.sendText(chatID, "No hay ninguna partida en curso, empieza una uniendote con /join")
					continue
				}

				game.Started = true
				game.SetNextPlayer()

				api.sendText(chatID, "¡La partida ha comenzado!")
				api.sendText(chatID, fmt.Sprintf("%s te encuentras en el baño haciendo kk, que quieres hacer?", game.CurrentPlayer.Name))
			} else if strings.HasPrefix(text, "/join") {
				if game == nil {
					games[chatID] = &Game{playerIndex: -1, Players: []Player{}}
					game = games[chatID]
					api.sendText(chatID, "Campaña grupal iniciada!")
				}

				if player := game.FindPlayer(message.User.ID); player != nil {
					api.sendText(chatID, fmt.Sprintf("¡Ya eres un jugador! %+v", player))
					continue
				}

				if description := strings.TrimSpace(strings.Replace(text, "/join", "", 1)); description != "" {
					newPlayer := Player{
						ID: message.User.ID,
						Character: Character{
							Name: message.User.FirstName,
							Desc: description,
						},
					}
					game.Players = append(game.Players, newPlayer)
					api.sendText(chatID, fmt.Sprintf("¡Te has unido a la campaña, bienvenido %s!", newPlayer.Name))
				} else {
					api.sendText(chatID, "Escribe una descripción para tu personaje /join <descripcion>")
				}
			} else if strings.HasPrefix(text, "/whoami") {
				if game == nil {
					api.sendText(chatID, "No hay ninguna partida en curso, empieza una uniendote con /join")
					continue
				}

				if player := game.FindPlayer(message.User.ID); player != nil {
					api.sendText(chatID, fmt.Sprintf("Eres %+v", player))
				} else {
					api.sendText(chatID, "No te has unido a la partida, unete con /join")
				}
			} else if game != nil && game.CurrentPlayer != nil && game.CurrentPlayer.ID == message.User.ID {
				api.sendText(chatID, fmt.Sprintf("%s se ha hecho kk encima, y ha muerto...", game.CurrentPlayer.Name))
				api.sendText(chatID, "¡La partida ha terminado!")
			}
		}
	}
}
