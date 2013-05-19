package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/stathat/go"
)

var StaticRoot *string

func postStat(name string, value float64) {
	if err := stathat.PostEZValue(name, "eikeon@eikeon.com", value); err != nil {
		log.Printf("error posting value %v: %d", err, value)
	}
}
func main() {
	log.Println("starting marvin")

	Address := flag.String("address", ":9999", "http service address")
	StaticRoot = flag.String("root", "static", "...")
	config := flag.String("config", "/etc/marvin.json", "file path to configuration file")
	verbose := flag.Bool("verbose", false, "verbose output")
	flag.Parse()

	err, s := NewSchedulerFromJSONPath(*config)
	if err == nil {
		go ListenAndServe(*Address, s)
		go listen(s)

		if flag.NArg() == 0 {
			s.Hue.Do("chime") // visual display of scheduler starting
			go s.run()
		} else {
			transition := flag.Arg(0)
			go s.Hue.Do(transition)
		}
	} else {
		log.Fatal(err)
	}

	t, err := NewTSL2561(1, ADDRESS_FLOAT)
	if err != nil {
		log.Println("WARNING: could not create TSL2561:", err)
	}

	notifyChannel := make(chan os.Signal, 1)
	signal.Notify(notifyChannel, os.Interrupt)
	ticker := time.NewTicker(1 * time.Second)

	dayLight := make(chan bool, 10)
	var lastDayLight interface{}
	lastDayLightTime := time.Now()
	for {
		select {
		case value := <-dayLight:
			if value {
				s.Hue.Do("daylight")
				log.Println("daylight")
			} else {
				s.Hue.Do("daylight off")
				log.Println("daylight off")
			}
		case <-ticker.C:
			if t != nil {
				if err = t.On(); err != nil {
					log.Fatal("could not turn on:", err)
				}
				time.Sleep(t.IntegrationDuration())

				if value, err := t.GetBroadband(); err == nil {
					dl := value > 5000
					if lastDayLight == nil {
						lastDayLight = dl
						lastDayLightTime = time.Now()
					} else if time.Since(lastDayLightTime) > time.Duration(60*time.Second) {
						if value > 5000 && lastDayLight == false {
							lastDayLight = true
							lastDayLightTime = time.Now()
							dayLight <- true
						} else if value < 4900 && lastDayLight == true {
							lastDayLight = false
							lastDayLightTime = time.Now()
							dayLight <- false
						}
					}
					if *verbose {
						log.Println("broadband:", value)
					}
					go postStat("light broadband", float64(value))
				} else {
					log.Println("error getting broadband value:", err)
				}
				if value, err := t.GetInfrared(); err == nil {
					if *verbose {
						log.Println("infrared:", value)
					}
					go postStat("light infrared", float64(value))
				} else {
					log.Println("error getting infrared value:", err)
				}

				if err := t.Off(); err != nil {
					log.Fatal("Could not turn off:", err)
				}
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
