package lightstates

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
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

type Lightstates struct {
	Lightstates map[string][]string
}

func Handler(in <-chan nog.Message, out chan<- nog.Message) {
	out <- nog.Message{What: "started"}
	a := &Lightstates{}
	go func() {
		name := "lightstates.html"
		if j, err := os.OpenFile(path.Join(Root, name), os.O_RDONLY, 0666); err == nil {
			if b, err := ioutil.ReadAll(j); err == nil {
				out <- nog.Message{What: string(b), Why: "template"}
			} else {
				log.Println("ERROR reading:", err)
			}
		} else {
			log.Println("WARNING: could not open ", name, err)
		}
	}()

	for m := range in {
		if m.Why == "statechanged" {
			dec := json.NewDecoder(strings.NewReader(m.What))
			if err := dec.Decode(a); err != nil {
				log.Println("lightstates decode err:", err)
			}
		}
	}
	out <- nog.Message{What: "stopped"}
	close(out)

}
