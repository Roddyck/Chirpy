package apiconfig

import (
	"encoding/json"
	"net/http"

	"github.com/Roddyck/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) HandlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, 401, "no api key found in headers")
	    return
	}

	if apiKey != cfg.PolkaKey {
		respondWithError(w, 401, "wrong api key")
		return
	}

	params := parameters{} 
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 500, "error decoding request body")
	    return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, 500, "error parsing user id: invalid uuid")
	    return
	}
	err = cfg.db.UpgradeUser(r.Context(), userID)
	if err != nil {
		respondWithError(w, 404, "user not found")
	    return
	}

	w.WriteHeader(204)
}
