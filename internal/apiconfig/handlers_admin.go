package apiconfig

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	msg := fmt.Sprintf(`
		<html>
		  <body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		  </body>
		</html>`,
		cfg.fileserverHits.Load())

	w.Write([]byte(msg))
}

func (cfg *apiConfig) Reset(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		respondWithError(w, 403, "you are not allowed here")
	}
	err := cfg.db.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("error deleting users from db: %v", err))
	}
	cfg.fileserverHits.Store(0)
	w.WriteHeader(200)
	w.Write([]byte("Hits reset to 0"))
}
