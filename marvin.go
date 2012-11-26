package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"time"
)

type schedule struct {
	When     string
	Interval string
	What     []command
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
	zone, _ := time.Now().Zone()
	on, err := time.Parse(time.Kitchen+" MST", e.When+" "+zone)
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
		t := time.NewTicker(duration)
		s.c <- event{time.Now(), e.What}
		for t := range t.C {
			s.c <- event{t, e.What}
		}
	}()
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
	f, err := os.OpenFile("/var/log/marvin.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)

	j, err := os.OpenFile("marvin.json", os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	
	s := NewSchedulerFromReader(j)
	j.Close()
	s.run()
}
