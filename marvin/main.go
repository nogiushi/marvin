package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"strings"

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
		marvin.Run()
	} else {
		log.Println("ERROR:", err)
	}

	log.Println("stopping Marvin")
}
