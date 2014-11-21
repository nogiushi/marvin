package daylights

import (
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

	go func() {
		out <- nog.Template("daylights")
	}()

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
