package switches

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/eikeon/marvin/nog"
)

type Switches struct {
	Switch map[string]bool
}

func (s *Switches) Run(in <-chan nog.Message, out chan<- nog.Message) {
	for {
		select {
		case m := <-in:
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(s); err != nil {
					log.Println("switches decode err:", err)
				}
			}
			const TURN = "turn "
			if strings.HasPrefix(m.What, TURN) {
				words := strings.Split(m.What[len(TURN):], " ")
				if len(words) == 2 {
					var value bool
					if words[0] == "on" {
						value = true
					} else {
						value = false
					}
					s.Switch[words[1]] = value
				}
				if what, err := json.Marshal(s); err == nil {
					out <- nog.NewMessage("Switches", string(what), "statechanged")
				} else {
					log.Println("StateChanged err:", err)
				}
			}
		}
	}
}
