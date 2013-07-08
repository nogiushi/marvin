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
	Lights map[string]light
	Groups map[string]group
}

type light struct {
	Name  string
	State lightStateGet
}

type group struct {
	Name   string
	Lights []string
	Action lightStateGet
}

type lightStateCommon struct {
	On     bool
	Bri    uint8
	Alert  string
	Effect string
}

type lightStateGet struct {
	lightStateCommon

	ColorMode string
	Hue       uint16 //uint16w
	Sat       uint8
	Xy        [2]float32
	Ct        uint16
	Reachable bool
}

func (ls lightStateGet) HTMLColor() string {
	return "blue"
}

type lightStatePut struct {
	lightStateCommon

	TransitionTime uint16
}

type lightStatePutHS struct {
	lightStatePut

	Hue uint16 //uint16w
	Sat uint8
}

type lightStatePutXY struct {
	lightStatePut

	Xy [2]float32
}

type lightStatePutCT struct {
	lightStatePut

	Ct uint16
}

type errorResponse struct {
	Error struct {
		Type        int
		Address     string
		Description string
	}
}

type bridge struct {
	Id                string
	Internalipaddress string
	Macaddress        string
}
type bridgeResponse []bridge

func (h *hue) getURL() string {
	if h.Host == "" {
		if response, err := http.Get("http://www.meethue.com/api/nupnp"); err == nil {
			dec := json.NewDecoder(response.Body)
			var br bridgeResponse
			if err = dec.Decode(&br); err == nil {
				if len(br) == 1 {
					b := br[0]
					h.Host = b.Internalipaddress
				} else {
					log.Fatal("bridgeResponse != 1 not yet implemented")
				}
			} else {
				log.Fatal("could not decode bridgeResponse:", err)
			}
			response.Body.Close()
		} else {
			log.Fatal("could not get:", err)
		}

	}
	if h.Key == "" {
		// TODO: fix me
		h.Key = "334b473e8c2555d5eb722e0c932df793"
	}
	return "http://" + h.Host + "/api/" + h.Key
}

func (h *hue) GetLights() {
	if response, err := http.Get(h.getURL() + "/lights"); err == nil {
		dec := json.NewDecoder(response.Body)
		if err = dec.Decode(&(h.Lights)); err != nil {
			log.Fatal("could not decode lightsResponse:", err)
		}
		response.Body.Close()
	} else {
		log.Fatal("could not get lights:", err)
	}
}

func (h *hue) GetGroups() {
	if response, err := http.Get(h.getURL() + "/groups"); err == nil {
		dec := json.NewDecoder(response.Body)
		if err = dec.Decode(&(h.Groups)); err != nil {
			log.Fatal("could not decode groupsResponse:", err)
		}
		response.Body.Close()
	} else {
		log.Fatal("could not get groups:", err)
	}
}

func (h *hue) GetState() {
	if h.Lights == nil {
		h.GetLights()
	}
	for name, _ := range h.Lights {
		if response, err := http.Get(h.getURL() + "/lights/" + name); err == nil {
			dec := json.NewDecoder(response.Body)
			var l light
			if err = dec.Decode(&l); err == nil {
				h.Lights[name] = l
			} else {
				log.Fatal("could not decode light:", err)
			}
			response.Body.Close()
		} else {
			log.Fatal("could not get light:", err)
		}
	}

	if h.Groups == nil {
		h.GetGroups()
		h.Groups["0"] = group{}
	}
	for name, _ := range h.Groups {
		if response, err := http.Get(h.getURL() + "/groups/" + name); err == nil {
			dec := json.NewDecoder(response.Body)
			var l group
			if err = dec.Decode(&l); err == nil {
				h.Groups[name] = l
			} else {
				log.Fatal("could not decode group:", err)
			}
			response.Body.Close()
		} else {
			log.Fatal("could not get group:", err)
		}
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
		h.Set(address, h.States[name])
	}
}

func (h *hue) Set(address string, value interface{}) {
	url := "http://" + h.Host + "/api/" + h.Key + address
	b, err := json.Marshal(value)
	if err != nil {
		log.Println("ERROR: json.Marshal: " + err.Error())
		return
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
