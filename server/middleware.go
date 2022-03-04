package server

import "net/http"

func middlewareCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://genredetector.com")
		next.ServeHTTP(w, r)
	})
}
