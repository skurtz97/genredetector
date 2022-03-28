package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func method(method string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "invalid method", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
func get(next http.HandlerFunc) http.HandlerFunc {
	return method(http.MethodGet, next)
}
func post(next http.HandlerFunc) http.HandlerFunc {
	return method(http.MethodPost, next)
}
func put(next http.HandlerFunc) http.HandlerFunc {
	return method(http.MethodPut, next)
}
func delete(next http.HandlerFunc) http.HandlerFunc {
	return method(http.MethodDelete, next)
}

func logging(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		end := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), end.Sub(start))
	}
	return http.HandlerFunc(fn)
}

func recovery(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func HandleEcho(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if !json.Valid(body) {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(body)
	if err != nil {
		fmt.Println(err)
	}
}
func main() {
	log.Println("Server started and listening on port 8080")
	http.HandleFunc("/echo", logging(method("POST", HandleEcho)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
