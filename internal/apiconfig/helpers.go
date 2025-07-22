package apiconfig

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

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
