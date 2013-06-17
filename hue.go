package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type hue struct {
	Host        string
	Key         string
	Addresses   map[string]string
	States      map[string]interface{}
	Transitions map[string][]struct {
		Light  string
		State  string
		Action string
	}
}

func (h *hue) Do(transition string) {
	for _, command := range h.Transitions[transition] {
		address := h.Addresses[command.Light]
		var name string
		if command.State != "" {
			name = command.State
			address += "/state"
		} else if command.Action != "" {
			name = command.Action
			address += "/action"
		}
		url := "http://" + h.Host + "/api/" + h.Key + address
		b, err := json.Marshal(h.States[name])
		if err != nil {
			log.Println("ERROR: json.Marshal: " + err.Error())
			continue
		}
		if r, err := http.NewRequest("PUT", url, bytes.NewReader(b)); err == nil {
			if response, err := http.DefaultClient.Do(r); err == nil {
				response.Body.Close()
				time.Sleep(100 * time.Millisecond)
			} else {
				log.Println("ERROR: client.Do: " + err.Error())
			}
		} else {
			log.Println("ERROR: NewRequest: " + err.Error())
		}
	}
}
