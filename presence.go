package main

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var hosts map[string]*host = map[string]*host{}

type host struct {
	name      string
	addressIn chan net.IP
	address   net.IP
	present   bool
}

func (h *host) ping() {
	c := make(chan bool)

	ports := []string{"22", "62078"}

	for _, port := range ports {
		go func(port string) {
			conn, err := net.DialTimeout("tcp", h.address.String()+":"+port, 10*time.Second)
			if err == nil {
				c <- true
				conn.Close()
			} else {
				c <- false
			}
		}(port)
	}

	timeout := time.After(10 * time.Second)

	for i := 0; i < len(ports); i++ {
		select {
		case r := <-c:
			if r == true {
				if h.present == false {
					h.present = true
					marvin.Do <- h.name + " on"
				}
				return
			}
		case <-timeout:
			break
		}
	}
	if h.present == true {
		h.present = false
		marvin.Do <- h.name + " off"
	}
}

func (h *host) watch() {
	c := time.Tick(60 * time.Second)
	for {
		select {
		case a := <-h.addressIn:
			h.address = a
			h.ping()
			time.Sleep(10)
		case <-c:
			h.ping()
		}
	}
}

func listen() {
	mcaddr, err := net.ResolveUDPAddr("udp", "224.0.0.251:5353") // mdns/bonjour
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenMulticastUDP("udp", nil, mcaddr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		buf := make([]byte, dns.MaxMsgSize)
		read, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("err:", err)
		}
		var msg dns.Msg
		if err := msg.Unpack(buf[:read]); err == nil {
			if msg.MsgHdr.Response {
				for i := 0; i < len(msg.Answer); i++ {
					rr := msg.Answer[i]
					rh := rr.Header()
					if rh.Rrtype == dns.TypeA {
						name := strings.NewReplacer(".local.", "").Replace(rh.Name)
						h, ok := hosts[name]
						if !ok {
							ch := make(chan net.IP)
							h = &host{name: name, address: rr.(*dns.A).A, addressIn: ch}
							hosts[name] = h
							go h.watch()
						}
						h.addressIn <- rr.(*dns.A).A
					}
				}
			}
		} else {
			log.Println("err:", err)
		}
	}
}
