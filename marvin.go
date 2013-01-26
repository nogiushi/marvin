package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/eikeon/hu"
)

var StaticRoot *string
var Address *string

func do_hue_transition(environment *hu.Environment, term hu.Term) hu.Term {
	var terms []string
	for _, t := range term.(hu.Tuple) {
		terms = append(terms, t.String())
	}
	name := strings.Join(terms, " ")
	h := environment.Get(hu.Symbol("hue"))
	h.(*hue).Do(name)
	return hu.String("done")
}

func turn_scheduler(environment *hu.Environment, term hu.Term) (response hu.Term) {
	scheduler := environment.Get(hu.Symbol("scheduler")).(*scheduler)
	terms := term.(hu.Tuple)
	if len(terms) > 0 {
		v := terms[0].String() == "on"
		scheduler.DoNotDisturb = !v
	}
	if scheduler.DoNotDisturb {
		response = hu.String("scheduler is now off")
	} else {
		response = hu.String("scheduler is now on")
	}
	return
}

func main() {
	log.Println("starting marvin")

	Address = flag.String("address", ":9999", "http service address")
	StaticRoot = flag.String("root", "static", "...")
	config := flag.String("config", "/etc/marvin.json", "file path to configuration file")
	flag.Parse()

	environment := hu.NewEnvironment()

	if err, s := NewSchedulerFromJSONPath(*config); err == nil {
		environment.Define(hu.Symbol("scheduler"), s)
		environment.Define(hu.Symbol("hue"), &(s.Hue))
		go s.run()
	} else {
		log.Fatal(err)
	}
	environment.AddPrimitive("do_hue", do_hue_transition)
	environment.AddPrimitive("turn_scheduler", turn_scheduler)

	ListenAndServe(environment, nil)
	Listen(environment, nil)

	reader := bufio.NewReader(os.Stdin)

	var result hu.Term
	fmt.Printf("marvin> ")
	for {
		expression := hu.ReadSentence(reader)
		if expression != nil {
			log.Println("expression:", expression)
			a := hu.Application(expression)
			result = environment.Evaluate(a)
			fmt.Fprintf(os.Stdout, "%v\n", result)
			fmt.Printf("marvin> ")
		} else {
			fmt.Fprintf(os.Stdout, "So long and thanks for all the fish!\n")
			break
		}
	}
}
