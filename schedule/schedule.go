package schedule

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/eikeon/marvin/nog"
	"github.com/eikeon/scheduler"
)

var Root = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	Root = path.Dir(filename)
}

type Schedule struct {
	nog.InOut
	Schedule scheduler.Schedule
	Switch   map[string]bool
}

func (s *Schedule) Run(in <-chan nog.Message, out chan<- nog.Message) {
	var scheduledEventsChannel <-chan scheduler.Event

	name := "schedule.html"
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
		case e := <-scheduledEventsChannel:
			if s.Switch["Schedule"] {
				out <- nog.NewMessage("Marvin", e.What, "Schedule")
			}
		case m := <-in:
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(s); err != nil {
					log.Println("schedule decode err:", err)
				}
				if scheduledEventsChannel == nil {
					if c, err := s.Schedule.Run(); err == nil {
						scheduledEventsChannel = c
					} else {
						log.Println("Warning: Scheduled events off:", err)
					}
				}
			}
		}
	}
}