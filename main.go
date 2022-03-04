package main

import (
	"errors"
	"genredetector/server"
	"log"
)

var ErrParseForm = errors.New("error parsing incoming request form")

func main() {
	s := server.NewServer("localhost:8080", false)
	log.Printf("listening on %s", s.Addr)
	log.Fatal(s.ListenAndServe())
}
