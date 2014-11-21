package motion

import (
	"encoding/json"
	"log"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/eikeon/gpio"
	"github.com/nogiushi/marvin/nog"
)

var Root = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	Root = path.Dir(filename)
}

type Motion struct {
	Motion        bool
	motionChannel <-chan bool
}

func Handler(in <-chan nog.Message, out chan<- nog.Message) {
	out <- nog.Message{What: "started"}
	s := &Motion{}

	go func() {
		out <- nog.Template("motion")
	}()

	var motionTimer *time.Timer
	var motionTimeout <-chan time.Time

	if c, err := gpio.GPIOInterrupt(7); err == nil {
		s.motionChannel = c
	} else {
		log.Println("Warning: Motion sensor off:", err)
		out <- nog.Message{What: "no motion sensor found"}
		goto done
	}

	for {
		select {
		case m, ok := <-in:
			if !ok {
				goto done
			}
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(s); err != nil {
					log.Println("motion decode err:", err)
				}
			}
		case motion := <-s.motionChannel:
			if motion {
				out <- nog.Message{What: "motion detected"}
				const duration = 60 * time.Second
				if motionTimer == nil {
					s.Motion = true
					motionTimer = time.NewTimer(duration)
					motionTimeout = motionTimer.C // enable motionTimeout case
				} else {
					motionTimer.Reset(duration)
				}
			}
		case <-motionTimeout:
			s.Motion = false
			motionTimer = nil
			motionTimeout = nil
			out <- nog.Message{What: "motion detected timeout"}
		}
	}
done:
	out <- nog.Message{What: "stopped"}
	close(out)

}

func (m *Motion) MotionSensor() bool {
	return m.motionChannel != nil
}
