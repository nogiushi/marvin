package ambientlight

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/eikeon/tsl2561"
	"github.com/nogiushi/marvin/nog"
)

var Root = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	Root = path.Dir(filename)
}

func Handler(in <-chan nog.Message, out chan<- nog.Message) {
	out <- nog.Message{What: "started"}

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

	var description string
	var lightChannel <-chan int
	var dayLightTime time.Time
	if t, err := tsl2561.NewTSL2561(1, tsl2561.ADDRESS_FLOAT); err == nil {
		lightChannel = t.Broadband()
	} else {
		log.Println("Warning: Light sensor off: ", err)
		out <- nog.Message{What: "no light sensor found"}
		goto done
		return
	}

	for {
		select {
		case _, ok := <-in:
			if !ok {
				goto done
			}
		case light, ok := <-lightChannel:
			if !ok {
				goto done
			}
			if time.Since(dayLightTime) > time.Duration(60*time.Second) {
				if light > 5000 && (description != "light") {
					description = "light"
					dayLightTime = time.Now()
					out <- nog.Message{What: "it is " + description}
				} else if light < 4900 && (description != "dark") {
					description = "dark"
					dayLightTime = time.Now()
					out <- nog.Message{What: "it is " + description}
				}
			}
		}
	}
done:
	out <- nog.Message{What: "stopped"}
	close(out)

}
