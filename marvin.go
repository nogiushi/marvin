package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type activity struct {
	Name string
	Next map[string]bool
}

type Marvin struct {
	Hue            hue
	Activities     map[string]*activity
	Activity       string
	Motion         bool
	DayLight       bool
	LastTransition string
	Present        map[string]bool
	Switch         map[string]bool
	Schedule       schedule
	//
	Transitions map[string]struct {
		Switch map[string]bool
	}

	do            chan string
	cond          *sync.Cond // a rendezvous point for goroutines waiting for or announcing state changed
	lightSensor   *TSL2561
	motionChannel <-chan bool
	lightChannel  <-chan int
	path          string
}

func NewMarvinFromFile(path string) (*Marvin, error) {
	var marvin Marvin
	marvin.path = path
	if j, err := os.OpenFile(marvin.path, os.O_RDONLY, 0666); err == nil {
		dec := json.NewDecoder(j)
		if err = dec.Decode(&marvin); err != nil {
			return nil, err
		}
		j.Close()
	} else {
		return nil, err
	}
	return &marvin, nil
}

func (m *Marvin) MotionSensor() bool {
	return m.motionChannel != nil
}

func (m *Marvin) LightSensor() bool {
	return m.lightChannel != nil
}

func (m *Marvin) GetActivity(name string) *activity {
	if name != "" {
		a, ok := m.Activities[name]
		if !ok {
			a = &activity{name, map[string]bool{}}
			m.Activities[name] = a
		}
		return a
	} else {
		return nil
	}
}

func (m *Marvin) loop() {
	m.Hue.Do("startup")
	m.StateChanged()
	if m.Switch == nil {
		m.Switch = make(map[string]bool)
	}
	if m.Activities == nil {
		m.Activities = make(map[string]*activity)
	}
	if m.Present == nil {
		m.Present = make(map[string]bool)
	}
	m.do = make(chan string, 2)

	var scheduledEventsChannel <-chan event
	if c, err := m.Schedule.Run(); err == nil {
		scheduledEventsChannel = c
	} else {
		log.Println("Warning: Scheduled events off:", err)
	}

	var dayLightTime time.Time
	if t, err := NewTSL2561(1, ADDRESS_FLOAT); err == nil {
		m.lightSensor = t
		m.lightChannel = t.Broadband()
	} else {
		log.Println("Warning: Light sensor off: ", err)
	}

	if c, err := GPIOInterrupt(7); err == nil {
		m.motionChannel = c
	} else {
		log.Println("Warning: Motion sensor off:", err)
	}
	var motionTimer *time.Timer
	var motionTimeout <-chan time.Time

	presenceChannel := Listen(m.Present)

	notifyChannel := make(chan os.Signal, 1)
	signal.Notify(notifyChannel, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)

	for {
		select {
		case what := <-m.do:
			log.Println("Do:", what)
			v, ok := m.Transitions[what]
			if ok {
				for k, v := range v.Switch {
					m.Switch[k] = v
				}
			}
			m.LastTransition = what
			m.Hue.Do(what)
			m.StateChanged()
		case e := <-scheduledEventsChannel:
			if m.Switch["Schedule"] {
				m.do <- e.What
			}
		case light := <-m.lightChannel:
			if time.Since(dayLightTime) > time.Duration(60*time.Second) {
				if light > 5000 && (m.DayLight != true) {
					m.DayLight = true
					dayLightTime = time.Now()
					m.StateChanged()
					if m.Switch["Daylights"] {
						m.do <- "daylight"
					}
				} else if light < 4900 && (m.DayLight != false) {
					m.DayLight = false
					dayLightTime = time.Now()
					m.StateChanged()
					if m.Switch["Daylights"] {
						m.do <- "daylight off"
					}
				}
			}
		case motion := <-m.motionChannel:
			if motion {
				const duration = 60 * time.Second
				if motionTimer == nil {
					m.Motion = true
					m.StateChanged()
					motionTimer = time.NewTimer(duration)
					motionTimeout = motionTimer.C // enable motionTimeout case
					if m.Switch["Nightlights"] {
						m.do <- "all nightlight"
					}
				} else {
					motionTimer.Reset(duration)
				}
				go postStatCount("motion", 1)
			}
		case <-motionTimeout:
			m.Motion = false
			m.StateChanged()
			motionTimer = nil
			motionTimeout = nil
			if m.Switch["Nightlights"] {
				m.do <- "all off"
			}
		case p := <-presenceChannel:
			if m.Present[p.name] != p.status {
				m.Present[p.name] = p.status
				m.StateChanged()
			}
		case <-time.NewTicker(60 * time.Second).C:
			if err := m.Save(m.path); err == nil {
				log.Println("saved:", m.path)
			} else {
				log.Println("ERROR: saving", err)
			}
		case sig := <-notifyChannel:
			log.Println("handling:", sig)
			goto Done
		}
	}
Done:
	if err := m.Save(m.path); err == nil {
		log.Println("saved:", m.path)
	} else {
		log.Println("ERROR: saving config", err)
	}
}

func (m *Marvin) getStateCond() *sync.Cond {
	if m.cond == nil {
		m.cond = sync.NewCond(&sync.Mutex{})
	}
	return m.cond
}

func (m *Marvin) StateChanged() {
	m.Hue.GetState()
	c := m.getStateCond()
	c.L.Lock()
	c.Broadcast()
	c.L.Unlock()
}

func (m *Marvin) WaitStateChanged() {
	c := m.getStateCond()
	c.L.Lock()
	c.Wait()
	c.L.Unlock()
}

func (m *Marvin) Save(path string) error {
	if j, err := os.Create(path); err == nil {
		dec := json.NewEncoder(j)
		var c Marvin = *m
		if err = dec.Encode(&c); err != nil {
			return err
		}
		j.Close()
	} else {
		return err
	}
	return nil
}
