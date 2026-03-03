package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

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

type API struct {
	token string
	base  string
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
	url := a.base + "/sendMessage"
	fmt.Printf("Sending text to chat %d: %s\n", chatID, text)
	jsonData, _ := json.Marshal(map[string]any{"chat_id": chatID, "text": text})
	http.Post(url, "application/json", bytes.NewBuffer(jsonData))
}
