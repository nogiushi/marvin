package daylights

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/nogiushi/marvin/nog"
)

var Root = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	Root = path.Dir(filename)
}

func Handler(in <-chan nog.Message, out chan<- nog.Message) {
	out <- nog.Message{What: "started"}

	name := "daylights.html"
	if j, err := os.OpenFile(path.Join(Root, name), os.O_RDONLY, 0666); err == nil {
		if b, err := ioutil.ReadAll(j); err == nil {
			out <- nog.Message{What: string(b), Why: "template"}
		} else {
			log.Println("ERROR reading:", err)
		}
	} else {
		log.Println("WARNING: could not open ", name, err)
	}

	for m := range in {
		if m.What == "it is light" {
			out <- nog.Message{What: "do daylights off"}
		}
		if m.What == "it is dark" {
			out <- nog.Message{What: "do daylights on"}
		}
	}

	out <- nog.Message{What: "stopped"}
	close(out)
}
