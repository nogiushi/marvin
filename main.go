package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"
)

var StaticRoot *string
var marvin Marvin
var config = flag.String("config", "/etc/marvin.json", "file path to configuration file")

func main() {
	log.Println("starting marvin")

	Address := flag.String("address", ":9999", "http service address")
	StaticRoot = flag.String("root", "static", "...")
	flag.Parse()

	if j, err := os.OpenFile(*config, os.O_RDONLY, 0666); err == nil {
		dec := json.NewDecoder(j)
		if err = dec.Decode(&marvin); err != nil {
			log.Fatal("err:", err)
		}
		j.Close()
	} else {
		log.Println("could not read config file:", *config)
		marvin = Marvin{}
	}

	go marvin.loop()

	go ListenAndServe(*Address, &marvin)

	notifyChannel := make(chan os.Signal, 1)
	signal.Notify(notifyChannel, os.Interrupt)
	for {
		select {
		case <-time.NewTicker(60 * time.Second).C:
			if j, err := os.Create(*config); err == nil {
				dec := json.NewEncoder(j)
				var c Marvin = marvin
				if err = dec.Encode(&c); err != nil {
					log.Fatal("err:", err)
				}
				j.Close()
			} else {
				log.Fatal(err)
			}
		case sig := <-notifyChannel:
			switch sig {
			case os.Interrupt:
				log.Println("handling:", sig)
				goto Done
			default:
				log.Fatal("Unexpected Signal:", sig)
			}
		}
	}
Done:
	log.Println("stopping Marvin")

}
