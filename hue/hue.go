package hue

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/eikeon/hue"
	"github.com/eikeon/marvin/nog"
)

type Hue struct {
	Hue         hue.Hue
	States      map[string]interface{}
	Transitions map[string]struct {
		Switch   map[string]bool
		Commands []struct {
			Address string
			State   string
		}
	}
}

func (h *Hue) Run(in <-chan nog.Message, out chan<- nog.Message) {
	var createUserChan <-chan time.Time
	for {
		select {
		case <-createUserChan:
			if err := h.Hue.CreateUser(h.Hue.Username, "Marvin"); err == nil {
				createUserChan = nil
			} else {
				out <- nog.NewMessage("Marvin", "press hue link button to authenticate", "setup")
			}
		case m := <-in:
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(h); err != nil {
					return
				}
				if createUserChan == nil {
					if err := h.Hue.GetState(); err != nil {
						createUserChan = time.NewTicker(1 * time.Second).C
					} else {
						// TODO:
					}
				}
			}
			const SETHUE = "set hue address "
			if strings.HasPrefix(m.What, SETHUE) {
				words := strings.Split(m.What[len(SETHUE):], " ")
				if len(words) == 3 {
					address := words[0]
					state := words[2]
					var s interface{}
					dec := json.NewDecoder(strings.NewReader(state))
					if err := dec.Decode(&s); err != nil {
						log.Println("json decode err:", err)
					} else {
						h.Hue.Set(address, s)
						err := h.Hue.GetState()
						if err != nil {
							log.Println("ERROR:", err)
						}
						if what, err := json.Marshal(h); err == nil {
							out <- nog.NewMessage("Marvin", string(what), "statechanged")
						} else {
							log.Println("StateChanged err:", err)
						}
					}
					/* TODO: move to activity?
					} else if {

					t, ok := m.Transitions[what]
					if ok {
						for k, v := range t.Switch {
							m.Switch[k] = v
						}
					}
					for _, command := range t.Commands {
						address := command.Address
						if strings.Contains(command.Address, "/light") {
							address += "/state"
						} else {
							address += "/action"
						}
						b, err := json.Marshal(m.States[command.State])
						if err != nil {
							log.Println("ERROR: json.Marshal: " + err.Error())
						} else {
							m.Do("Marvin", "set hue address "+address+" to "+string(b), what)
						}
					}
					m.StateChanged()
					*/
				} else {
					log.Println("unexpected number of words in:", m)
				}
			}
		}
	}
}
