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

func Handler(in <-chan nog.Message, out chan<- nog.Message) {
	out <- nog.Message{What: "started"}
	a := &AmbientLight{}
	name := "ambientlight.html"
	if j, err := os.OpenFile(path.Join(Root, name), os.O_RDONLY, 0666); err == nil {
		if b, err := ioutil.ReadAll(j); err == nil {
			out <- nog.Message{What: string(b), Why: "template"}
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
		out <- nog.Message{What: "no light sensor found"}
		goto done
		return
	}

	for {
		select {
		case m, ok := <-in:
			if !ok {
				goto done
			}
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(a); err != nil {
					log.Println("ambientlight decode err:", err)
				}
			}
		case light, ok := <-a.lightChannel:
			if !ok {
				goto done
			}
			if time.Since(dayLightTime) > time.Duration(60*time.Second) {
				if light > 5000 && (a.DayLight != true) {
					a.DayLight = true
					dayLightTime = time.Now()
					out <- nog.Message{What: "it is light"}
				} else if light < 4900 && (a.DayLight != false) {
					a.DayLight = false
					dayLightTime = time.Now()
					out <- nog.Message{What: "it is dark"}
				}
			}
		}
	}
done:
	out <- nog.Message{What: "stopped"}
	close(out)

}

func (a *AmbientLight) LightSensor() bool {
	return a.lightChannel != nil
}
