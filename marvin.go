package main

import (
	"flag"
	"log"
)

var StaticRoot *string

func main() {
	log.Println("starting marvin")

	Address := flag.String("address", ":9999", "http service address")
	StaticRoot = flag.String("root", "static", "...")
	config := flag.String("config", "/etc/marvin.json", "file path to configuration file")
	flag.Parse()

	if err, s := NewSchedulerFromJSONPath(*config); err == nil {
		go ListenAndServe(*Address, s)

		if flag.NArg() == 0 {
			s.Hue.Do("chime") // visual display of scheduler starting
			s.run()
		} else {
			transition := flag.Arg(0)
			s.Hue.Do(transition)
		}
	} else {
		log.Fatal(err)
	}
}
