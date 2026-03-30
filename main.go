package main

// TODO: Add final objective that determines the end of the campaign at the start
// TODO: Say the name of the person with every action taken as well, like buttons
// TODO: Create session alone for the creation of character, and a bot alone for it
// TODO: Permitir sugerencias de otros jugadores, diciendo que no es su turno

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
/pause - Tus mensajes no se procesan como una acción hasta usar /pause otra vez

Comandos de Creación de Jugadores:

/ready - Marca a un jugador como listo para jugar

Nota: <> se utiliza para señalar la ayuda no necesitas ponerlos literalmente
`
const MSG_PAUSE = "¡Tu turno ha sido pausado, lo que digas ya no sera una accion!"
const MSG_UNPAUSE = "¡Tu turno ha sido reanudado, ahora lo que digas sera una accion!"

const BUTTON_ROLL = "roll"
const BUTTON_PASS = "pass"
const BUTTON_PAUSE = "pause"

const BUTTON_MAGIC = "magic"
const BUTTON_NOMAGIC = "nomagic"

const BUTTON_INVENTORY = "inventory"
const BUTTON_STATS = "stats"
const BUTTON_SKILLS = "skills"

const SESSION_PROMPT = `
We will start creating the player characters before starting the game.
`

const START_GAME_PROMPT = `
All players have been created, now it's time to start the game.

Do not ask questions! Start right away with the information you have.
`
const JOIN_PROMPT = `
A new player has joined, say something short to them as a welcome, do not extend too much.
The campaign hasn't started yet, so you can't talk to them.
`
const ROLL_PROMPT = "%s has rolled, here is what he would have gotten:\n\n%s"
const TURN_PROMPT = "It's now %s's turn. Say something short to them continuing their story!\n\n%s"
const TURN_FIRST_PROMPT = "\nRight now is %s's turn."
const SAYS_PROMPT = "%s says %s"

func failIf(condition bool, msg string) {
	if condition {
		fmt.Println(msg)
		panic(msg)
	}
}

const API_BASE = "http://localhost:3000"

type ChatResponse struct {
	Response  map[string]any `json:"response"`
	SessionID string         `json:"sessionId"`
	Type      string         `json:"type"`
}

func createSession(prompt string) (string, error) {
	resp, err := http.Post(API_BASE+"/api/init", "application/json", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return "", fmt.Errorf("error creating session: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success   bool   `json:"success"`
		SessionID string `json:"sessionId"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Success {
		return result.SessionID, nil
	}
	return "", fmt.Errorf("error creating session")
}

func queryAI(session string, prompt string) (string, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"message":   prompt,
		"sessionId": session,
	})
	resp, err := http.Post(API_BASE+"/api/chat", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("error querying AI: %w", err)
	}
	defer resp.Body.Close()

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if narrative, ok := result.Response["narrative"].(string); ok {
		return narrative, nil
	}
	return "", fmt.Errorf("no narrative in response")
}

func main() {
	api := NewAPI(os.Getenv("TOKEN"))

	if len(api.token) != 46 {
		fmt.Println("Invalid token. Set TOKEN environment variable.")
		return
	}

	db := NewDatabase()

	games, err := db.LoadAllGames()
	if err != nil {
		fmt.Println("Error loading games:", err)
		games = make(map[int64]*Game)
	}
	fmt.Printf("%d games loaded from database\n", len(games))

	offset := 0

	fmt.Printf("Bot started with token ending in %s... Press Ctrl+C to stop.\n", api.token[len(api.token)-8:])

	emptyLayout := [][]InlineKeyboardButton{}
	defaultLayout := [][]InlineKeyboardButton{{
		{Text: "Inventario", CallbackData: BUTTON_INVENTORY},
		{Text: "Stats", CallbackData: BUTTON_STATS},
		{Text: "Skills", CallbackData: BUTTON_SKILLS},
	}, {
		{Text: "Roll", CallbackData: BUTTON_ROLL},
		{Text: "Pass", CallbackData: BUTTON_PASS},
		{Text: "Pause", CallbackData: BUTTON_PAUSE},
	}}

	go func() {
		for {
			time.Sleep(time.Second * 60)
			fmt.Println("Saving game states to database...")

			for chatID, game := range games {
				if err := db.SaveGame(chatID, game); err != nil {
					fmt.Printf("Error saving game %d: %v\n", chatID, err)
				}
			}
		}
	}()

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

					api.sendText(chatID, player.toString())
					api.sendText(chatID, "Ahora en base a tus decisiones, habilidades e historia de ti eres...")

					message, err := queryAI(game.SessionID, JOIN_PROMPT+fmt.Sprintf("New Player:\n%s", player.toString()))
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
					case BUTTON_MAGIC:
						api.editMessage(chatID, messageID, "¡Siempre mola verdad!", emptyLayout)
						go finalDecision()
					case BUTTON_NOMAGIC:
						api.editMessage(chatID, messageID, "¡Yo tambien opino que es para pussies!", emptyLayout)
						go finalDecision()
					}
				case StateReady:
					switch buttonKey {
					case BUTTON_ROLL:
						api.editMessage(chatID, messageID, message.Text, emptyLayout)

						if game.CurrentPlayer != nil && game.CurrentPlayer.ID == userID {
							var output bytes.Buffer

							for _, dice := range []int{4, 6, 8, 10, 12, 20} {
								output.WriteString(fmt.Sprintf("D%d: %d\n", dice, game.CurrentPlayer.Roll(dice)))
							}

							output.WriteString("\n")

							for key, value := range game.CurrentPlayer.Stats {
								output.WriteString(fmt.Sprintf("%s: %d (+%d)\n", key, value, game.CurrentPlayer.RollModifier(key)))
							}

							input := fmt.Sprintf(ROLL_PROMPT, game.CurrentPlayer.Name, output.String())
							api.sendText(chatID, input)

							message, err := queryAI(game.SessionID, input)

							if err != nil {
								api.sendText(chatID, err.Error())
								continue
							}

							api.sendButtons(chatID, message, defaultLayout)
						} else {
							api.sendText(chatID, ERR_GAME_TURN)
						}
					case BUTTON_PASS:
						api.sendText(chatID, fmt.Sprintf("%s ha pasado el turno!", game.CurrentPlayer.Name))

						if !game.SetNextPlayer() {
							game.Started = false
							api.sendText(chatID, MSG_GAME_ENDED)
						} else {
							player := game.CurrentPlayer

							api.sendText(chatID, fmt.Sprintf("Ahora es el turno de %s!", player.Name))

							message, err := queryAI(game.SessionID, fmt.Sprintf(TURN_PROMPT, player.Name, player.toString()))
							if err != nil {
								api.sendText(chatID, err.Error())
								continue
							}

							api.sendButtons(chatID, message, defaultLayout)
						}
					case BUTTON_PAUSE:
						game.CurrentPlayer.State = StatePaused
						api.sendText(chatID, MSG_PAUSE)
					case BUTTON_INVENTORY:
						api.sendText(chatID, player.Inventory())
					case BUTTON_STATS:
						api.sendText(chatID, player.toString())
					case BUTTON_SKILLS:
						api.sendText(chatID, player.Skills())
					}
				case StatePaused:
					switch buttonKey {
					case BUTTON_PAUSE:
						game.CurrentPlayer.State = StateReady
						api.sendText(chatID, MSG_UNPAUSE)
					}
				default:
					api.editMessage(chatID, messageID, buttonKey, emptyLayout)
				}

				fmt.Printf("Button from message %d\n", messageID)
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
				api.sendText(chatID, "Ahora mismo es el turno de "+game.CurrentPlayer.Name+"! La historia está cargando...")

				message, err := queryAI(game.SessionID, START_GAME_PROMPT+fmt.Sprintf(TURN_FIRST_PROMPT, game.CurrentPlayer.Name))
				if err != nil {
					api.sendText(chatID, err.Error())
					continue
				}

				api.sendButtons(chatID, message, defaultLayout)
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

					games[chatID] = &Game{PlayerIndex: -1, Players: []*Player{}, SessionID: sessionID}
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
						var out bytes.Buffer

						for key := range player.Stats {
							out.WriteString(fmt.Sprintf("En cuanto a tu %s...\n\n", key))

							rolls := make([]int, 4)
							smallest := 100
							total := 0

							for i := range 4 {
								rolls[i] = player.Roll(6)
								smallest = min(rolls[i], smallest)
								total += rolls[i]

								out.WriteString(fmt.Sprintf("%d", rolls[i]))
								out.WriteString(" ")
							}

							result := total - smallest
							player.Stats[key] = result

							out.WriteString("\n")
							if result > 16 {
								out.WriteString(fmt.Sprintf("¡Tu %s es de %d, en verdad que eres habilidoso!", key, result))
							} else if result > 14 {
								out.WriteString(fmt.Sprintf("¡Tu %s es de %d, estas sobre el promedio!", key, result))
							} else if result > 9 {
								out.WriteString(fmt.Sprintf("¡Tu %s es de %d, un valor decente!", key, result))
							} else if result > 6 {
								out.WriteString(fmt.Sprintf("Tu %s es de %d, un poco debajo del promedio...", key, result))
							} else {
								out.WriteString(fmt.Sprintf("Tu %s es de %d... oof, que mala suerte no?", key, result))
							}
							out.WriteString("\n")
						}

						out.WriteString("\nEstas satisfecho con este resultado?")

						api.sendButtons(chatID, out.String(), [][]InlineKeyboardButton{{
							{Text: "Si", CallbackData: BUTTON_MAGIC},
						}})
					}()
				} else {
					api.sendText(chatID, ERR_JOINED)
				}
			case "/pause":
				if game == nil {
					api.sendText(chatID, ERR_GAME_NIL)
					continue
				}

				if game.CurrentPlayer != nil && game.CurrentPlayer.ID == userID {
					if game.CurrentPlayer.State == StatePaused {
						game.CurrentPlayer.State = StateReady
						api.sendText(chatID, MSG_UNPAUSE)
					} else {
						game.CurrentPlayer.State = StatePaused
						api.sendText(chatID, MSG_PAUSE)
					}
				} else {
					api.sendText(chatID, ERR_GAME_TURN)
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
					if game.CurrentPlayer != nil && game.CurrentPlayer.ID == userID && game.CurrentPlayer.State == StateReady {
						prompt := fmt.Sprintf(SAYS_PROMPT, game.CurrentPlayer.Name, text)
						api.sendText(chatID, prompt)

						message, err := queryAI(game.SessionID, fmt.Sprintf("%s.\n\n%s", prompt, game.CurrentPlayer.toString()))
						if err != nil {
							api.sendText(chatID, err.Error())
							continue
						}

						api.sendButtons(chatID, message, defaultLayout)
					}
				}
			}
		}
	}
}
