package main

import (
	"genredetector/server"
	"log"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := server.NewServer("localhost:"+port, false)
	log.Printf("listening on %s", s.Addr)
	log.Fatal(s.ListenAndServe())
}
