package presence

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/eikeon/marvin/nog"
	"github.com/eikeon/presence"
)

type Presence struct {
	Present map[string]bool
}

func (s *Presence) Run(in <-chan nog.Message, out chan<- nog.Message) {
	var presenceChannel chan presence.Presence

	for {
		select {
		case m := <-in:
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(s); err != nil {
					return
				}
				if presenceChannel == nil {
					presenceChannel = presence.Listen(s.Present)
				}
			}
		case p := <-presenceChannel:
			if s.Present[p.Name] != p.Status {
				s.Present[p.Name] = p.Status
				var status string
				if p.Status {
					status = "home"
				} else {
					status = "away"
				}
				out <- nog.NewMessage("Marvin", p.Name+" is "+status, "presence")

				if what, err := json.Marshal(s); err == nil {
					out <- nog.NewMessage("Marvin", string(what), "statechanged")
				} else {
					log.Println("StateChanged err:", err)
				}
			}
		}
	}
}
