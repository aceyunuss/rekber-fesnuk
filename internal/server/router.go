package server

import (
	"database/sql"
	"log"
	"net/http"
	"rekber-fesnuk/internal/config"
	"rekber-fesnuk/internal/handlers"
	"rekber-fesnuk/internal/middleware"
)

func NewRouter(cfg *config.Config, database *sql.DB) http.Handler {

	sellerHandler := handlers.NewSellerHandler(database, cfg.JWTSecret)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/sellers/login", sellerHandler.LoginSeller)

	mux.HandleFunc("POST /api/sellers", sellerHandler.CreateSeller)
	mux.HandleFunc("GET /api/sellers/{id}", middleware.ReqAuth(cfg.JWTSecret)(sellerHandler.GetSeller))
	mux.HandleFunc("PUT /api/sellers/{id}", middleware.ReqAuth(cfg.JWTSecret)(sellerHandler.UpdateSeller))
	mux.HandleFunc("DELETE /api/sellers/{id}", middleware.ReqAuth(cfg.JWTSecret)(sellerHandler.DeleteSeller))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"OK"}`))
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`Server is runyaw 😸`))
	})

	return logRequests(mux)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
