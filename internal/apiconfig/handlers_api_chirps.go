package apiconfig

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Roddyck/Chirpy/internal/auth"
	"github.com/Roddyck/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type chirpParams struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid authentication token")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.TokenSecret)
	if err != nil {
		fmt.Printf("error validating jwt: %v", err)
		respondWithError(w, 401, "invalid authentication token")
		return
	}

	params := chirpParams{}
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("error decoding request body: %v", err))
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("error creating chirp: %v", err))
		return
	}

	respondeWithJSON(w, 201, dbChirpToChirp(chirp))
}

func (cfg *apiConfig) HandleListChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.ListChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("error retriveing chirps from db: %v", err))
		return
	}

	resp := []Chirp{}
	for _, chirp := range chirps {
		resp = append(resp, dbChirpToChirp(chirp))
	}

	respondeWithJSON(w, 200, resp)
}

func (cfg *apiConfig) HandleGetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("error parsing id from url param: %v", err))
	}
	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}

	respondeWithJSON(w, 200, dbChirpToChirp(chirp))
}

func (cfg *apiConfig) HandleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "token missing from headers")
		return
	}

	chirpID := r.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, 404, "invalid chirp id")
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.TokenSecret)
	if err != nil {
		respondWithError(w, 401, "invalid access token")
	    return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, 404, "chirp not found")
	    return
	}

	if chirp.UserID != userID {
		respondWithError(w, 403, "you are not the author of that chirp")
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, 500, "error deleting chirp")
	    return
	}

	w.WriteHeader(204)
}
