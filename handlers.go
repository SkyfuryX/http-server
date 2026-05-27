package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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
		ID:    uuid.New(),
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
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	})
}

type Chirp struct {
	ID         uuid.UUID `json:"id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Body       string    `json:"body"`
	User_id    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerAddChirp(w http.ResponseWriter, r *http.Request) {
	var chirp Chirp
	err := json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if len(chirp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	newChirp, err := cfg.dbQueries.CreateChirp(r.Context(), db.CreateChirpParams{
		ID:     uuid.New(),
		Body:   chirp.Body,
		UserID: chirp.User_id,
	})
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	respondWithJSON(w, 201, Chirp{
		ID:         newChirp.ID,
		Created_at: newChirp.CreatedAt,
		Updated_at: newChirp.UpdatedAt,
		Body:       cleanText(newChirp.Body),
		User_id:    newChirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 404, "Invalid User ID")
		return
	}

	getChirp, err := cfg.dbQueries.GetChirp(r.Context(), ID)
	if err != nil {
		respondWithError(w, 404, "Chirp not found")
		return
	}
	respondWithJSON(w, 200, Chirp{
		ID:         getChirp.ID,
		Created_at: getChirp.CreatedAt,
		Updated_at: getChirp.UpdatedAt,
		Body:       getChirp.Body,
		User_id:    getChirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	data, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	var chirps []Chirp
	for _, row := range data {
		chirps = append(chirps, Chirp{
			ID:         row.ID,
			Created_at: row.CreatedAt,
			Updated_at: row.UpdatedAt,
			Body:       row.Body,
			User_id:    row.UserID,
		})
	}

	respondWithJSON(w, 200, chirps)
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
