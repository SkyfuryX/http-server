package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/SkyfuryX/http-server/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries *database.Queries
	platform string
}

type User struct {
	ID       uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) middlewareMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Print("Error connecting to DB")
		os.Exit(1)
	}

	apiCfg := apiConfig{
		dbQueries: database.New(db),
		platform: os.Getenv("PLATFORM"),
	}
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	handler := http.FileServer(http.Dir("."))
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetrics(handler)))
	mux.HandleFunc("GET /api/healthz", handlerReady)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidate)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	

	err = server.ListenAndServe()
	if err != nil {
		return
	}
}