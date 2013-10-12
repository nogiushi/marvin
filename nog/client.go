package nog

import (
	"log"

	"code.google.com/p/go.net/websocket"
)

func RemoteAdd(url_, protocol, origin string, r Bit) {
	ws, err := websocket.Dial(url_, protocol, origin)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		var m Message
		if err := websocket.JSON.Receive(ws, &m); err == nil {
			log.Println("out:", m)
			r.SendIn() <- m
		} else {
			log.Println("Message Websocket receive err:", err)
			return
		}
	}()
	go func() {
		for m := range r.ReceiveOut() {
			if err := websocket.JSON.Send(ws, &m); err != nil {
				log.Println("Message Websocket send err:", err)

			} else {
				log.Println("in:", m)
			}
		}
	}()
	r.Run(r.ReceiveIn(), r.SendOut())
}
