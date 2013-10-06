package ambientlight

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/eikeon/marvin/nog"
	"github.com/eikeon/tsl2561"
)

var Root = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	Root = path.Dir(filename)
}

type AmbientLight struct {
	nog.InOut
	Switch      map[string]bool
	DayLight    bool
	lightSensor *tsl2561.TSL2561

	lightChannel <-chan int
}

func (a *AmbientLight) Run(in <-chan nog.Message, out chan<- nog.Message) {

	name := "ambientlight.html"
	if j, err := os.OpenFile(path.Join(Root, name), os.O_RDONLY, 0666); err == nil {
		if b, err := ioutil.ReadAll(j); err == nil {
			out <- nog.NewMessage("Marvin", string(b), "template")
		} else {
			log.Println("ERROR reading:", err)
		}
	} else {
		log.Println("WARNING: could not open ", name, err)
	}

	var dayLightTime time.Time
	if t, err := tsl2561.NewTSL2561(1, tsl2561.ADDRESS_FLOAT); err == nil {
		a.lightSensor = t
		a.lightChannel = t.Broadband()
	} else {
		log.Println("Warning: Light sensor off: ", err)
		return
	}

	for {
		select {
		case m := <-in:
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(a); err != nil {
					log.Println("ambientlight decode err:", err)
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
