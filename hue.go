package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type command struct {
	Address string
	State   string
}

type hue struct {
	Host        string
	Key         string
	States      map[string]interface{}
	Transitions map[string][]command
}

func (h *hue) Do(transition string) {
	for _, command := range h.Transitions[transition] {
		h.run(command)
	}
}

func (h *hue) state(name string) string {
	b, err := json.Marshal(h.States[name])
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

func (h *hue) run(command command) (err error) {
	state := h.state(command.State)
	client := &http.Client{}
	url := "http://" + h.Host + "/api/" + h.Key + command.Address
	body := strings.NewReader(state)
	if r, err := http.NewRequest("PUT", url, body); err == nil {
		if response, err := client.Do(r); err == nil {
			response.Body.Close()
		} else {
			log.Println("ERROR: client.Do: " + err.Error())
		}
	} else {
		log.Println("ERROR: NewRequest: " + err.Error())
	}
	return err
}
