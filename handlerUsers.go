package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SkyfuryX/http-server/internal/auth"
	db "github.com/SkyfuryX/http-server/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "Forbidden")
		return
	}

	err := cfg.dbQueries.ResetUsers(context.Background())
	if err != nil {
		fmt.Print(err)
		return
	}
	cfg.fileserverHits.Store(0)
	respondWithJSON(w, 200, "Users and Hits were Reset")
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondWithError(w, 400, "Bad Request")
		return
	}
	hash, err := auth.HashPassword(user.Password)
	if err != nil {
		respondWithError(w, 400, err.Error())
	}

	newUser, err := cfg.dbQueries.CreateUser(r.Context(), db.CreateUserParams{
		ID:             uuid.New(),
		HashedPassword: hash,
		Email:          user.Email,
	})
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	user.CreatedAt = newUser.CreatedAt
	user.UpdatedAt = newUser.UpdatedAt
	user.ID = newUser.ID
	respondWithJSON(w, 201, User{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	})
}

func (cgf *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondWithError(w, 400, "Bad Request")
		return
	}

	gotUser, err := cgf.dbQueries.GetUser(r.Context(), user.Email)
	if err != nil {
		respondWithError(w, 401, "User not found")
		return
	}

	valid, err := auth.CheckPWHash(user.Password, gotUser.HashedPassword)
	if err != nil || !valid {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	respondWithJSON(w, 200, User{
		ID:        gotUser.ID,
		CreatedAt: gotUser.CreatedAt,
		UpdatedAt: gotUser.UpdatedAt,
		Email:     gotUser.Email,
	})
}
