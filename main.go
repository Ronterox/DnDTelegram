package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const ALBERTO_ID = 1426815752

const ERR_GAME_NIL = "No hay ninguna campaña en curso, empieza una uniendote con /join"
const ERR_GAME_STARTED = "La campaña ya ha comenzado, ¡No puedes unirte!"
const ERR_GAME_TURN = "¡No es tu turno aun!"
const ERR_READY = "¡%s no está listo para jugar utiliza /ready para marcarlo!"
const ERR_JOINED = "No te has unido a la campaña, unete con /join"

const MSG_GAME_STARTED = "¡Una nueva campaña ha sido iniciada!"
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

const SESSION_PROMPT = `
We will start creating the player characters before starting the game.
`
const START_GAME_PROMPT = `
All players have been created, now it's time to start the game.
`
const JOIN_PROMPT = `
A new player has joined, say something short to them as a welcome, do not extend too much.
The campaign hasn't started yet, so you can't talk to them.
`

func failIf(condition bool, msg string) {
	if condition {
		fmt.Println(msg)
		panic(msg)
	}
}

// Creates a session and returns the ID of the last session created
func createSession(prompt string) (string, error) {
	sessionCommand := exec.Command("opencode", "run", "--agent", "dnd", prompt)

	// opencode session list | grep ses | cut -d' ' -f1 | head -1
	sessionList := exec.Command("opencode", "session", "list")
	grepSessions := exec.Command("grep", "ses")
	cutRest := exec.Command("cut", "-d' '", "-f1")
	getHead := exec.Command("head", "-1")

	if _, err := sessionCommand.Output(); err != nil {
		return "", fmt.Errorf("error creating session: %w", err)
	}

	grepSessions.Stdin, _ = sessionList.StdoutPipe()
	cutRest.Stdin, _ = grepSessions.StdoutPipe()
	getHead.Stdin, _ = cutRest.StdoutPipe()

	sessionList.Start()
	grepSessions.Start()
	cutRest.Start()

	sessionID, err := getHead.CombinedOutput()

	sessionList.Wait()
	grepSessions.Wait()
	cutRest.Wait()

	if err != nil {
		return "", fmt.Errorf("error getting session ID: %w", err)
	}

	return string(sessionID), nil
}

// Queries the AI for a response to the given prompt, returns the response
func queryAI(session string, prompt string) (string, error) {
	cmd := exec.Command("opencode", "run", "-s", session, "--agent", "dnd", prompt)

	stdout, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error querying AI: %w", err)
	}

	return string(stdout), nil
}

func main() {
	api := NewAPI(os.Getenv("TOKEN"))

	if len(api.token) != 46 {
		fmt.Println("Invalid token. Set TOKEN environment variable.")
		return
	}

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
			var state PlayerState
			var player *Player
			var game *Game

			fmt.Printf("Update: %+v\n", update)

			message := update.Message
			callback := update.CallbackQuery

			chatID := message.Chat.ID
			userID := message.User.ID

			text := message.Text
			if game = games[chatID]; game != nil {
				fmt.Printf("Game found: %+v\n", game)
				if player = game.FindPlayer(userID); player != nil {
					state = player.State
					fmt.Printf("Player found: %+v\n", player)
				}
			}

			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}

			// If is a button callback
			if buttonKey := callback.Data; buttonKey != "" {
				messageID := callback.Message.ID
				chatID := callback.Message.Chat.ID
				userID := callback.User.ID

				if game = games[chatID]; game != nil {
					fmt.Printf("Game found: %+v\n", game)
					if player = game.FindPlayer(userID); player != nil {
						state = player.State
						fmt.Printf("Player found: %+v\n", player)
					}
				}

				fmt.Printf("Pressing button %s, %d\n", buttonKey, userID)

				finalDecision := func() {
					failIf(player == nil, "Player not found")

					api.sendText(chatID, player.Character.String())
					api.sendText(chatID, "Ahora en base a tus decisiones, habilidades e historia de ti eres...")

					message, err := queryAI(game.SessionID, JOIN_PROMPT+fmt.Sprintf("New Player:\n%s", &player.Character))
					if err != nil {
						api.sendText(chatID, err.Error())
						return
					}
					api.sendText(chatID, message)

					player.State = StateReady

					api.sendText(chatID, "Estas listo para la campaña!")
				}

				switch state {
				case StateSettingUp:
					switch buttonKey {
					case BUTTON_FIRST_ITEMS:
						api.editMessage(chatID, messageID, "Ahora investiguemos tu forma de ser...", [][]InlineKeyboardButton{})
						go func() {
							time.Sleep(time.Second * 2)
							api.sendButtons(chatID, "¿Te gusta golpear los enemigos de cerca o de lejos?", [][]InlineKeyboardButton{{
								{Text: "Cerca", CallbackData: BUTTON_HIT_CLOSE},
								{Text: "Lejos", CallbackData: BUTTON_HIT_FAR},
							}})
						}()
					case BUTTON_HIT_CLOSE:
						api.editMessage(chatID, messageID, "Eres alguien duro no?", [][]InlineKeyboardButton{})
						go func() {
							time.Sleep(time.Second * 2)
							api.sendButtons(chatID, "¿Defender o atacar?", [][]InlineKeyboardButton{{
								{Text: "Defensa", CallbackData: BUTTON_DEFEND},
								{Text: "Ataque", CallbackData: BUTTON_ATTACK},
							}})
						}()
					case BUTTON_HIT_FAR:
						api.editMessage(chatID, messageID, "Si se puede por que no cierto?", [][]InlineKeyboardButton{})
						go func() {
							time.Sleep(time.Second * 2)
							api.sendButtons(chatID, "¿Que opinas de la magia?", [][]InlineKeyboardButton{{
								{Text: "Me parece una buena idea", CallbackData: BUTTON_MAGIC},
								{Text: "Es una mala idea", CallbackData: BUTTON_NOMAGIC},
							}})
						}()
					case BUTTON_MAGIC:
						api.editMessage(chatID, messageID, "¡Siempre mola verdad!", [][]InlineKeyboardButton{})
						go finalDecision()
					case BUTTON_NOMAGIC:
						api.editMessage(chatID, messageID, "¡Yo tambien opino que es para pussies!", [][]InlineKeyboardButton{})
						go finalDecision()
					case BUTTON_DEFEND:
						api.editMessage(chatID, messageID, "Hay que proteger lo que queremos despues de todo", [][]InlineKeyboardButton{})
						go finalDecision()
					case BUTTON_ATTACK:
						api.editMessage(chatID, messageID, "La mejor defensa es el mejor ataque", [][]InlineKeyboardButton{})
						go finalDecision()
					}
				case StateDeciding:
					if buttonKey != BUTTON_ROLL_CONSTITUTION {
						continue
					}

					var result string

					dice := game.CurrentPlayer.Roll(20)
					if dice < 13 {
						result = fmt.Sprintf("%s se ha hecho kk encima, y ha muerto...", game.CurrentPlayer.Name)
					} else {
						result = fmt.Sprintf("%s se ha podido aguantar las ganas, ha sobrevivido por ahora...", game.CurrentPlayer.Name)
					}

					api.editMessage(chatID, messageID, "Veamos que dice el destino...", [][]InlineKeyboardButton{})
					api.sendText(chatID, fmt.Sprintf("D20: %d (+%d Constitution)!\n\nEso significa %s", dice, game.CurrentPlayer.RollModifier("Constitution"), result))

					if !game.SetNextPlayer() {
						game.Started = false
						api.sendText(chatID, MSG_GAME_ENDED)
					} else {
						api.sendText(chatID, fmt.Sprintf("%s es ahora tu turno!", game.CurrentPlayer.Name))
					}

					game.CurrentPlayer.State = StateReady
				default:
					api.editMessage(chatID, messageID, buttonKey, [][]InlineKeyboardButton{})
				}

				fmt.Printf("Edited message %d\n", messageID)
				continue
			}

			command, rest, hasArgs := strings.Cut(text, " ")

			switch command {
			case "/start":
				if game == nil {
					api.sendText(chatID, ERR_GAME_NIL)
					continue
				}

				playerNotReady := false
				for _, player := range game.Players {
					if player.State != StateReady {
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
				api.sendText(chatID, "Right now is "+game.CurrentPlayer.Name+"'s turn... The story is loading...")

				message, err := queryAI(game.SessionID, START_GAME_PROMPT+fmt.Sprintf("\nRight now is %s's turn.", game.CurrentPlayer.Name))
				if err != nil {
					api.sendText(chatID, err.Error())
					continue
				}

				api.sendButtons(chatID, message,
					[][]InlineKeyboardButton{{
						{Text: "Inventario", CallbackData: "inventory"},
						{Text: "Stats", CallbackData: "stats"},
						{Text: "Skills", CallbackData: "skills"},
					}})
			case "/join":
				if player != nil {
					api.sendText(chatID, fmt.Sprintf("¡Ya eres un jugador! %+v", player))
					continue
				}

				if game == nil {
					api.sendText(chatID, "Creando nueva sesión...")

					sessionID, err := createSession(SESSION_PROMPT)
					if err != nil {
						api.sendText(chatID, err.Error())
						continue
					}

					games[chatID] = &Game{playerIndex: -1, Players: []*Player{}, SessionID: sessionID}
					game = games[chatID]

					api.sendText(chatID, MSG_GAME_STARTED)
				}

				if game.Started {
					api.sendText(chatID, ERR_GAME_STARTED)
					continue
				}

				if hasArgs {
					newPlayer := NewPlayer(userID, message.User.FirstName, rest)
					game.Players = append(game.Players, newPlayer)
					api.sendText(chatID, fmt.Sprintf(MSG_JOINED, newPlayer.Name))
				} else {
					api.sendText(chatID, "Escribe una descripción para tu personaje /join <descripcion>")
				}
			case "/whoami":
				fmt.Println("Running whoami")

				if game == nil {
					api.sendText(chatID, ERR_GAME_NIL)
					continue
				}

				if player != nil {
					api.sendText(chatID, fmt.Sprintf("Eres %+v", player))
				} else {
					api.sendText(chatID, ERR_JOINED)
				}
			case "/roll":
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
			case "/ready":
				fmt.Println("Running ready")

				if game == nil {
					api.sendText(chatID, ERR_GAME_NIL)
					continue
				}

				if state == StateSettingUp {
					api.sendText(chatID, "Termina de configurar tu personaje")
					continue
				}

				if player != nil {
					if state == StateReady {
						api.sendText(chatID, fmt.Sprintf("¡%s ya está listo para jugar!", player.Name))
						continue
					}

					player.State = StateSettingUp
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

								// if i == 3 {
								// 	time.Sleep(time.Second * 3)
								// } else {
								// 	time.Sleep(time.Second * 1)
								// }
								//
								// api.sendText(chatID, fmt.Sprintf("%d", rolls[i]))
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

							// time.Sleep(time.Second * 3)
						}

						api.sendButtons(chatID, "Estas satisfecho con este resultado?", [][]InlineKeyboardButton{{
							{Text: "Si", CallbackData: BUTTON_FIRST_ITEMS},
							{Text: "No", CallbackData: BUTTON_FUCKYOU},
						}})
					}()
				} else {
					api.sendText(chatID, ERR_JOINED)
				}
			case "/help":
				api.sendText(chatID, "¡Bienvenido a DnD!")
				api.sendText(chatID, MSG_HELP)
			case "/":
				api.sendText(chatID, "¡Error, estos son los comandos disponibles!")
				api.sendText(chatID, MSG_HELP)
			default:
				fmt.Println("Running default")
				if game != nil && game.Started {
					if game.CurrentPlayer != nil && game.CurrentPlayer.ID == userID && state != StateDeciding {
						prompt := fmt.Sprintf("%s says %s", game.CurrentPlayer.Name, text)
						api.sendText(chatID, prompt)

						// game.CurrentPlayer.State = StateDeciding

						message, err := queryAI(game.SessionID, fmt.Sprintf("%s.\n\n%s", prompt, &game.CurrentPlayer.Character))
						if err != nil {
							api.sendText(chatID, err.Error())
							continue
						}

						api.sendButtons(chatID, message, [][]InlineKeyboardButton{{{Text: "Roll", CallbackData: BUTTON_ROLL_CONSTITUTION}}})
					}
				}
			}
		}
	}
}
