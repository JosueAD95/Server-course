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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set ")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	apiCfg := handler.ApiConfig{
		Db:          db.New(dbConn),
		Environment: os.Getenv("Environment"),
		JWTSecret:   jwtSecret,
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(filepath)))))

	mux.HandleFunc("GET /admin/metrics", apiCfg.Metrics)

	mux.HandleFunc("POST /admin/reset", apiCfg.Reset)

	mux.HandleFunc("GET /api/healthz", handler.Healthz)

	mux.HandleFunc("POST /api/chirps", apiCfg.CreateChirp)

	mux.HandleFunc("GET /api/chirps", apiCfg.GetAllChirps)

	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirpById)

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.DeleteChirp)

	mux.HandleFunc("POST /api/users", apiCfg.AddUser)

	mux.HandleFunc("PUT /api/users", apiCfg.UpdateUserCredentials)

	mux.HandleFunc("POST /api/login", apiCfg.Login)

	mux.HandleFunc("POST /api/refresh", apiCfg.RefreshToken)

	mux.HandleFunc("POST /api/revoke", apiCfg.RevokeToken)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Fatal(server.ListenAndServe())

}
