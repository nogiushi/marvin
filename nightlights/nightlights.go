package nightlights

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
		out <- nog.Template("nightlights")
	}()

	for m := range in {
		if m.What == "motion detected" {
			out <- nog.Message{What: "do nightlights on", Why: "nightlights detected"}
		}
		if m.What == "motion detected timeout" {
			out <- nog.Message{What: "set light All to off", Why: "motion detected timeout"}
		}
	}

	out <- nog.Message{What: "stopped"}
	close(out)
}
