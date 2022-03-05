package main

import (
	"errors"
	"genredetector/server"
	"log"
	"os"
)

var ErrParseForm = errors.New("error parsing incoming request form")

func main() {
	port := os.Getenv("PORT")

	s := server.NewServer("localhost:"+port, false)
	log.Printf("listening on %s", s.Addr)
	log.Fatal(s.ListenAndServe())
}
