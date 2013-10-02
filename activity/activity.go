package activity

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/eikeon/marvin/nog"
)

type activity struct {
	Name string
	Next map[string]bool
}

type Activity struct {
	Activities map[string]*activity
	Activity   string
}

func (m *Activity) GetActivity(name string) *activity {
	if name != "" {
		a, ok := m.Activities[name]
		if !ok {
			a = &activity{name, map[string]bool{}}
			m.Activities[name] = a
		}
		return a
	} else {
		return nil
	}
}

func (m *Activity) UpdateActivity(name string) {
	s := m.GetActivity(m.Activity)
	if s != nil {
		s.Next[name] = true
	}
	m.GetActivity(name)
	m.Activity = name
}

func (a *Activity) Run(in <-chan nog.Message, out chan<- nog.Message) {
	for {
		select {
		case m := <-in:
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(a); err != nil {
					log.Println("switches decode err:", err)
				}
			}
			const IAM = "I am "
			if strings.HasPrefix(m.What, IAM) {
				what := m.What[len(IAM):]
				a.UpdateActivity(what)
			}
		}
	}
}
