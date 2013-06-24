package main

import (
	"log"
	"time"
)

type event struct {
	When     string
	Interval string
	What     string
	On       string
	ExceptOn string
	Days     map[string]string
	time     time.Time
	c        chan event
}

func (e event) schedule(c chan event) (err error) {
	e.c = c
	var on time.Time
	var duration time.Duration
	now := time.Now()
	zone, _ := now.Zone()
	if on, err = time.Parse("2006-01-02 "+time.Kitchen+" MST", now.Format("2006-01-02 ")+e.When+" "+zone); err != nil {
		log.Println("could not parse when of '" + e.When + "' for " + e.What)
		return
	}
	if duration, err = time.ParseDuration(e.Interval); err != nil {
		log.Println("could not parse interval of '" + e.Interval + "' for " + e.What)
		return
	}

	go func() {
		log.Println("scheduled '" + e.What + "' for: " + on.String())
		wait := time.Duration((on.UnixNano() - time.Now().UnixNano()) % int64(duration))
		if wait < 0 {
			wait += duration
		}
		time.Sleep(wait)
		e.maybeRun(time.Now())
		for t := range time.NewTicker(duration).C {
			e.maybeRun(t)
		}
	}()
	return
}

var WEEKDAYS = []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}

var HOLIDAYS = map[string]string{"Christmas Day": "2012-12-25", "New Year's Day": "2013-01-01", "Birthday of Martin Luther King, Jr.": "2013-01-21", "Washington's Birthday": "2013-02-18", "Memorial Day": "2013-05-27", "Independence Day": "2013-07-04", "Labor Day": "2013-09-02", "Columbus Day": "2013-10-14", "Veterans Day": "2013-11-11", "Thanksgiving Day": "2013-11-28", "Christmas Day 2013": "2013-12-25"}

func (e event) maybeRun(t time.Time) {
	t = t.In(time.Local)
	run := false
	if e.On == "" {
		run = true
	} else if e.On == "weekdays" {
		d := t.Weekday().String()
		for _, wd := range WEEKDAYS {
			if d == wd {
				run = true
				break
			}
		}
	} else if e.On == "weekends" {
		d := t.Weekday().String()
		for _, wd := range []string{"Saturday", "Sunday"} {
			if d == wd {
				run = true
				break
			}
		}
	}
	if e.ExceptOn == "" {

	} else if e.ExceptOn == "holidays" {
		s := t.Format("2006-01-02")
		for _, v := range HOLIDAYS {
			if s == v {
				log.Println("not running due to holiday:", e)
				run = false
				break
			}
		}
	}
	if run {
		log.Println(e.What + " at " + t.String())
		e.time = t
		e.c <- e
	}
}

type schedule []event

func (s schedule) Run() (chan event, error) {
	eventsCh := make(chan event, 1)
	for _, e := range s {
		if err := e.schedule(eventsCh); err != nil {
			return nil, err
		}
	}
	return eventsCh, nil
}
