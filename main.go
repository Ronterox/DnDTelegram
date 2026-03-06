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

const BUTTON_ROLL_CONSTITUTION = "roll-constitution"

const BUTTON_FIRST_ITEMS = "first-items"
const BUTTON_FUCKYOU = "te jodes"

const BUTTON_HIT_CLOSE = "close"
const BUTTON_HIT_FAR = "far"

const BUTTON_ATTACK = "attack"
const BUTTON_DEFEND = "defend"

const BUTTON_MAGIC = "magic"
const BUTTON_NOMAGIC = "nomagic"

func failIf(condition bool, msg string) {
	if condition {
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

	// TODO: Make this a state enum inside of player
	settingUp := make(map[int64]bool)
	deciding := make(map[int64]bool)
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

			fmt.Printf("%v, %d\n", settingUp, userID)

			text := message.Text
			game := games[chatID]

			isCommand := func(prefix string) bool {
				return strings.HasPrefix(text, prefix)
			}

			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}

			fmt.Printf("Received: (text: %s), (button: %s)\n", text, update.CallbackQuery.Data)

			// If is a button callback
			if buttonKey := callback.Data; buttonKey != "" {
				messageID := callback.Message.ID
				chatID := callback.Message.Chat.ID
				userID := callback.User.ID
				game := games[chatID]

				fmt.Printf("Pressing button %s, %d\n", buttonKey, userID)

				finalDecision := func() {
					time.Sleep(time.Second * 2)
					api.sendText(chatID, "Ahora en base a tus decisiones, habilidades e historia de ti eres...")

					api.sendText(chatID, "<Inserta AI analizando a tu personaje>")
					api.sendText(chatID, "<Inserta mostrar toda tu informacion de personaje>")
					delete(settingUp, userID)

					player := game.FindPlayer(userID)
					player.Ready = true

					api.sendText(chatID, "Estas listo para la campaña!")
				}

				if buttonKey == BUTTON_FIRST_ITEMS && settingUp[userID] {
					api.editMessage(chatID, messageID, "Ahora investiguemos tu forma de ser...", [][]InlineKeyboardButton{})
					go func() {
						time.Sleep(time.Second * 2)
						api.sendButtons(chatID, "¿Te gusta golpear los enemigos de cerca o de lejos?", [][]InlineKeyboardButton{{
							{Text: "Cerca", CallbackData: BUTTON_HIT_CLOSE},
							{Text: "Lejos", CallbackData: BUTTON_HIT_FAR},
						}})
					}()
				} else if buttonKey == BUTTON_HIT_CLOSE && settingUp[userID] {
					api.editMessage(chatID, messageID, "Eres alguien duro no?", [][]InlineKeyboardButton{})
					go func() {
						time.Sleep(time.Second * 2)
						api.sendButtons(chatID, "¿Defender o atacar?", [][]InlineKeyboardButton{{
							{Text: "Defensa", CallbackData: BUTTON_DEFEND},
							{Text: "Ataque", CallbackData: BUTTON_ATTACK},
						}})
					}()
				} else if buttonKey == BUTTON_HIT_FAR && settingUp[userID] {
					api.editMessage(chatID, messageID, "Si se puede por que no cierto?", [][]InlineKeyboardButton{})
					go func() {
						time.Sleep(time.Second * 2)
						api.sendButtons(chatID, "¿Que opinas de la magia?", [][]InlineKeyboardButton{{
							{Text: "Me parece una buena idea", CallbackData: BUTTON_MAGIC},
							{Text: "Es una mala idea", CallbackData: BUTTON_NOMAGIC},
						}})
					}()
				} else if buttonKey == BUTTON_MAGIC && settingUp[userID] {
					api.editMessage(chatID, messageID, "¡Siempre mola verdad!", [][]InlineKeyboardButton{})
					go finalDecision()
				} else if buttonKey == BUTTON_NOMAGIC && settingUp[userID] {
					api.editMessage(chatID, messageID, "¡Yo tambien opino que es para pussies!", [][]InlineKeyboardButton{})
					go finalDecision()
				} else if buttonKey == BUTTON_DEFEND && settingUp[userID] {
					api.editMessage(chatID, messageID, "Hay que proteger lo que queremos despues de todo", [][]InlineKeyboardButton{})
					go finalDecision()
				} else if buttonKey == BUTTON_ATTACK && settingUp[userID] {
					api.editMessage(chatID, messageID, "La mejor defensa es el mejor ataque", [][]InlineKeyboardButton{})
					go finalDecision()
				} else if buttonKey == BUTTON_ROLL_CONSTITUTION {
					if player := game.FindPlayer(userID); player != nil {
						var result string

						dice := player.Roll(20)
						if dice < 13 {
							result = fmt.Sprintf("%s se ha hecho kk encima, y ha muerto...", game.CurrentPlayer.Name)
						} else {
							result = fmt.Sprintf("%s se ha podido aguantar las ganas, ha sobrevivido por ahora...", game.CurrentPlayer.Name)
						}

						api.editMessage(chatID, messageID, "Veamos que dice el destino...", [][]InlineKeyboardButton{})
						api.sendText(chatID, fmt.Sprintf("D20: %d (+%d Constitution)!\n\nEso significa %s", dice, player.RollModifier("Constitution"), result))

						if !game.SetNextPlayer() {
							game.Started = false
							api.sendText(chatID, MSG_GAME_ENDED)
						} else {
							api.sendText(chatID, fmt.Sprintf("%s es ahora tu turno!", game.CurrentPlayer.Name))
						}

						delete(deciding, userID)
					} else {
						api.editMessage(chatID, messageID, buttonKey, [][]InlineKeyboardButton{})
					}
				} else {
					api.editMessage(chatID, messageID, buttonKey, [][]InlineKeyboardButton{})
				}

				fmt.Printf("Edited message %d\n", messageID)
				continue
			}

			if isCommand("/start") {
				fmt.Println("Running start")

				if game == nil {
					api.sendText(chatID, ERR_GAME_NIL)
					continue
				}

				playerNotReady := false
				for _, player := range game.Players {
					if !player.Ready {
						api.sendText(chatID, fmt.Sprintf(ERR_READY, player.Name))
						playerNotReady = true
					}
				}

				if playerNotReady {
					continue
				}

				failIf(len(game.Players) == 0, "No players in the game")

				game.Started = true

				failIf(!game.SetNextPlayer(), "Couldn't find next player")

				api.sendText(chatID, "¡La campaña ha comenzado!")
				api.sendButtons(chatID,
					fmt.Sprintf("%s te encuentras en el baño haciendo kk, que quieres hacer?", game.CurrentPlayer.Name),
					[][]InlineKeyboardButton{{
						{Text: "Inventario", CallbackData: "inventory"},
						{Text: "Stats", CallbackData: "stats"},
						{Text: "Skills", CallbackData: "skills"},
					}})
			} else if isCommand("/join") {
				fmt.Println("Running join")

				if game == nil {
					games[chatID] = &Game{playerIndex: -1, Players: []*Player{}}
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
					newPlayer := NewPlayer(userID, message.User.FirstName, description)
					game.Players = append(game.Players, newPlayer)
					api.sendButtons(chatID, fmt.Sprintf(MSG_JOINED, newPlayer.Name), [][]InlineKeyboardButton{{
						{Text: "ready", CallbackData: "ready"},
					}})
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

				if settingUp[userID] {
					api.sendText(chatID, "Termina de configurar tu personaje")
					continue
				}

				if player := game.FindPlayer(userID); player != nil {
					if player.Ready {
						api.sendText(chatID, fmt.Sprintf("¡%s ya está listo para jugar!", player.Name))
						continue
					}

					settingUp[userID] = true
					api.sendText(chatID, "Veamos que destino depara para tu fisico y fuerza mental...")

					go func() {
						for key := range player.Character.Stats {
							api.sendText(chatID, fmt.Sprintf("En cuanto a tu %s...", key))

							rolls := make([]int, 4)
							smallest := 100
							total := 0

							for i := range 4 {
								rolls[i] = player.Roll(6)
								smallest = min(rolls[i], smallest)
								total += rolls[i]

								if i == 3 {
									time.Sleep(time.Second * 3)
								} else {
									time.Sleep(time.Second * 1)
								}

								api.sendText(chatID, fmt.Sprintf("%d", rolls[i]))
							}

							result := total - smallest
							player.Character.Stats[key] = result

							if result > 16 {
								api.sendText(chatID, fmt.Sprintf("¡Tu %s es de %d, en verdad que eres habilidoso!", key, result))
							} else if result > 14 {
								api.sendText(chatID, fmt.Sprintf("¡Tu %s es de %d, estas sobre el promedio!", key, result))
							} else if result > 9 {
								api.sendText(chatID, fmt.Sprintf("¡Tu %s es de %d, un valor decente!", key, result))
							} else if result > 6 {
								api.sendText(chatID, fmt.Sprintf("Tu %s es de %d, un poco debajo del promedio...", key, result))
							} else {
								api.sendText(chatID, fmt.Sprintf("Tu %s es de %d... oof, que mala suerte no?", key, result))
							}

							time.Sleep(time.Second * 3)
						}

						api.sendButtons(chatID, "Estas satisfecho con este resultado?", [][]InlineKeyboardButton{{
							{Text: "Si", CallbackData: BUTTON_FIRST_ITEMS},
							{Text: "No", CallbackData: BUTTON_FUCKYOU},
						}})
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
			} else if game != nil && game.Started {
				if game.CurrentPlayer != nil && game.CurrentPlayer.ID == userID && !deciding[userID] {
					deciding[userID] = true
					api.sendButtons(chatID,
						"Estas a punto de cagarte encima, rollea Constitucion para aguantar las ganas, o decide algo mas!",
						[][]InlineKeyboardButton{{{Text: "Roll", CallbackData: BUTTON_ROLL_CONSTITUTION}}})
				}
			}
		}
	}
}
