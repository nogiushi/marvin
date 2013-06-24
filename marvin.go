package main

import (
	"log"
	"time"
)

type Marvin struct {
	Hue              hue
	Schedule         schedule
	ScheduleActive   bool
	NightlightActive bool
	DaylightActive   bool
	Transitions      []string
	dayLightSensor   *TSL2561
}

func (m *Marvin) Do(what string) {
	log.Println("Do:", what)
	if what == "sleep" {
		m.ScheduleActive = true
		m.DaylightActive = false
		m.NightlightActive = true
	} else if what == "sleep in" {
		m.ScheduleActive = false
		m.DaylightActive = false
		m.NightlightActive = true
		what = "sleep" // use the "sleep" hue transistion
	} else if what == "wake" {
		m.ScheduleActive = true
		m.DaylightActive = false
		m.NightlightActive = false
	} else if what == "movie" {
		m.ScheduleActive = false
		m.DaylightActive = false
		m.NightlightActive = false
	} else if what == "awake" {
		m.ScheduleActive = true
		m.DaylightActive = true
		m.NightlightActive = false
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
	} else if what == "deactivate daylights" {
		m.DaylightActive = false
		what = "chime"
	} else if what == "deactivate nightlights" {
		m.NightlightActive = false
		what = "chime"
	}
	m.Hue.Do(what)
}

func (m *Marvin) loop() {
	m.Do("startup")
	m.ScheduleActive = true

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
			if m.ScheduleActive {
				m.Do(e.What)
			}
		case dayLight := <-dayLightChannel:
			if m.DaylightActive {
				if dayLight {
					m.Do("daylight")
				} else {
					m.Do("daylight off")
				}
			}
		case motion := <-motionChannel:
			if motion {
				if m.NightlightActive {
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
			if m.NightlightActive {
				m.Do("all off")
				motionTimer = nil
				motionTimeout = nil
			}
		}
	}
}
