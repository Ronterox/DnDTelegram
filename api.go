package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

type Message struct {
	ID   int64  `json:"message_id"`
	Text string `json:"text"`
	Chat struct {
		ID   int64  `json:"id"`
		Type string `json:"type"` // private, group, supergroup, channel
	} `json:"chat"`
	User User `json:"from"`
}

type Update struct {
	UpdateID      int     `json:"update_id"`
	Message       Message `json:"message"`
	CallbackQuery struct {
		ID      string   `json:"id"`
		Data    string   `json:"data"`
		User    User     `json:"from"`
		Message *Message `json:"message"`
	} `json:"callback_query"`
}

type UpdateResult struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type API struct {
	token string
	base  string
}

type TelegramResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
	Parameters  struct {
		RetryAfter int `json:"retry_after"`
	} `json:"parameters"`
}

func NewAPI(token string) *API {
	return &API{token: token, base: "https://api.telegram.org/bot" + token}
}

func (a *API) getUpdates(offset int) (UpdateResult, error) {
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=30", a.base, offset)

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

func (a *API) sendText(chatID int64, text string) {
	fmt.Printf("Sending text to chat %x: %s\n", chatID, text)
	jsonData, _ := json.Marshal(map[string]any{"chat_id": chatID, "text": text})
	http.Post(a.base+"/sendMessage", "application/json", bytes.NewBuffer(jsonData))
}

const MAX_TRIES = 3

func (a *API) sendButtons(chatID int64, text string, buttons [][]InlineKeyboardButton) {
	jsonData, _ := json.Marshal(map[string]any{
		"chat_id":      chatID,
		"text":         text,
		"reply_markup": InlineKeyboardMarkup{InlineKeyboard: buttons},
	})

	for range MAX_TRIES {
		resp, err := http.Post(a.base+"/sendMessage", "application/json", bytes.NewBuffer(jsonData))

		if err != nil {
			fmt.Printf("Network Error: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var tgResp TelegramResponse
			json.NewDecoder(resp.Body).Decode(&tgResp)

			if resp.StatusCode == 429 {
				seconds := tgResp.Parameters.RetryAfter
				if seconds == 0 {
					seconds = 1
				}

				fmt.Printf("Rate limited! Sleeping for %d seconds...\n", seconds)
				time.Sleep(time.Duration(seconds) * time.Second)
				continue
			}

			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Telegram Rejected Request: %s\n", string(body))
			continue
		}

		fmt.Println("Message sent successfully")
		break
	}
}

func (a *API) editMessage(chatID int64, messageID int64, text string, buttons [][]InlineKeyboardButton) {
	jsonData, _ := json.Marshal(map[string]any{
		"chat_id":      chatID,
		"message_id":   messageID,
		"text":         text,
		"reply_markup": InlineKeyboardMarkup{InlineKeyboard: buttons},
	})

	for i := range MAX_TRIES {
		resp, err := http.Post(a.base+"/editMessageText", "application/json", bytes.NewBuffer(jsonData))

		if err != nil {
			fmt.Printf("Attempt %d: Network error: %v\n", i+1, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("Message edited successfully!")
			break
		}

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Edit failed for message %d: %s\n", messageID, string(body))

		if strings.Contains(string(body), "message is not modified") {
			break
		}

		var tgResp TelegramResponse
		json.NewDecoder(resp.Body).Decode(&tgResp)

		if resp.StatusCode == 429 {
			seconds := tgResp.Parameters.RetryAfter
			if seconds == 0 {
				seconds = 1
			}

			fmt.Printf("Rate limited! Sleeping for %d seconds...\n", seconds)
			time.Sleep(time.Duration(seconds) * time.Second)
			continue
		}
	}
}
