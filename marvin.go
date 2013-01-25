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
	transition := term.(hu.Tuple)[0].(hu.String)
	h := environment.Get(hu.Symbol("hue"))
	h.(*hue).Do(transition.String())
	return nil
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
	environment.AddPrimitive("do_hue_transition", do_hue_transition)
	environment.AddPrimitive("listen_and_serve", ListenAndServe)
	environment.AddPrimitive("listen", Listen)

	if f, err := os.Open(flag.Arg(0)); err == nil {
		d := hu.ReadDocument(bufio.NewReader(f))

		for _, p := range d {
			if strings.HasPrefix(p.String(), "§ init") {
				for _, e := range p.(hu.Part)[1].(hu.Part) {
					if e, ok := e.(hu.Part); ok {
						a := hu.Application(e[0 : len(e)-1])
						environment.Evaluate(a)
					}
				}
			} else {
				log.Printf("part: %t %v\n", p, p)
			}
		}
	} else {
		log.Fatal("err:", err)
	}

	reader := bufio.NewReader(os.Stdin)

	var result hu.Term
	fmt.Printf("marvin> ")
	for {
		expression := hu.Read(reader)
		if expression != nil {
			if expression == hu.Symbol("\n") {
				if result != nil {
					fmt.Fprintf(os.Stdout, "%v\n", result)
				}
				fmt.Printf("marvin> ")
				continue
			} else {
				result = environment.Evaluate(expression)
			}
		} else {
			fmt.Fprintf(os.Stdout, "Goodbye!\n")
			break
		}
	}
}
