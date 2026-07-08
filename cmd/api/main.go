package main

import (
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"rekber-fesnuk/internal/config"
	"rekber-fesnuk/internal/db"
	"rekber-fesnuk/internal/server"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	database, err := db.Connect(cfg.DbUrl)
	if err != nil {
		log.Fatalf("Failed connect database : %v", err)
	}
	defer database.Close()
	log.Println("Database connected successfully")

	log.Println("Starting server on port : ", cfg.Port)

	router := server.NewRouter(cfg, database)

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal("Server error:", err)
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
