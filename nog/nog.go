package nog

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
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

type InOut struct {
	in, out chan Message
}

func (b *InOut) ReceiveOut() <-chan Message {
	if b.out == nil {
		b.out = make(chan Message, 10)
	}
	return b.out
}
func (b *InOut) SendOut() chan<- Message {
	if b.out == nil {
		b.out = make(chan Message, 10)
	}
	return b.out
}
func (b *InOut) ReceiveIn() <-chan Message {
	if b.in == nil {
		b.in = make(chan Message, 10)
	}
	return b.in
}
func (b *InOut) SendIn() chan<- Message {
	if b.in == nil {
		b.in = make(chan Message, 20)
	}
	return b.in
}

type Rudiment interface {
	Run(in <-chan Message, out chan<- Message)
	ReceiveOut() <-chan Message
	SendIn() chan<- Message
	ReceiveIn() <-chan Message
	SendOut() chan<- Message
}

type BitOptions struct {
	Name     string
	Required bool
}

type listeners struct {
	m map[chan<- Message]*BitOptions
	sync.Mutex
}

func (l *listeners) Register(c chan<- Message, options *BitOptions) {
	l.Lock()
	defer l.Unlock()
	l.m[c] = options
}

func (l *listeners) Unregister(c chan<- Message) {
	l.Lock()
	defer l.Unlock()
	close(c)
	delete(l.m, c)

}

type Nog struct {
	In  chan<- Message
	out <-chan Message
	*listeners
	*persist
	path  string
	state map[string]interface{}
}

func NewNogFromFile(path string) (*Nog, error) {
	n := Nog{}
	ch := make(chan Message, 10)
	n.In = ch
	n.out = ch
	n.listeners = &listeners{m: make(map[chan<- Message]*BitOptions)}
	n.persist = &persist{}

	go func() {
		persistMessages := make(chan Message, 50)
		n.Register(persistMessages, &BitOptions{Name: "Persist", Required: true})
		n.persist.Run(persistMessages, nil)
		n.Unregister(persistMessages)
	}()

	n.path = path
	if j, err := os.OpenFile(n.path, os.O_RDONLY, 0666); err == nil {
		dec := json.NewDecoder(j)
		if err = dec.Decode(&n.state); err != nil {
			return nil, err
		}
		j.Close()
	} else {
		return nil, err
	}
	if n.state["Switch"] == nil {
		n.state["Switch"] = make(map[string]bool)
	}
	n.state["Bits"] = make(map[string]bool)
	return &n, nil
}

func (n *Nog) isOn(name string) bool {
	switches := n.state["Switch"].(map[string]interface{})
	if val, _ := switches[name].(bool); val {
		return true
	} else {
		return false
	}
}

func (n *Nog) notify(m *Message) {
	n.Lock()
	for o, info := range n.m {
		if info.Required || n.isOn(info.Name) {
			select {
			case o <- *m:
			default:
				log.Println("unable to send to channel for:", info.Name)
			}
		}
	}
	n.Unlock()
}

func (n *Nog) Run() {
	notifyChannel := make(chan os.Signal, 1)
	signal.Notify(notifyChannel, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)

	saveChannel := time.NewTicker(3600 * time.Second).C

	for {
		select {
		case m := <-n.out:
			log.Println("Message:", m)

			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				var ps map[string]interface{}
				if err := dec.Decode(&ps); err != nil {
					log.Println("statechanged err:", err)
				}
				for k, v := range ps {
					n.state[k] = v
				}
				n.StateChanged()
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
				n.StateChanged()
			}

			n.notify(&m)

		case <-saveChannel:
			if err := n.Save(n.path); err == nil {
				log.Println("saved:", n.path)
			} else {
				log.Println("ERROR: saving", err)
			}
		case sig := <-notifyChannel:
			log.Println("handling:", sig)
			goto Done
		}
	}
Done:
	if err := n.Save(n.path); err == nil {
		log.Println("saved:", n.path)
	} else {
		log.Println("ERROR: saving config", err)
	}
}

func (n *Nog) Save(path string) error {
	if j, err := os.Create(path); err == nil {
		dec := json.NewEncoder(j)
		if err = dec.Encode(&n.state); err != nil {
			return err
		}
		j.Close()
	} else {
		return err
	}
	return nil
}

func (n *Nog) Add(r Rudiment, options *BitOptions) {
	if options != nil && options.Name != "" && options.Required == false {
		name := options.Name
		switches := n.state["Switch"].(map[string]interface{})
		if _, ok := switches[name].(bool); !ok {
			switches[name] = true
		}
	}
	n.Register(r.SendIn(), options)
	n.state["Bits"].(map[string]bool)[options.Name] = true
	go func() {
		for m := range r.ReceiveOut() {

			if m.Why == "template" {
				if t, ok := n.state["templates"].(map[string]interface{}); ok {
					t[options.Name] = m.What
				} else {
					n.state["templates"] = make(map[string]string)
				}
				n.StateChanged()
				continue
			}

			switches := n.state["Switch"].(map[string]interface{})
			if options != nil && switches[options.Name].(bool) {
				n.In <- m
			}
		}
	}()
	r.Run(r.ReceiveIn(), r.SendOut())
	n.listeners.Unregister(r.SendIn())
	delete(n.state["Bits"].(map[string]bool), options.Name)
	n.StateChanged()
}

func (n *Nog) statechanged() *Message {
	if what, err := json.Marshal(&n.state); err == nil {
		m := NewMessage("Nog", string(what), "statechanged")
		return &m
	} else {
		panic(fmt.Sprintf("StateChanged err:%v", err))
	}
}

func (n *Nog) Register(c chan<- Message, options *BitOptions) {
	n.listeners.Register(c, options)
	c <- *n.statechanged()
}

func (n *Nog) StateChanged() {
	n.notify(n.statechanged())
}
