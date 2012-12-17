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

func (s *scheduler) maybeRun(t time.Time, e event) {
	t = t.In(time.Local)
	day, ok := e.Days[t.Weekday().String()]
	if ok {
		if day == "off" {
			log.Println(e.What + " at " + t.String() + " (marked as off for day)")
			return
		}
	}
	log.Println(e.What + " at " + t.String())
	e.time = t
	s.c <- e
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
