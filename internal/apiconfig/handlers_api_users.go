package apiconfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Roddyck/Chirpy/internal/auth"
	"github.com/Roddyck/Chirpy/internal/database"
)

func (cfg *apiConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	params := parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("error decoding request body: %v", err))
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("error hashing password: %v", err))
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		HashedPassword: hash,
		Email:          params.Email,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("error adding user to db: %v", err))
		return
	}

	respondeWithJSON(w, 201, dbUserToUser(user))
}

func (cfg *apiConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type respParams struct {
		User
		AccessToken  string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	params := parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("error decoding request body: %v", err))
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 401, "incorrect password or email")
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.TokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("error making authentication token: %v", err))
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("error making refresh token: %v", err))
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 401, "incorrect password or email")
		return
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})

	if err != nil {
		respondWithError(w, 500, "error saving refresh token to db")
		return
	}

	resp := respParams{
		User:         dbUserToUser(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	respondeWithJSON(w, 200, resp)
}

func (cfg *apiConfig) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type respParams struct {
		User
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "access token is missing")
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.TokenSecret)
	if err != nil {
		respondWithError(w, 401, "invalid access token")
		return
	}

	params := parameters{}
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("error decoding request body: %v", err))
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "error hashing password")
		return
	}

	dbUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		HashedPassword: hash,
		Email:          params.Email,
	})

	respondeWithJSON(w, 200, dbUserToUser(dbUser))
}
