package actions

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/eikeon/marvin/nog"
)

var Root = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	Root = path.Dir(filename)
}

type Actions struct {
	nog.InOut
	Actions map[string][]string
}

func (a *Actions) Run(in <-chan nog.Message, out chan<- nog.Message) {
	name := "actions.html"
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
		case m := <-in:
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(a); err != nil {
					log.Println("actions decode err:", err)
				}
			}

			const DOACTION = "do "

			what := ""
			if strings.HasPrefix(m.What, DOACTION) {
				what = m.What[len(DOACTION):]
			}

			t, ok := a.Actions[what]
			if ok {
				for _, m := range t {
					out <- nog.NewMessage("Marvin", m, "Actions")
				}
			}
		}
	}
}
