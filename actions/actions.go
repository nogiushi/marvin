package actions

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/eikeon/hu"
	"github.com/nogiushi/marvin/nog"
)

var Root = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	Root = path.Dir(filename)
}

type Sentence []hu.Term

func (sentence Sentence) String() string {
	var terms []string
	for _, term := range sentence {
		terms = append(terms, term.String())
	}
	return strings.Join(terms, " ")
}

type Actions struct {
	Actions map[string]string
}

func Handler(in <-chan nog.Message, out chan<- nog.Message) {
	out <- nog.Message{What: "started"}
	a := &Actions{}

	go func() {
		name := "actions.html"
		if j, err := os.OpenFile(path.Join(Root, name), os.O_RDONLY, 0666); err == nil {
			if b, err := ioutil.ReadAll(j); err == nil {
				out <- nog.Message{What: "started"}
				out <- nog.Message{What: string(b), Why: "template"}
			} else {
				log.Println("ERROR reading:", err)
			}
		} else {
			log.Println("WARNING: could not open ", name, err)
		}
	}()

	for m := range in {
		if m.Why == "statechanged" {
			dec := json.NewDecoder(strings.NewReader(m.What))
			if err := dec.Decode(a); err != nil {
				log.Println("actions decode err:", err)
			} else {
				if a.Actions == nil {
					a.Actions = make(map[string]string)
				}
			}
		}

		const DOACTION = "do "

		what := ""
		if strings.HasPrefix(m.What, DOACTION) {
			what = m.What[len(DOACTION):]
		}

		t, ok := a.Actions[what]
		if ok {
			reader := strings.NewReader(t)
			for {
				expression := hu.ReadSentence(reader)
				if expression == nil {
					break
				}
				m := Sentence(expression).String()
				out <- nog.Message{What: m}
			}
		}

		const SET = "set action "
		if strings.HasPrefix(m.What, SET) {
			e := strings.SplitN(m.What[len(SET):], " to ", 2)
			if len(e) == 2 {
				a.Actions[e[0]] = e[1]
			}
			if what, err := json.Marshal(a); err == nil {
				out <- nog.Message{What: string(what), Why: "statechanged"}
			} else {
				log.Println("StateChanged err:", err)
			}
		}
	}
	out <- nog.Message{What: "stopped"}
	close(out)
}
