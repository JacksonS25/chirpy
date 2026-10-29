package main

import (
	"net/http"
	"time"

	"github.com/JacksonS25/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No Refresh Token", err)
		return
	}

	refresh_token, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	currentTime := time.Now().UTC()
	expiration := refresh_token.ExpiresAt
	if expiration.Before(currentTime) || refresh_token.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	authToken, err := auth.MakeJWT(refresh_token.UserID, cfg.secret, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to Create Token", err)
		return
	}

	type auth struct {
		Token string `json:"token"`
	}

	returnAuth := auth{
		Token: authToken,
	}

	respondWithJSON(w, http.StatusOK, returnAuth)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No Refresh Token", err)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
