package apiconfig

import (
	"net/http"
	"time"

	"github.com/Roddyck/Chirpy/internal/auth"
)

func (cfg *apiConfig) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	type respParams struct {
		Token string `json:"token"`
	}

	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "refresh token wasn't provided")
		return
	}

	dbToken, err := cfg.db.GetRefreshToken(r.Context(), refresh_token)
	if err != nil {
		respondWithError(w, 401, "no refresh token in db for given user")
		return
	}

	if dbToken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, 401, "refresh token expired")
		return
	}

	if dbToken.RevokedAt.Valid {
		respondWithError(w, 401, "refresh token expired")
		return
	}

	accessToken, err := auth.MakeJWT(dbToken.UserID, cfg.TokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, 500, "error creating access token")
		return
	}

	resp := respParams{
		Token: accessToken,
	}

	respondeWithJSON(w, 200, resp)
}

func (cfg *apiConfig) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "refresh token wasn't provided")
		return
	}

	err = cfg.db.RevokeToken(r.Context(), refresh_token)
	if err != nil {
		respondWithError(w, 500, "error revoking refresh token")
		return
	}

	w.WriteHeader(204)
}
