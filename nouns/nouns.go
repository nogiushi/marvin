package nouns

import (
	"encoding/json"
	"log"
	"path"
	"runtime"
	"strings"

	"github.com/nogiushi/marvin/nog"
)

var Root = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	Root = path.Dir(filename)
}

type Nouns struct {
	Nouns map[string]string
}

func Handler(in <-chan nog.Message, out chan<- nog.Message) {
	out <- nog.Message{What: "started"}
	a := &Nouns{}

	go func() {
		out <- nog.Template("nouns")
	}()
	
	for m := range in {
		if m.Why == "statechanged" {
			dec := json.NewDecoder(strings.NewReader(m.What))
			if err := dec.Decode(a); err != nil {
				log.Println("nouns decode error:", err)
			}
		}
	}
	out <- nog.Message{What: "stopped"}
	close(out)

}
