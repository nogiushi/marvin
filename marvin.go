package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/eikeon/marvin/actions"
	"github.com/eikeon/marvin/activity"
	"github.com/eikeon/marvin/ambientlight"
	"github.com/eikeon/marvin/hue"
	"github.com/eikeon/marvin/lightstates"
	"github.com/eikeon/marvin/motion"
	"github.com/eikeon/marvin/nog"
	"github.com/eikeon/marvin/nouns"
	"github.com/eikeon/marvin/presence"
	"github.com/eikeon/marvin/schedule"
	"github.com/eikeon/marvin/web"
)

func main() {
	config := flag.String("config", "/etc/marvin.json", "file path to configuration file")
	address := flag.String("address", ":9999", "http service address")
	cert := flag.String("cert", "", "certificate file")
	key := flag.String("key", "", "key file")
	flag.Parse()

	log.Println("starting marvin")

	if n, err := nog.NewNogFromFile(*config); err == nil {
		go n.Add(&actions.Actions{}, &nog.BitOptions{Name: "Actions"})
		go n.Add(&activity.Activity{}, &nog.BitOptions{Name: "Activity"})
		go n.Add(&schedule.Schedule{}, &nog.BitOptions{Name: "Schedule"})
		go n.Add(&hue.Hue{}, &nog.BitOptions{Name: "Lights"})
		go n.Add(&lightstates.Lightstates{}, &nog.BitOptions{Name: "LightStates"})
		go n.Add(&presence.Presence{}, &nog.BitOptions{Name: "Presence"})
		go n.Add(&motion.Motion{}, &nog.BitOptions{Name: "Motion"})
		go n.Add(&nouns.Nouns{}, &nog.BitOptions{Name: "Nouns"})
		go n.Add(&ambientlight.AmbientLight{}, &nog.BitOptions{Name: "Ambient Light"})

		web.AddHandlers(n)
		addresses := strings.Split(*address, ",")
		for i, addr := range addresses {
			if i == 1 || (len(addresses) == 1 && (*cert != "" || *key != "")) {
				log.Println("starting secure:", addr, i)
				go func(a string) {
					config := &tls.Config{ClientAuth: tls.RequestClientCert}
					server := &http.Server{Addr: a, TLSConfig: config}
					err = server.ListenAndServeTLS(*cert, *key)
					if err != nil {
						log.Print("ListenAndServe:", err)
					}
				}(addr)
			} else {
				log.Println("starting:", addr, i)
				go func(a string) {
					server := &http.Server{Addr: a}
					err := server.ListenAndServe()
					if err != nil {
						log.Print("ListenAndServe:", err)
					}
				}(addr)
			}
		}
		n.Run()
	} else {
		log.Println("ERROR:", err)
	}

	log.Println("stopping Marvin")
}
