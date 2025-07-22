package apiconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Roddyck/Chirpy/internal/auth"
	"github.com/Roddyck/Chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

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
		Token: refreshToken,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID: user.ID,
		ExpiresAt: time.Now().Add(time.Hour*24*60),
	})

	if err != nil {
		respondWithError(w, 500, "error saving refresh token to db")
	    return
	}

	resp := respParams{
		User:        dbUserToUser(user),
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}

	respondeWithJSON(w, 200, resp)
}

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
	fmt.Println("UserId:", userID)
	fmt.Println("Token:", token)
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

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorParams struct {
		Error string `json:"error"`
	}
	resp := errorParams{
		Error: msg,
	}
	dat, err := json.Marshal(resp)
	if err != nil {
		log.Printf("error marshalling JSON: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}


func respondeWithJSON(w http.ResponseWriter, code int, payload any) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error marshalling JSON: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func replaceProfanity(input string) string {
	badWords := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	inputSplit := strings.Split(input, " ")
	for i, word := range inputSplit {
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				inputSplit[i] = "****"
			}
		}
	}

	return strings.Join(inputSplit, " ")
}

func dbUserToUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
}

func dbChirpToChirp(dbChirp database.Chirp) Chirp {
	return Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
}
