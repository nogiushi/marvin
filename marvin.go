package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/nogiushi/marvin/actions"
	"github.com/nogiushi/marvin/activity"
	"github.com/nogiushi/marvin/ambientlight"
	"github.com/nogiushi/marvin/hue"
	"github.com/nogiushi/marvin/lightstates"
	"github.com/nogiushi/marvin/motion"
	"github.com/nogiushi/marvin/nog"
	"github.com/nogiushi/marvin/nouns"
	"github.com/nogiushi/marvin/presence"
	"github.com/nogiushi/marvin/schedule"
	"github.com/nogiushi/marvin/web"
)

func main() {
	config := flag.String("config", "/etc/marvin.json", "file path to configuration file")
	address := flag.String("address", ":9999", "http service address")
	cert := flag.String("cert", "", "certificate file")
	key := flag.String("key", "", "key file")
	flag.Parse()

	log.Println("starting marvin")

	if n, err := nog.NewNogFromFile(*config); err == nil {
		for _, b := range []nog.Bit{&actions.Actions{}, &activity.Activity{}, &schedule.Schedule{}, &hue.Hue{}, &lightstates.Lightstates{}, &presence.Presence{}, &motion.Motion{}, &nouns.Nouns{}, &ambientlight.AmbientLight{}} {
			c := nog.InOut{}
			go n.Add(c.ReceiveOut(), c.SendIn())
			go b.Run(c.ReceiveIn(), c.SendOut())
		}

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
