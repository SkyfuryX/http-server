package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	db "github.com/SkyfuryX/http-server/internal/database"
	"github.com/google/uuid"
)

func handlerReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(
		`<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`,
		cfg.fileserverHits.Load())))
}

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
	newUser, err := cfg.dbQueries.CreateUser(r.Context(), db.CreateUserParams{
		ID: uuid.New(),
		Email: user.Email,
	})
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	user.CreatedAt = newUser.CreatedAt
	user.UpdatedAt = newUser.UpdatedAt
	user.ID = newUser.ID
	respondWithJSON(w, 201, User{
		ID: newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email: newUser.Email,
	})
}

func handlerValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	params := parameters{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
	} else if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
	} else {
		type response struct {
			Valid        bool   `json:"valid"`
			Cleaned_Body string `json:"cleaned_body"`
		}
		respondWithJSON(w, 200, response{Valid: true, Cleaned_Body: cleanText(params.Body)})
	}
}
func cleanText(text string) string {
	textSplit := strings.Split(text, " ")
	for i, text := range textSplit {
		if strings.ToLower(text) == "kerfuffle" || strings.ToLower(text) == "sharbert" || strings.ToLower(text) == "fornax" { //"bad words" boot wanted filtered out.
			textSplit[i] = "****"
		}
	}
	joined := strings.Join(textSplit, " ")
	return joined
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	w.Write([]byte(msg))
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}
