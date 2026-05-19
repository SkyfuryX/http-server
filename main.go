package main

import (
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	apiCfg := apiConfig{}
	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	handler := http.FileServer(http.Dir("."))
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetrics(handler)))
	mux.HandleFunc("GET /healthz", handlerReady)
	mux.HandleFunc("GET /metrics", apiCfg.handlerHits)
	mux.HandleFunc("POST /reset", apiCfg.handlerReset)

	err := server.ListenAndServe()
	if err != nil {
		return
	}
}
