package presence

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
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
	s := &Presence{}
	var presenceChannel chan presence.Presence

	go func() {
		name := "presence.html"
		if j, err := os.OpenFile(path.Join(Root, name), os.O_RDONLY, 0666); err == nil {
			if b, err := ioutil.ReadAll(j); err == nil {
				out <- nog.NewMessage("Presence", string(b), "template")
			} else {
				log.Println("ERROR reading:", err)
			}
		} else {
			log.Println("WARNING: could not open ", name, err)
		}
	}()

	for {
		select {
		case m := <-in:
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
				out <- nog.NewMessage("Marvin", p.Name+" is "+status, "Presence")

				if what, err := json.Marshal(s); err == nil {
					out <- nog.NewMessage("Marvin", string(what), "statechanged")
				} else {
					log.Println("StateChanged err:", err)
				}
			}
		}
	}
}
