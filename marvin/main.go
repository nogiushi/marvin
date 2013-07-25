package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"

	"github.com/eikeon/marvin"
	"github.com/eikeon/marvin/web"
)

var StaticRoot *string

func main() {
	config := flag.String("config", "/etc/marvin.json", "file path to configuration file")
	address := flag.String("address", ":9999", "http service address")
	cert := flag.String("cert", "", "certificate file")
	key := flag.String("key", "", "key file")
	StaticRoot = flag.String("root", "static", "...")
	flag.Parse()

	log.Println("starting marvin")

	if marvin, err := marvin.NewMarvinFromFile(*config); err == nil {
		web.AddHandlers(marvin)
		if *cert != "" || *key != "" {
			go func() {
				config := &tls.Config{ClientAuth: tls.RequestClientCert}
				server := &http.Server{Addr: *address, TLSConfig: config}
				err = server.ListenAndServeTLS(*cert, *key)
				if err != nil {
					log.Print("ListenAndServe:", err)
				}
			}()
		} else {
			go func() {
				err := http.ListenAndServe(*address, nil)
				if err != nil {
					log.Print("ListenAndServe:", err)
				}
			}()
		}
		marvin.Run()
	} else {
		log.Println("ERROR:", err)
	}

	log.Println("stopping Marvin")
}
