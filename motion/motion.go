package motion

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/eikeon/gpio"
	"github.com/eikeon/marvin/nog"
)

type Motion struct {
	Motion        bool
	Switch        map[string]bool
	motionChannel <-chan bool
}

func (s *Motion) Run(in <-chan nog.Message, out chan<- nog.Message) {
	if c, err := gpio.GPIOInterrupt(7); err == nil {
		s.motionChannel = c
	} else {
		log.Println("Warning: Motion sensor off:", err)
	}
	var motionTimer *time.Timer
	var motionTimeout <-chan time.Time

	for {
		select {
		case m := <-in:
			if m.Why == "statechanged" {
				dec := json.NewDecoder(strings.NewReader(m.What))
				if err := dec.Decode(s); err != nil {
					return
				}
			}
		case motion := <-s.motionChannel:
			if motion {
				out <- nog.NewMessage("Marvin", "motion detected", "sensors")
				if s.Switch["Nightlights"] {
					out <- nog.NewMessage("Marvin", "all nightlight", "motion detected")
				}
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
			if s.Switch["Nightlights"] {
				out <- nog.NewMessage("Marvin", "all off", "motion timeout")
			}
		}
	}
}

func (m *Motion) MotionSensor() bool {
	return m.motionChannel != nil
}
