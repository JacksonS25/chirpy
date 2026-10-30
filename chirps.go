package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JacksonS25/chirpy/internal/auth"
	"github.com/JacksonS25/chirpy/internal/database"
)

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	params.Body = profanityCleanse(params.Body)

	createdChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: userId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	filteredChirp := chirp{
		ID:        createdChirp.ID,
		CreatedAt: createdChirp.CreatedAt,
		UpdatedAt: createdChirp.UpdatedAt,
		Body:      createdChirp.Body,
		UserID:    createdChirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, filteredChirp)
}

func profanityCleanse(s string) string {
	words := strings.Split(s, " ")
	for i, word := range words {
		switch strings.ToLower(word) {
		case "kerfuffle", "sharbert", "fornax":
			words[i] = "****"
		default:
		}
	}
	return strings.Join(words, " ")
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	author_id := r.URL.Query().Get("author_id")
	author_uuid, err := uuid.Parse(author_id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse author id", err)
		return
	}

	var chirps []database.Chirp
	if author_id != "" {
		chirps, err = cfg.db.GetAuthorsChirps(r.Context(), author_uuid)
	} else {
		chirps, err = cfg.db.GetAllChirps(r.Context())
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps", err)
		return
	}

	retrievedChirps := []chirp{}

	for _, c := range chirps {
		retrievedChirps = append(retrievedChirps, chirp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, retrievedChirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	path, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid ID", err)
		return
	}

	retrievedChirp, err := cfg.db.GetChirp(r.Context(), path)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "No chirp found", err)
		return
	}

	returnChirp := chirp{
		ID:        retrievedChirp.ID,
		CreatedAt: retrievedChirp.CreatedAt,
		UpdatedAt: retrievedChirp.UpdatedAt,
		Body:      retrievedChirp.Body,
		UserID:    retrievedChirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, returnChirp)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	path, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid ID", err)
		return
	}

	chirpToDelete, err := cfg.db.GetChirp(r.Context(), path)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "No chirp found", err)
		return
	}

	if chirpToDelete.UserID != userId {
		respondWithError(w, http.StatusForbidden, "Forbidden", err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), path)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "No chirp found", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
