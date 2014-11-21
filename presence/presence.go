package presence

import (
	"encoding/json"
	"log"
	"path"
	"runtime"
	"strings"

	"github.com/eikeon/presence"
	"github.com/nogiushi/marvin/nog"
)

var Root = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	Root = path.Dir(filename)
}

type Presence struct {
	Present map[string]bool
}

func Handler(in <-chan nog.Message, out chan<- nog.Message) {
	out <- nog.Message{What: "started"}
	s := &Presence{}
	var presenceChannel chan presence.Presence

	go func() {
		out <- nog.Template("presence")
	}()

	for {
		select {
		case m, ok := <-in:
			if !ok {
				goto done
			}
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(s); err != nil {
					log.Println("presence decode err:", err)
				}
				if s.Present == nil {
					s.Present = make(map[string]bool)
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
				out <- nog.Message{What: p.Name + " is " + status}

				if what, err := json.Marshal(s); err == nil {
					out <- nog.Message{What: string(what), Why: "statechanged"}
				} else {
					log.Println("StateChanged err:", err)
				}
			}
		}
	}
done:
	out <- nog.Message{What: "stopped"}
	close(out)

}
