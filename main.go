package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	handler "github.com/JosueAD95/Server-course/handlers"
	db "github.com/JosueAD95/Server-course/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	const port = "8080"
	const filepath = "."

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	apiCfg := handler.ApiConfig{
		Db:          db.New(dbConn),
		Environment: os.Getenv("Environment"),
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(filepath)))))

	mux.HandleFunc("GET /admin/metrics", apiCfg.Metrics)

	mux.HandleFunc("POST /admin/reset", apiCfg.Reset)

	mux.HandleFunc("GET /api/healthz", handler.Healthz)

	mux.HandleFunc("POST /api/validate_chirp", handler.ValidateChirp)

	mux.HandleFunc("POST /api/users", apiCfg.AddUser)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	log.Fatal(server.ListenAndServe())
}
