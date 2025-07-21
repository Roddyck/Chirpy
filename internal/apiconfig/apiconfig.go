package apiconfig

import (
	"net/http"
	"sync/atomic"

	"github.com/Roddyck/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	Platform       string
}

func (cfg *apiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func New(dbQueries *database.Queries, platform string) *apiConfig {
	return &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		Platform: platform,
	}
}
