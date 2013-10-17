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

	"github.com/eikeon/tsl2561"
	"github.com/nogiushi/marvin/nog"
)

var Root = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	Root = path.Dir(filename)
}

type AmbientLight struct {
	Switch      map[string]bool
	DayLight    bool
	lightSensor *tsl2561.TSL2561

	lightChannel <-chan int
}

func (a *AmbientLight) Run(in <-chan nog.Message, out chan<- nog.Message) {
	options := nog.BitOptions{Name: "Ambient Light", Required: false}
	if what, err := json.Marshal(&options); err == nil {
		out <- nog.NewMessage("Ambient Light", string(what), "register")
	} else {
		log.Println("StateChanged err:", err)
	}

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
		out <- nog.NewMessage("Marvin", "no light sensor found", "Ambient Light")
		close(out)
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
					out <- nog.NewMessage("Marvin", "it is light", "Ambient Light")
					if a.Switch["Daylights"] {
						out <- nog.NewMessage("Marvin", "do daylights off", "Ambient Light")
					}
				} else if light < 4900 && (a.DayLight != false) {
					a.DayLight = false
					dayLightTime = time.Now()
					out <- nog.NewMessage("Marvin", "it is dark", "Ambient Light")
					if a.Switch["Daylights"] {
						out <- nog.NewMessage("Marvin", "do daylights on", "Ambient Light")
					}
				}
			}
		}
	}
}

func (a *AmbientLight) LightSensor() bool {
	return a.lightChannel != nil
}
