package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
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
}

type scheduler struct {
	Hue      hue
	Schedule []event
	c        chan event
}

func NewSchedulerFromJSON(j io.Reader) (err error, s *scheduler) {
	s = &scheduler{}
	dec := json.NewDecoder(j)
	if err = dec.Decode(s); err == nil {
		for _, item := range s.Schedule {
			if err = s.schedule(item); err != nil {
				return err, nil
			}
		}
	}
	s.c = make(chan event, 1)
	return err, s
}

func (s *scheduler) schedule(e event) (err error) {
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
		s.maybeRun(time.Now(), e)
		for t := range time.NewTicker(duration).C {
			s.maybeRun(t, e)
		}
	}()
	return
}

var WEEKDAYS = []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}

var HOLIDAYS = map[string]string{"Christmas Day": "2012-12-25",	"New Year's Day": "2013-01-01",	"Birthday of Martin Luther King, Jr.": "2013-01-21", "Washington's Birthday": "2013-02-18", "Memorial Day": "2013-05-27", "Independence Day": "2013-07-04", "Labor Day": "2013-09-02", "Columbus Day": "2013-10-14", "Veterans Day": "2013-11-11", "Thanksgiving Day": "2013-11-28", "Christmas Day 2013": "2013-12-25"}


func (s *scheduler) maybeRun(t time.Time, e event) {
	t = t.In(time.Local)
	run := false
	if e.On == "" {
		run = true
	} else if e.On == "weekdays" {
		d := t.Weekday().String()
		for _, wd := range WEEKDAYS  {
			if d == wd {
				run = true
				break
			}
		}
	} else if e.On == "weekends" {
		d := t.Weekday().String()
		for _, wd := range []string{"Saturday", "Sunday"}  {
			if d == wd {
				run = true
				break
			}
		}
	}
	if e.ExceptOn == "" {

	} else if e.ExceptOn == "holidays" {
		s := t.Format("2006-01-02")
		for _, v := range HOLIDAYS  {
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
		s.c <- e
	}
}

func (s *scheduler) run() {
	for e := range s.c {
		s.Hue.Do(e.What)
	}
}

func NewSchedulerFromJSONPath(p string) (err error, s *scheduler) {
	if j, err := os.OpenFile(p, os.O_RDONLY, 0666); err == nil {
		defer j.Close()
		err, s = NewSchedulerFromJSON(j)
	}
	return err, s
}
