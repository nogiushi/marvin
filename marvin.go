package main

import (
	"log"
	"time"
)

type Marvin struct {
	Hue          hue
	Schedule     schedule
	DoNotDisturb bool
	Sleeping     bool
	c            chan event
}

func (m *Marvin) Do(what string) {
	log.Println("Do:", what)
	if what == "sleep" {
		m.Sleeping = true
		what = "all off"
	} else if what == "dawn" {
		m.Sleeping = false
	}
	m.Hue.Do(what)
}

func (m *Marvin) loop() {
	m.Do("chime") // visual display of marvin starting

	var scheduledEventsChannel <-chan event
	if c, err := m.Schedule.Run(); err == nil {
		scheduledEventsChannel = c
	} else {
		log.Println("Warning: Scheduled events off:", err)
	}

	var dayLightChannel <-chan bool
	if t, err := NewTSL2561(1, ADDRESS_FLOAT); err == nil {
		dayLightChannel = t.DayLight()
	} else {
		log.Println("Warning: Daylight sensor off: ", err)
	}

	var motionChannel <-chan bool
	if c, err := GPIOInterrupt(7); err == nil {
		motionChannel = c
	} else {
		log.Println("Warning: Motion sensor off:", err)
	}
	var motionTimer *time.Timer
	var motionTimeout <-chan time.Time

	for {
		select {
		case e := <-scheduledEventsChannel:
			if m.DoNotDisturb == false {
				m.Do(e.What)
			}
		case dayLight := <-dayLightChannel:
			if dayLight {
				m.Do("daylight")
			} else {
				m.Do("daylight off")
			}
		case motion := <-motionChannel:
			if motion {
				if m.Sleeping {
					m.Do("all nightlight")
					const duration = 30 * time.Second
					if motionTimer == nil {
						motionTimer = time.NewTimer(duration)
						motionTimeout = motionTimer.C // enable motionTimeout case
					} else {
						motionTimer.Reset(duration)
					}
				}
				go postStatCount("motion", 1)
			}
		case <-motionTimeout:
			if m.Sleeping {
				m.Do("all off")
				motionTimer = nil
				motionTimeout = nil
			}
		}
	}
}
