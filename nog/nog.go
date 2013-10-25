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
	in, out chan Message
	handler Handler
	done    chan bool
}

func (b *bit) Run() {
	b.handler(b.in, b.out)
	b.done <- true
}

type Nog struct {
	in    chan Message
	bits  []*bit
	bitsL sync.Mutex
	state map[string]interface{}
}

func NewNog() *Nog {
	n := &Nog{}
	n.in = make(chan Message)
	n.state = make(map[string]interface{})
	return n
}

func (n *Nog) Load(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(&n.state)
}

func (n *Nog) Save(w io.Writer) error {
	dec := json.NewEncoder(w)
	return dec.Encode(&n.state)
}

func (n *Nog) Register(h Handler) *bit {
	b := &bit{in: make(chan Message, 50), out: make(chan Message, 50), handler: h, done: make(chan bool, 1)}
	n.bitsL.Lock()
	n.bits = append(n.bits, b)
	n.bitsL.Unlock()
	n.in <- NewMessage("Nog", "{}", "statechanged")
	go func() {
		m := <-b.out
		n.bitsL.Lock()
		b.name = m.Who
		n.bitsL.Unlock()
		n.in <- m
		for m := range b.out {
			n.in <- m
		}
	}()
	go func() {
		<-b.done
		n.bitsL.Lock()
		var nb []*bit
		for _, bb := range n.bits {
			if bb != b {
				nb = append(nb, bb)
			}
		}
		n.bits = nb
		n.bitsL.Unlock()
		n.in <- NewMessage("Nog", "{}", "statechanged")
	}()
	return b
}

func (n *Nog) Run() {
	for m := range n.in {
		stateChanged := false
		if m.Why == "statechanged" {
			dec := json.NewDecoder(strings.NewReader(m.What))
			var ps map[string]interface{}
			if err := dec.Decode(&ps); err != nil {
				log.Println("statechanged err:", err)
			}
			for k, v := range ps {
				if n.state[k] == nil {
					n.state[k] = make(map[string]interface{})
				}
				n.state[k] = v
			}
			stateChanged = true
		} else if m.Why == "template" {
			name := m.Who
			n.state["templates"].(map[string]interface{})[name] = m.What
			switches := n.state["Switch"].(map[string]interface{})
			if _, ok := switches[name].(interface{}); !ok {
				switches[name] = true
			}
			stateChanged = true
		}
		const TURN = "turn "
		if strings.HasPrefix(m.What, TURN) {
			words := strings.SplitN(m.What[len(TURN):], " ", 2)
			if len(words) == 2 {
				var value bool
				if words[0] == "on" {
					value = true
				} else {
					value = false
				}
				switches := n.state["Switch"].(map[string]interface{})
				switches[words[1]] = value
			}
			stateChanged = true
		}
		if n.isOn(m.Why) {
			n.broadcast(&m)
		}
		if stateChanged {
			bits := make(map[string]interface{})
			n.bitsL.Lock()
			for _, b := range n.bits {
				if b.name != "" {
					bits[b.name] = true
				}
			}
			n.bitsL.Unlock()

			n.state["Bits"] = bits

			if what, err := json.Marshal(&n.state); err == nil {
				m := NewMessage("Nog", string(what), "statechanged")
				n.broadcast(&m)
			} else {
				panic(fmt.Sprintf("StateChanged err:%v", err))
			}
		}
	}
}

func (n *Nog) broadcast(m *Message) {
	n.bitsL.Lock()
	for _, b := range n.bits {
		select {
		case b.in <- *m:
		default:
			log.Println("could not broadcast message")
		}
	}
	n.bitsL.Unlock()
}

func (n *Nog) isOn(name string) bool {
	switches := n.state["Switch"].(map[string]interface{})
	val, ok := switches[name].(bool)
	return !ok || val
}
