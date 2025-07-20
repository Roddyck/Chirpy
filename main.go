package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	httpServer := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	mux.Handle("GET /app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	log.Fatal(httpServer.ListenAndServe())
}
