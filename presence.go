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
	name    string
	address net.IP
}

func (h *host) ping(ch chan pair) {
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
				ch <- pair{h.name, true}
				return
			}
		case <-timeout:
			break
		}
	}
	ch <- pair{h.name, false}
}

type pair struct {
	name   string
	status bool
}

func Listen(current map[string]bool) chan pair {
	for name, _ := range current {
		hosts[name] = &host{name: name}
		if a, err := net.ResolveTCPAddr("tcp4", name+".local:80"); err == nil {
			hosts[name].address = a.IP
		}
	}

	ch := make(chan pair)
	mcaddr, err := net.ResolveUDPAddr("udp", "224.0.0.251:5353") // mdns/bonjour
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenMulticastUDP("udp", nil, mcaddr)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
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
								h = &host{name: name, address: rr.(*dns.A).A}
								hosts[name] = h
							}
							h.address = rr.(*dns.A).A
							go h.ping(ch)
						}
					}
				}
			} else {
				log.Println("err:", err)
			}
		}
	}()

	go func() {
		c := time.Tick(60 * time.Second)
		for {
			select {
			case <-c:
				for _, h := range hosts {
					go h.ping(ch)
				}
			}
		}
	}()

	return ch
}
