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
	Host   string
	Key    string
	States map[string]interface{}
}

func (h *hue) state(name string) string {
	b, err := json.Marshal(h.States[name])
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

func (h *hue) run(command command) {
	state := h.state(command.State)
	client := &http.Client{}
	url := "http://" + h.Host + "/api/" + h.Key + command.Address
	log.Println("put: " + url + " body:" + state)
	body := strings.NewReader(state)
	r, err := http.NewRequest("PUT", url, body)
	if err != nil {
		log.Fatal(err)
	}
	response, err := client.Do(r)
	if err != nil {
		log.Println(err)
	}
	defer response.Body.Close()
}
