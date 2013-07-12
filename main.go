package main

import (
	"flag"
	"log"
	"net/http"
)

var StaticRoot *string

func main() {
	config := flag.String("config", "/etc/marvin.json", "file path to configuration file")
	Address := flag.String("address", ":9999", "http service address")
	StaticRoot = flag.String("root", "static", "...")
	flag.Parse()

	log.Println("starting marvin")

	if marvin, err := NewMarvinFromFile(*config); err == nil {
		marvin.AddHandlers()
		go func() {
			err := http.ListenAndServe(*Address, nil)
			if err != nil {
				log.Print("ListenAndServe:", err)
			}
		}()
		marvin.loop()
	} else {
		log.Println("ERROR:", err)
	}

	log.Println("stopping Marvin")
}
