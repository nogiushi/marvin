package ambientlight

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/eikeon/marvin/nog"
	"github.com/eikeon/tsl2561"
)

type AmbientLight struct {
	Switch      map[string]bool
	DayLight    bool
	lightSensor *tsl2561.TSL2561

	lightChannel <-chan int
}

func (a *AmbientLight) Run(in <-chan nog.Message, out chan<- nog.Message) {
	var dayLightTime time.Time
	if t, err := tsl2561.NewTSL2561(1, tsl2561.ADDRESS_FLOAT); err == nil {
		a.lightSensor = t
		a.lightChannel = t.Broadband()
	} else {
		log.Println("Warning: Light sensor off: ", err)
	}

	for {
		select {
		case m := <-in:
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(a); err != nil {
					return
				}
			}
		case light := <-a.lightChannel:
			if time.Since(dayLightTime) > time.Duration(60*time.Second) {
				if light > 5000 && (a.DayLight != true) {
					a.DayLight = true
					dayLightTime = time.Now()
					out <- nog.NewMessage("Marvin", "it is light", "sensors")
					if a.Switch["Daylights"] {
						out <- nog.NewMessage("Marvin", "daylight", "it is light")
					}
				} else if light < 4900 && (a.DayLight != false) {
					a.DayLight = false
					dayLightTime = time.Now()
					out <- nog.NewMessage("Marvin", "it is dark", "sensors")
					if a.Switch["Daylights"] {
						out <- nog.NewMessage("Marvin", "daylight off", "it is dark")
					}
				}
			}
		}
	}
}

func (a *AmbientLight) LightSensor() bool {
	return a.lightChannel != nil
}
