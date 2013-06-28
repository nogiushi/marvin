package main

import (
	"log"
	"sync"
	"time"
)

type lightTime struct {
	value bool
	time  time.Time
}

type Marvin struct {
	Hue      hue
	Schedule schedule
	//
	Transitions map[string]struct {
		Active map[string]bool
	}
	//
	State struct {
		Active         map[string]bool
		LastTransition string
	}

	Do          chan string
	cond        *sync.Cond // a rendezvous point for goroutines waiting for or announcing state changed
	lightSensor *TSL2561
}

func (m *Marvin) loop() {
	m.Do = make(chan string, 2)
	m.Do <- "startup"

	var scheduledEventsChannel <-chan event
	if c, err := m.Schedule.Run(); err == nil {
		scheduledEventsChannel = c
	} else {
		log.Println("Warning: Scheduled events off:", err)
	}

	var lightChannel <-chan int
	if t, err := NewTSL2561(1, ADDRESS_FLOAT); err == nil {
		m.lightSensor = t
		lightChannel = t.Broadband()
	} else {
		log.Println("Warning: Light sensor off: ", err)
	}
	var lastLight *lightTime

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
		case what := <-m.Do:
			log.Println("Do:", what)
			v, ok := m.Transitions[what]
			if ok {
				for k, v := range v.Active {
					m.State.Active[k] = v
				}
			}
			m.State.LastTransition = what
			m.BroadcastStateChanged()
			go m.Hue.Do(what)
		case e := <-scheduledEventsChannel:
			if m.State.Active["Schedule"] {
				m.Do <- e.What
			}
		case light := <-lightChannel:
			if m.State.Active["Daylights"] {
				if lastLight == nil || time.Since(lastLight.time) > time.Duration(60*time.Second) {
					if light > 5000 && (lastLight == nil || lastLight.value != true) {
						lastLight = &lightTime{true, time.Now()}
						m.Do <- "daylight"
					} else if light < 4900 && (lastLight == nil || lastLight.value != false) {
						lastLight = &lightTime{false, time.Now()}
						m.Do <- "daylight off"
					}
				}
			} else {
				lastLight = nil
			}
		case motion := <-motionChannel:
			if motion {
				if m.State.Active["Nightlights"] {
					m.Do <- "all nightlight"
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
				m.Do <- "all off"
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
