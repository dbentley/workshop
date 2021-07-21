package main

import (
	"log"

	"github.com/tilt-dev/workshop/wslocal"
)

func main() {
	l, err := wslocal.NewLooper()
	if err != nil {
		log.Fatal(err)
	}
	err = l.Loop()
	if err != nil {
		log.Fatal(err)
	}
}
