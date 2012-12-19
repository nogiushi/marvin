package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type command struct {
	Light string
	State   string
	Action   string
}

type hue struct {
	Host        string
	Key         string
	Addresses   map[string]string
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
	address := h.Addresses[command.Light]
	var body string
	if command.State != ""{
		body = h.state(command.State)
		address += "/state"
	}
	if command.Action != "" {
		body = h.state(command.Action)
		address += "/action"
	}
	client := &http.Client{}
	url := "http://" + h.Host + "/api/" + h.Key + address
	if r, err := http.NewRequest("PUT", url, strings.NewReader(body)); err == nil {
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
