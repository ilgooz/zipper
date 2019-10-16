package main

import (
	"flag"
	"log"
	"time"
)

const (
	gracefulShutdownTimeout = 1 * time.Minute
)

var addr = flag.String("addr", ":8080", "network addr to listen from")

func main() {
	flag.Parse()

	s, err := New(*addr, gracefulShutdownTimeout)
	if err != nil {
		log.Fatal(err)
	}
	if err := s.GracefulStart(); err != nil {
		log.Fatal(err)
	}
}
