package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/JacksonS25/chirpy/internal/auth"
	"github.com/JacksonS25/chirpy/internal/database"

	"github.com/google/uuid"
)

type user struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type userParameters struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := userParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	retrievedUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	createdUser := user{
		ID:        retrievedUser.ID,
		CreatedAt: retrievedUser.CreatedAt,
		UpdatedAt: retrievedUser.UpdatedAt,
		Email:     retrievedUser.Email,
	}

	respondWithJSON(w, http.StatusCreated, createdUser)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := userParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	retrievedUser, err := cfg.db.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid Email or Password", err)
		return
	}

	valid, err := auth.CheckPasswordHash(params.Password, retrievedUser.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't Validate", err)
		return
	}

	if !valid {
		respondWithError(w, http.StatusUnauthorized, "Invalid Email or Password", err)
		return
	}

	token, err := auth.MakeJWT(retrievedUser.ID, cfg.secret, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to Create Token", err)
		return
	}

	refresh_token := auth.MakeRefreshToken()

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refresh_token,
		UserID:    retrievedUser.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token", err)
		return
	}

	returnUser := user{
		ID:           retrievedUser.ID,
		CreatedAt:    retrievedUser.CreatedAt,
		UpdatedAt:    retrievedUser.UpdatedAt,
		Email:        retrievedUser.Email,
		Token:        token,
		RefreshToken: refresh_token,
	}

	respondWithJSON(w, http.StatusOK, returnUser)
}

func (cfg *apiConfig) handlerUpdate(w http.ResponseWriter, r *http.Request) {
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

	decoder := json.NewDecoder(r.Body)
	params := userParameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't Hash Password", err)
		return
	}

	updatedUser, err := cfg.db.UpdateUserByID(r.Context(), database.UpdateUserByIDParams{
		Email:          params.Email,
		HashedPassword: hashed_password,
		ID:             userId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't Update User", err)
		return
	}

	returnUser := user{
		ID:        updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Email:     updatedUser.Email,
	}

	respondWithJSON(w, http.StatusOK, returnUser)
}
