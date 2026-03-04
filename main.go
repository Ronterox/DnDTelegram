package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const ERR_GAME_NIL = "No hay ninguna campaña en curso, empieza una uniendote con /join"
const ERR_GAME_STARTED = "La campaña ya ha comenzado, ¡No puedes unirte!"
const ERR_GAME_TURN = "¡No es tu turno aun!"

const MSG_GAME_STARTED = "¡La campaña ha comenzado!"
const MSG_GAME_ENDED = "¡La campaña ha terminado!"
const MSG_HELP = `

/start - Empieza la campaña
/join <descripcion de tu personaje> - Unirse a la campaña
/whoami - Información sobre ti
/roll - Lanza un dado

Nota: <> se utiliza para señalar la ayuda no necesitas ponerlos literalmente
`

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

			isCommand := func(prefix string) bool {
				return strings.HasPrefix(text, prefix)
			}

			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}

			fmt.Printf("Received: %s\n", text)

			if isCommand("/start") {
				if game == nil {
					api.sendText(chatID, ERR_GAME_NIL)
					continue
				}

				game.Started = true
				game.SetNextPlayer()

				api.sendText(chatID, "¡La campaña ha comenzado!")
				api.sendText(chatID, fmt.Sprintf("%s te encuentras en el baño haciendo kk, que quieres hacer?", game.CurrentPlayer.Name))
			} else if isCommand("/join") {
				if game == nil {
					games[chatID] = &Game{playerIndex: -1, Players: []Player{}}
					game = games[chatID]
					api.sendText(chatID, MSG_GAME_STARTED)
				}

				if player := game.FindPlayer(message.User.ID); player != nil {
					api.sendText(chatID, fmt.Sprintf("¡Ya eres un jugador! %+v", player))
					continue
				}

				if game.Started {
					api.sendText(chatID, ERR_GAME_STARTED)
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
			} else if isCommand("/whoami") {
				if game == nil {
					api.sendText(chatID, ERR_GAME_NIL)
					continue
				}

				if player := game.FindPlayer(message.User.ID); player != nil {
					api.sendText(chatID, fmt.Sprintf("Eres %+v", player))
				} else {
					api.sendText(chatID, "No te has unido a la campaña, unete con /join")
				}
			} else if isCommand("/roll") {
				if game == nil {
					api.sendText(chatID, ERR_GAME_NIL)
					continue
				}

				if game.CurrentPlayer != nil && game.CurrentPlayer.ID == message.User.ID {
					for _, dice := range []int{4, 6, 8, 10, 12, 20} {
						api.sendText(chatID, fmt.Sprintf("D%d: %d", dice, game.CurrentPlayer.Roll(dice)))
					}
				} else {
					api.sendText(chatID, ERR_GAME_TURN)
				}
			} else if isCommand("/help") {
				api.sendText(chatID, "¡Bienvenido a DnD!")
				api.sendText(chatID, MSG_HELP)
			} else if isCommand("/") {
				api.sendText(chatID, "¡Error, estos son los comandos disponibles!")
				api.sendText(chatID, MSG_HELP)
			} else if game != nil && game.CurrentPlayer != nil && game.CurrentPlayer.ID == message.User.ID {
				api.sendText(chatID, fmt.Sprintf("%s se ha hecho kk encima, y ha muerto...", game.CurrentPlayer.Name))
				if !game.SetNextPlayer() {
					game.Started = false
					api.sendText(chatID, MSG_GAME_ENDED)
				}
			}
		}
	}
}
