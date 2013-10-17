package hue

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/eikeon/hue"
	"github.com/nogiushi/marvin/nog"
)

var Root = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	Root = path.Dir(filename)
}

type Hue struct {
	Hue         hue.Hue
	Nouns       map[string]string
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
	options := nog.BitOptions{Name: "Lights", Required: false}
	if what, err := json.Marshal(&options); err == nil {
		out <- nog.NewMessage("Lights", string(what), "register")
	} else {
		log.Println("StateChanged err:", err)
	}

	var createUserChan <-chan time.Time

	name := "hue.html"
	if j, err := os.OpenFile(path.Join(Root, name), os.O_RDONLY, 0666); err == nil {
		if b, err := ioutil.ReadAll(j); err == nil {
			out <- nog.NewMessage("Marvin", string(b), "template")
		} else {
			log.Println("ERROR reading:", err)
		}
	} else {
		log.Println("WARNING: could not open ", name, err)
	}

	for {
		select {
		case <-createUserChan:
			if err := h.Hue.CreateUser(h.Hue.Username, "Marvin"); err == nil {
				createUserChan.Stop()
			} else {
				out <- nog.NewMessage("Marvin", fmt.Sprintf("%s: press hue link button to authenticate", err), "Lights")
			}
		case m := <-in:
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(h); err != nil {
					log.Println("hue decode err:", err)
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
				} else {
					log.Println("unexpected number of words in:", m)
				}
			}
			const SET = "set light "
			if strings.HasPrefix(m.What, SET) {
				e := strings.SplitN(m.What[len(SET):], " to ", 2)
				if len(e) == 2 {
					address := h.Nouns[e[0]]
					state := h.States[e[1]]
					if strings.Contains(address, "/light") {
						address += "/state"
					} else {
						address += "/action"
					}
					h.Hue.Set(address, state)
					err := h.Hue.GetState()
					if err != nil {
						log.Println("ERROR:", err)
					}
					if what, err := json.Marshal(h); err == nil {
						out <- nog.NewMessage("Marvin", string(what), "statechanged")
					} else {
						log.Println("StateChanged err:", err)
					}
				} else {
					log.Println("unexpected number of words in:", m)
				}
			}
		}
	}
}
