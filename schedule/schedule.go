package schedule

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/eikeon/marvin/nog"
	"github.com/eikeon/scheduler"
)

type Schedule struct {
	Schedule scheduler.Schedule
	Switch   map[string]bool
}

func (s *Schedule) Run(in <-chan nog.Message, out chan<- nog.Message) {
	var scheduledEventsChannel <-chan scheduler.Event

	for {
		select {
		case e := <-scheduledEventsChannel:
			if s.Switch["Schedule"] {
				out <- nog.NewMessage("Marvin", e.What, "schedule")
			}
		case m := <-in:
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(s); err != nil {
					return
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
