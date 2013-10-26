package nog

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"
)

type Message struct {
	Hash string `db:"HASH"`
	When string `db:"RANGE"`
	Who  string
	What string
	Why  string
}

func NewMessage(who, what, why string) Message {
	when := time.Now().Format(time.RFC3339Nano)
	return Message{Hash: when[0:10], When: when, What: what, Who: who, Why: why}
}

type Handler func(in <-chan Message, out chan<- Message)

type bit struct {
	name    string
	handler Handler
	in, out chan Message
	done    chan bool
}

func (b *bit) Run() {
	b.handler(b.in, b.out)
	b.done <- true
}

type Nog struct {
	in    chan Message
	bits  map[string]*bit
	state map[string]interface{}
	sync.Mutex
}

func NewNog() *Nog {
	n := &Nog{}
	n.in = make(chan Message)
	n.bits = make(map[string]*bit)
	n.state = make(map[string]interface{})
	n.state["Switch"] = make(map[string]interface{})
	n.state["templates"] = make(map[string]interface{})
	return n
}

func (n *Nog) Load(r io.Reader) error {
	dec := json.NewDecoder(r)
	n.Lock()
	err := dec.Decode(&n.state)
	n.Unlock()
	return err
}

func (n *Nog) Save(w io.Writer) error {
	dec := json.NewEncoder(w)
	n.Lock()
	err := dec.Encode(&n.state)
	n.Unlock()
	return err
}

func (n *Nog) updateBits() {
	n.state["Bits"] = make(map[string]interface{})
	for k, _ := range n.bits {
		n.state["Bits"].(map[string]interface{})[k] = true
	}
}

func (n *Nog) Register(name string, h Handler) {
	b := &bit{name: name, handler: h}
	n.Lock()
	n.bits[name] = b
	n.updateBits()
	n.Unlock()
}

func (n *Nog) Unregister(name string) {
	n.Lock()
	delete(n.bits, name)
	n.updateBits()
	n.Unlock()
}

func (n *Nog) Start(name string) {
	n.Lock()
	b, ok := n.bits[name]
	running := b.in != nil
	n.Unlock()
	if !ok {
		log.Println("WARNING: could not start:", name)
	} else if !running {
		b.in = make(chan Message, 50)
		b.out = make(chan Message, 50)
		b.done = make(chan bool, 1)
		n.Lock()
		switches := n.state["Switch"].(map[string]interface{})
		switches[name] = true
		n.Unlock()
		n.in <- NewMessage("Nog", "{}", "statechanged")
		go func() {
			for m := range b.out {
				m.Who = name
				m.When = time.Now().Format(time.RFC3339Nano)
				n.in <- m
			}
		}()
		go func() {
			<-b.done
			n.Lock()
			switches := n.state["Switch"].(map[string]interface{})
			switches[name] = false
			n.Unlock()
			n.in <- NewMessage("Nog", "{}", "statechanged")
		}()
		b.Run()
	}
}

func (n *Nog) Stop(name string) {
	b, ok := n.bits[name]
	if ok {
		log.Println("Stopping:", name)
		n.Lock()
		if b.in != nil {
			close(b.in)
			b.in = nil
		}
		n.Unlock()
	} else {
		log.Println("WARNING: could not start:", name)
	}
}

const TURN = "turn "

func (n *Nog) Run() {
	n.Lock()
	for _, b := range n.bits {
		if n.isOn(b.name) {
			go n.Start(b.name)
		}
	}
	n.Unlock()

	for m := range n.in {
		stateChanged := false
		if m.Why == "statechanged" {
			dec := json.NewDecoder(strings.NewReader(m.What))
			var ps map[string]interface{}
			if err := dec.Decode(&ps); err != nil {
				log.Println("statechanged err:", err)
			}
			n.Lock()
			for k, v := range ps {
				if n.state[k] == nil {
					n.state[k] = make(map[string]interface{})
				}
				n.state[k] = v
			}
			n.Unlock()
			stateChanged = true
		} else if m.Why == "template" {
			name := m.Who
			n.Lock()
			n.state["templates"].(map[string]interface{})[name] = m.What
			switches := n.state["Switch"].(map[string]interface{})
			if _, ok := switches[name].(interface{}); !ok {
				switches[name] = true
			}
			n.Unlock()
			stateChanged = true
		} else if strings.HasPrefix(m.What, TURN) {
			words := strings.SplitN(m.What[len(TURN):], " ", 2)
			if len(words) == 2 {
				if words[0] == "on" {
					go n.Start(words[1])
				} else {
					go n.Stop(words[1])
				}
			}
			stateChanged = true
		} else {
			n.Lock()
			n.broadcast(&m)
			n.Unlock()
		}
		if stateChanged {
			n.Lock()
			if what, err := json.Marshal(&n.state); err == nil {
				m := NewMessage("Nog", string(what), "statechanged")
				n.broadcast(&m)
			} else {
				panic(fmt.Sprintf("StateChanged err:%v", err))
			}
			n.Unlock()
		}
	}
}

func (n *Nog) broadcast(m *Message) {
	for _, b := range n.bits {
		if b.in != nil {
			select {
			case b.in <- *m:
			default:
				log.Println("could not broadcast message:", b)
			}
		}
	}
}

func (n *Nog) isOn(name string) bool {
	switches := n.state["Switch"].(map[string]interface{})
	val, ok := switches[name].(bool)
	return !ok || val
}
