package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"time"
)

type schedule struct {
	Name     string
	When     string
	Interval string
	What     []command
	Days     map[string]string
}

type event struct {
	t        time.Time
	commands []command
}

type scheduler struct {
	Hue       hue
	Schedules []schedule
	c         chan event
}

func NewSchedulerFromJSON(j io.Reader) (err error, s *scheduler) {
	s = &scheduler{}
	dec := json.NewDecoder(j)
	if err = dec.Decode(s); err == nil {
		for _, value := range s.Schedules {
			if err = s.schedule(value); err != nil {
				return err, nil
			}
		}
	}
	s.c = make(chan event, 1)
	return err, s
}

func (s *scheduler) schedule(e schedule) (err error) {
	var on time.Time
	var duration time.Duration
	name := e.Name
	now := time.Now()
	zone, _ := now.Zone()
	if on, err = time.Parse("2006-01-02 "+time.Kitchen+" MST", now.Format("2006-01-02 ")+e.When+" "+zone); err != nil {
		log.Println("could not parse when of '" + e.When + "' for " + name)
		return
	}
	if duration, err = time.ParseDuration(e.Interval); err != nil {
		log.Println("could not parse interval of '" + e.Interval + "' for " + name)
		return
	}

	go func() {
		log.Println("scheduled '" + name + "' for: " + on.String())
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

func (s *scheduler) maybeRun(t time.Time, sched schedule) {
	day, ok := sched.Days[t.In(time.Local).Weekday().String()]
	if ok {
		if day == "off" {
			return
		}
	}
	s.c <- event{t, sched.What}
}

func (s *scheduler) run() {
	for e := range s.c {
		for _, command := range e.commands {
			s.Hue.run(command)
		}
	}
}

func NewSchedulerFromJSONPath(p string) (err error, s *scheduler) {
	if j, err := os.OpenFile(p, os.O_RDONLY, 0666); err == nil {
		defer j.Close()
		err, s = NewSchedulerFromJSON(j)
	}
	return err, s
}

func main() {
	log.Println("starting marvin")

	config := flag.String("config", "/etc/marvin.json", "file path to configuration file")
	logfile := flag.String("logfile", "", "file path to logfile (defaults to stderr)")
	flag.Parse()

	if *logfile != "" {
		if f, err := os.OpenFile(*logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err == nil {
			defer f.Close()
			log.SetOutput(f)
		} else {
			log.Fatal(err)
		}
	}

	if err, s := NewSchedulerFromJSONPath(*config); err == nil {
		s.Hue.run(command{"/groups/0/action", "blink"}) // visual display of scheduler starting
		s.run()
	} else {
		log.Fatal(err)
	}
}
