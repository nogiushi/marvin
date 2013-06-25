package main

import (
	"log"
	"sync"
	"time"
)

type Marvin struct {
	Hue         hue
	Schedule    schedule
	Transitions []string
	State       struct {
		Active         map[string]bool
		LastTransition string
	}
	cond           *sync.Cond // a rendezvous point for goroutines waiting for or announcing state changed
	dayLightSensor *TSL2561
}

func (m *Marvin) Do(what string) {
	log.Println("Do:", what)
	if what == "sleep" {
		m.State.Active["Schedule"] = true
		m.State.Active["Daylights"] = false
		m.State.Active["Nightlights"] = true
	} else if what == "sleep in" {
		m.State.Active["Schedule"] = false
		m.State.Active["Daylights"] = false
		m.State.Active["Nightlights"] = true
		what = "sleep" // use the "sleep" hue transistion
	} else if what == "wake" {
		m.State.Active["Schedule"] = true
		m.State.Active["Daylights"] = false
		m.State.Active["Nightlights"] = false
	} else if what == "movie" {
		m.State.Active["Schedule"] = false
		m.State.Active["Daylights"] = false
		m.State.Active["Nightlights"] = false
	} else if what == "awake" {
		m.State.Active["Schedule"] = true
		m.State.Active["Daylights"] = true
		m.State.Active["Nightlights"] = false
		if m.dayLightSensor != nil {
			if value, err := m.dayLightSensor.DayLightSingle(); err == nil {
				dayLight := value > 5000
				if dayLight {
					m.Do("daylight")
				} else {
					m.Do("daylight off")
				}
			} else {
				log.Println("error getting broadband value:", err)
			}
		}
	}
	m.State.LastTransition = what
	m.BroadcastStateChanged()
	go m.Hue.Do(what)
}

func (m *Marvin) loop() {
	m.Do("startup")

	var scheduledEventsChannel <-chan event
	if c, err := m.Schedule.Run(); err == nil {
		scheduledEventsChannel = c
	} else {
		log.Println("Warning: Scheduled events off:", err)
	}

	var dayLightChannel <-chan bool
	if t, err := NewTSL2561(1, ADDRESS_FLOAT); err == nil {
		m.dayLightSensor = t
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
			if m.State.Active["Schedule"] {
				m.Do(e.What)
			}
		case dayLight := <-dayLightChannel:
			if m.State.Active["Daylights"] {
				if dayLight {
					m.Do("daylight")
				} else {
					m.Do("daylight off")
				}
			}
		case motion := <-motionChannel:
			if motion {
				if m.State.Active["Nightlights"] {
					m.Do("all nightlight")
					const duration = 60 * time.Second
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
			if m.State.Active["Nightlights"] {
				m.Do("all off")
				motionTimer = nil
				motionTimeout = nil
			}
		}
	}
}

func (m *Marvin) getStateCond() *sync.Cond {
	if m.cond == nil {
		m.cond = sync.NewCond(&sync.Mutex{})
	}
	return m.cond
}

func (m *Marvin) BroadcastStateChanged() {
	m.getStateCond().Broadcast()
}

func (m *Marvin) WaitStateChanged() {
	c := m.getStateCond()
	c.L.Lock()
	c.Wait()
	c.L.Unlock()
}
