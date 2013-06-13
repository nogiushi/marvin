package main

import (
	"log"

	"github.com/stathat/go"
)

func postStat(name string, value float64) {
	if err := stathat.PostEZValue(name, "eikeon@eikeon.com", value); err != nil {
		log.Printf("error posting value %v: %d", err, value)
	}
}

func postStatCount(name string, value int) {
	if err := stathat.PostEZCount(name, "eikeon@eikeon.com", value); err != nil {
		log.Printf("error posting value %v: %d", err, value)
	}
}
