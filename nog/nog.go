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

type Rudiment interface {
	Run(in <-chan Message, out chan<- Message)
}

type listeners struct {
	m map[*chan Message]bool
	sync.Mutex
}

func (l *listeners) notify(m *Message) {
	l.Lock()
	for o := range l.m {
		select {
		case *o <- *m:
		default:
			log.Println("unable to send to channel:", *o)
		}
	}
	l.Unlock()
}

func (l *listeners) Register(c *chan Message) {
	l.Lock()
	defer l.Unlock()
	l.m[c] = true
}

func (l *listeners) Unregister(c *chan Message) {
	l.Lock()
	defer l.Unlock()
	delete(l.m, c)
	close(*c)
}

type Nog struct {
	In  chan<- Message
	out <-chan Message
	*listeners
	path  string
	state map[string]interface{}
}

func NewNogFromFile(path string) (*Nog, error) {
	n := Nog{}
	ch := make(chan Message, 10)
	n.In = ch
	n.out = ch
	n.listeners = &listeners{m: make(map[*chan Message]bool)}

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

	return &n, nil
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

func (n *Nog) Add(r Rudiment) {
	c := make(chan Message, 10)
	n.Register(&c)
	r.Run(c, n.In)
	n.listeners.Unregister(&c)
}

func (n *Nog) statechanged() *Message {
	if what, err := json.Marshal(&n.state); err == nil {
		m := NewMessage("Nog", string(what), "statechanged")
		return &m
	} else {
		panic(fmt.Sprintf("StateChanged err:%v", err))
	}
}

func (n *Nog) Register(c *chan Message) {
	n.listeners.Register(c)
	*c <- *n.statechanged()
}

func (n *Nog) StateChanged() {
	n.notify(n.statechanged())
}
