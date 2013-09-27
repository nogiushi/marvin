package marvin

import (
	"log"

	"github.com/eikeon/scheduler"
)

func schedule(messages chan Message, states chan State) {
	on := true
	var scheduledEventsChannel <-chan scheduler.Event

	for {
		select {
		case e := <-scheduledEventsChannel:
			if on {
				messages <- NewMessage("Marvin", e.What, "schedule")
			}
		case sc := <-states:
			on = sc.Switch["Schedule"]
			if scheduledEventsChannel == nil {
				sr := sc.Schedule
				if c, err := sr.Run(); err == nil {
					scheduledEventsChannel = c
				} else {
					log.Println("Warning: Scheduled events off:", err)
					break
				}
			}
		}
	}
}
