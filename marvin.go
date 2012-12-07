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
	Schedules map[string]schedule
	c         chan event
}

func NewSchedulerFromReader(j io.Reader) *scheduler {
	s := &scheduler{}
	dec := json.NewDecoder(j)
	if err := dec.Decode(s); err != io.EOF {
		log.Println("incomplete decode")
	} else if err != nil {
		log.Fatal(err)
	}
	s.c = make(chan event, 1)

	for name, value := range s.Schedules {
		s.schedule(name, value)
	}
	return s
}

func (s *scheduler) schedule(name string, e schedule) {
	now := time.Now()
	zone, _ := now.Zone()
	on, err := time.Parse("2006-01-02 "+time.Kitchen+" MST",
		now.Format("2006-01-02 ")+e.When+" "+zone)
	if err != nil {
		log.Fatal(err)
	}
	duration, err := time.ParseDuration(e.Interval)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Println("scheduled '" + name + "' for: " + on.String())
		wait := time.Duration((on.UnixNano() - time.Now().UnixNano()) % int64(duration))
		if wait < 0 {
			wait += duration
		}
		log.Println("waiting for " + wait.String())
		time.Sleep(wait)
		s.maybeRun(time.Now(), e)
		for t := range time.NewTicker(duration).C {
			s.maybeRun(t, e)
		}
	}()
}

func (s *scheduler)maybeRun(t time.Time, sched schedule) {
	day, ok := sched.Days[t.In(time.Local).Weekday().String()]
	if ok {
		if day=="off" {
			return
		}
	}
	s.c <- event{t, sched.What}
}


func (s *scheduler) run() {
	log.Println("scheduler started on:" + time.Now().In(time.Local).String())
	// visual display of scheduler starting
	s.Hue.run(command{"/groups/0/action", "blink"})

	for e := range s.c {
		for _, command := range e.commands {
			s.Hue.run(command)
		}
	}
}

func main() {
	config := flag.String("config", "/etc/marvin.json", "file path to configuration file")
	logfile := flag.String("logfile", "", "file path to logfile")
	flag.Parse()

	if *logfile != "" {
		f, err := os.OpenFile("/var/log/marvin.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	j, err := os.OpenFile(*config, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	s := NewSchedulerFromReader(j)
	j.Close()
	s.run()
}
