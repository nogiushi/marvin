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
	port := "62078"
	if h.name == "chatte" || h.name == "gato" {
		port = "22"
	}
	conn, err := net.DialTimeout("tcp", h.address.String()+":"+port, 10*time.Second)
	if err == nil {
		conn.Close()
		if h.present == false {
			h.present = true
			log.Println(h.name, "ON")
			//scheduler.Hue.Do(h.name)
		}
	} else {
		log.Println("err:", err)
		if h.present == true {
			h.present = false
			log.Println(h.name, "OFF")
		}
	}
}

func (h *host) watch(scheduler *scheduler) {
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

func listen(scheduler *scheduler) {
	mcaddr, err := net.ResolveUDPAddr("udp", "224.0.0.251:5353") // mdns/bonjour
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenMulticastUDP("udp", nil, mcaddr)
	if err != nil {
		log.Fatal(err)
	}
	ping := func(name string) {
		m := new(dns.Msg)
		m.SetQuestion(name, dns.TypeA)
		out, err := m.Pack()
		if err != nil {
			log.Println("ERR?:", err)
		}
		if err != nil {
			n, err := conn.WriteToUDP(out, mcaddr)
			log.Println("n", n, "ERR?:", err)
		}
	}
	go ping("Lynx.local.")
	go ping("gato.local.")
	go ping("chatte.local.")
	go ping("Siri-Dells-iPhone.local.")

	for {
		buf := make([]byte, 1500)
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
							go h.watch(scheduler)
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
