package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const ALBERTO_ID = 1426815752

const ERR_GAME_NIL = "No hay ninguna campaña en curso, empieza una uniendote con /join"
const ERR_GAME_STARTED = "La campaña ya ha comenzado, ¡No puedes unirte!"
const ERR_GAME_TURN = "¡No es tu turno aun!"
const ERR_READY = "¡%s no está listo para jugar utiliza /ready para marcarlo!"
const ERR_JOINED = "No te has unido a la campaña, unete con /join"

const MSG_GAME_STARTED = "¡Se ha creado una campaña nueva!"
const MSG_GAME_ENDED = "¡La campaña ha terminado!"
const MSG_JOINED = `
¡Te has unido a la campaña, bienvenido %s!

Cuando estés listo para jugar y terminar de configurar tu personaje, utiliza /ready.
`
const MSG_HELP = `

/start - Empieza la campaña
/join <descripcion de tu personaje> - Unirse a la campaña
/whoami - Información sobre ti
/roll - Lanza un dado

Comandos de Creación de Jugadores:

/ready - Marca a un jugador como listo para jugar

Nota: <> se utiliza para señalar la ayuda no necesitas ponerlos literalmente
`

func failIfFalse(condition bool, msg string) {
	if !condition {
		fmt.Println(msg)
		panic(msg)
	}
}

func main() {
	api := NewAPI(os.Getenv("TOKEN"))

	if len(api.token) != 46 {
		fmt.Println("Invalid token. Set TOKEN environment variable.")
		return
	}

	settingUp := make(map[int64]int)
	games := make(map[int64]*Game)

	offset := 0

	fmt.Printf("Bot started with token ending in %s... Press Ctrl+C to stop.\n", api.token[len(api.token)-8:])

	for {
		updates, err := api.getUpdates(offset)
		if err != nil {
			fmt.Println("Error fetching updates:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, update := range updates.Result {
			message := update.Message
			callback := update.CallbackQuery

			chatID := message.Chat.ID
			userID := message.User.ID

			text := message.Text
			game := games[chatID]

			isCommand := func(prefix string) bool {
				return strings.HasPrefix(text, prefix)
			}

			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}

			fmt.Printf("Received: (text: %s), (button: %s)\n", text, update.CallbackQuery.Data)

			if update.CallbackQuery.Data != "" {
				api.editMessage(callback.Message.Chat.ID, callback.Message.ID, callback.Data, [][]InlineKeyboardButton{})
				fmt.Printf("Edited message %d\n", callback.Message.ID)
				continue
			}

			if isCommand("/start") {
				fmt.Println("Running start")
				if game == nil {
					api.sendText(chatID, ERR_GAME_NIL)
					continue
				}

				for _, player := range game.Players {
					if !player.Ready {
						api.sendText(chatID, fmt.Sprintf(ERR_READY, player.Name))
						continue
					}
				}

				failIfFalse(len(game.Players) > 0, "No hay jugadores en la partida")

				game.Started = true
				game.SetNextPlayer()

				failIfFalse(game.CurrentPlayer != nil, "No hay jugador actual")

				api.sendText(chatID, "¡La campaña ha comenzado!")
				// api.sendText(chatID, fmt.Sprintf("%s te encuentras en el baño haciendo kk, que quieres hacer?", game.CurrentPlayer.Name))
			} else if isCommand("/join") {
				fmt.Println("Running join")
				if game == nil {
					games[chatID] = &Game{playerIndex: -1, Players: []Player{}}
					game = games[chatID]
					api.sendText(chatID, MSG_GAME_STARTED)
				}

				if player := game.FindPlayer(userID); player != nil {
					api.sendText(chatID, fmt.Sprintf("¡Ya eres un jugador! %+v", player))
					continue
				}

				if game.Started {
					api.sendText(chatID, ERR_GAME_STARTED)
					continue
				}

				if description := strings.TrimSpace(strings.Replace(text, "/join", "", 1)); description != "" {
					newPlayer := Player{
						ID: userID,
						Character: Character{
							Name: message.User.FirstName,
							Desc: description,
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
					game.Players = append(game.Players, newPlayer)
					api.sendText(chatID, fmt.Sprintf(MSG_JOINED, newPlayer.Name))
				} else {
					api.sendText(chatID, "Escribe una descripción para tu personaje /join <descripcion>")
				}
			} else if isCommand("/whoami") {
				fmt.Println("Running whoami")
				if game == nil {
					api.sendText(chatID, ERR_GAME_NIL)
					continue
				}

				if player := game.FindPlayer(userID); player != nil {
					api.sendText(chatID, fmt.Sprintf("Eres %+v", player))
				} else {
					api.sendText(chatID, ERR_JOINED)
				}
			} else if isCommand("/roll") {
				fmt.Println("Running roll")
				if game == nil {
					api.sendText(chatID, ERR_GAME_NIL)
					continue
				}

				if game.CurrentPlayer != nil && game.CurrentPlayer.ID == userID {
					for _, dice := range []int{4, 6, 8, 10, 12, 20} {
						api.sendText(chatID, fmt.Sprintf("D%d: %d", dice, game.CurrentPlayer.Roll(dice)))
					}
				} else {
					api.sendText(chatID, ERR_GAME_TURN)
				}
			} else if isCommand("/ready") {
				fmt.Println("Running ready")
				if game == nil {
					api.sendText(chatID, ERR_GAME_NIL)
					continue
				}

				if settingUp[userID] != 0 {
					api.sendText(chatID, "Termina de configurar tu personaje")
					continue
				}

				if player := game.FindPlayer(userID); player != nil {
					if player.Ready {
						api.sendText(chatID, fmt.Sprintf("¡%s ya está listo para jugar!", player.Name))
						continue
					}

					settingUp[chatID] = 1
					api.sendText(chatID, "Veamos que destino depara para tu fisico y fuerza mental...")

					go func() {
						for key := range player.Character.Stats {
							api.sendText(chatID, fmt.Sprintf("En cuanto a tu %s...", key))
							time.Sleep(time.Second * 5)

							rolls := make([]int, 4)
							smallest := 100
							total := 0

							for i := range 4 {
								rolls[i] = player.Roll(6)
								smallest = min(rolls[i], smallest)
								total += rolls[i]

								api.sendText(chatID, fmt.Sprintf("%d", rolls[i]))
							}

							result := total - smallest
							player.Character.Stats[key] = result

							if result > 16 {
								api.sendText(chatID, fmt.Sprintf("¡Tu %s es de %d, en verdad que eres habilidoso!", key, result))
							} else if result > 12 {
								api.sendText(chatID, fmt.Sprintf("¡Tu %s es de %d, un valor decente!", key, result))
							} else if result > 8 {
								api.sendText(chatID, fmt.Sprintf("Tu %s es de %d, un poco debajo del promedio...", key, result))
							} else {
								api.sendText(chatID, fmt.Sprintf("Tu %s es de %d... oof, que mala suerte no?", key, result))
							}
						}

						api.sendButtons(chatID, "Estas satisfecho con este resultado?", [][]InlineKeyboardButton{
							{
								{
									Text:         "Si",
									CallbackData: "ready",
								},
								{
									Text:         "No",
									CallbackData: "configure",
								},
							},
						})
					}()
				} else {
					api.sendText(chatID, ERR_JOINED)
				}
			} else if isCommand("/help") {
				api.sendText(chatID, "¡Bienvenido a DnD!")
				api.sendText(chatID, MSG_HELP)
			} else if isCommand("/") {
				api.sendText(chatID, "¡Error, estos son los comandos disponibles!")
				api.sendText(chatID, MSG_HELP)
			} else if game != nil {
				if game.CurrentPlayer != nil && game.CurrentPlayer.ID == userID {
					// api.sendText(chatID, fmt.Sprintf("%s se ha hecho kk encima, y ha muerto...", game.CurrentPlayer.Name))
					if !game.SetNextPlayer() {
						game.Started = false
						api.sendText(chatID, MSG_GAME_ENDED)
					}
				}
			}
		}
	}
}
