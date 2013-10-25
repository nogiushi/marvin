package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nogiushi/marvin/actions"
	"github.com/nogiushi/marvin/activity"
	"github.com/nogiushi/marvin/ambientlight"
	"github.com/nogiushi/marvin/hue"
	"github.com/nogiushi/marvin/lightstates"
	"github.com/nogiushi/marvin/motion"
	"github.com/nogiushi/marvin/nog"
	"github.com/nogiushi/marvin/nouns"
	"github.com/nogiushi/marvin/persist"
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

	n := nog.NewNog()
	if j, err := os.OpenFile(*config, os.O_RDONLY, 0666); err == nil {
		if err = n.Load(j); err != nil {
			panic(err)
		}
		j.Close()
	} else {
		log.Println("ERROR: could not open config:", err)
	}

	go n.Run()
	for _, handler := range []nog.Handler{actions.Handler, activity.Handler, schedule.Handler, hue.Handler, lightstates.Handler, presence.Handler, motion.Handler, nouns.Handler, ambientlight.Handler} {
		b := n.Register(handler)
		go b.Run()

	}

	p := &persist.Persist{}
	b := n.Register(p.Handler)
	b.Run()
	web.AddPersistenceHandlers(p)

	web.AddHandlers(n)
	addresses := strings.Split(*address, ",")
	for i, addr := range addresses {
		if i == 1 || (len(addresses) == 1 && (*cert != "" || *key != "")) {
			log.Println("starting secure:", addr, i)
			go func(a string) {
				config := &tls.Config{ClientAuth: tls.RequestClientCert}
				server := &http.Server{Addr: a, TLSConfig: config}
				if err := server.ListenAndServeTLS(*cert, *key); err != nil {
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

	notifyChannel := make(chan os.Signal, 1)
	signal.Notify(notifyChannel, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
	sig := <-notifyChannel
	log.Println("handling:", sig)

	if j, err := os.Create(*config); err == nil {
		if err = n.Save(j); err == nil {
			log.Println("saved:", *config)
		} else {
			log.Println("ERROR: saving config", err)
		}
		j.Close()
	} else {
		log.Println("ERROR: could not create config file:", err)
	}

	log.Println("stopping Marvin")
}
