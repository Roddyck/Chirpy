package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/Roddyck/Chirpy/internal/apiconfig"
	"github.com/Roddyck/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)


func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	platform := os.Getenv("PLATFORM")
	tokenSecret := os.Getenv("TOKEN_SECRET")

	dbQueries := database.New(db)

	cfg := apiconfig.New(dbQueries, platform, tokenSecret)

	mux := http.NewServeMux()

	httpServer := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	fileserverHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", cfg.MiddlewareMetricsInc(fileserverHandler))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /api/chirps", cfg.HandleCreateChirp)
	mux.HandleFunc("POST /api/users", cfg.HandleCreateUser)
	mux.HandleFunc("GET /api/chirps", cfg.HandleListChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.HandleGetChirp)
	mux.HandleFunc("POST /api/login", cfg.HandleLogin)

	mux.HandleFunc("GET /admin/metrics", cfg.GetMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.Reset)

	log.Fatal(httpServer.ListenAndServe())
}
