package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/JosueAD95/Server-course/internal/auth"
	database "github.com/JosueAD95/Server-course/internal/database"
	model "github.com/JosueAD95/Server-course/models"
	util "github.com/JosueAD95/Server-course/utils"
)

func (cfg ApiConfig) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")

	dbChirps, err := cfg.Db.GetChirps(r.Context())
	if err != nil {
		log.Printf("Error retriaving all chirps: %s", err)
		w.WriteHeader(500)
		return
	}

	chirps := make([]model.Chirp, len(dbChirps))
	for i, c := range dbChirps {
		chirps[i].MapDBChirp(c)
	}

	data, err := json.Marshal(chirps)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (cfg ApiConfig) GetChirpById(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/json")
	chirpId := r.PathValue("chirpID")
	err := uuid.Validate(chirpId)
	if err != nil {
		log.Printf("Error validating Chirp ID (%s): %s", chirpId, err)
		w.WriteHeader(500)
		return
	}
	id := uuid.MustParse(r.PathValue("chirpID"))
	dbChirp, err := cfg.Db.GetChirpById(r.Context(), id)
	if err != nil {
		log.Printf("Error retriaving chirp : %s", err)
		w.WriteHeader(404)
		return
	}

	chirp := model.Chirp{}
	chirp.MapDBChirp(dbChirp)
	data, err := json.Marshal(chirp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (cfg ApiConfig) CreateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Add("Content-type", "application/json")

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Couldn't find JWT: %s", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Couldn't validate JWT: %s", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	newChirp := model.Chirp{}
	if err := decoder.Decode(&newChirp); err != nil {
		response := model.JsonErrorResponse{Error: "Something went wrong"}
		data, err := json.Marshal(response)
		w.WriteHeader(http.StatusInternalServerError)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			return
		}
		w.Write(data)
		return
	}

	if len(newChirp.Body) > 140 {
		response := model.JsonErrorResponse{Error: "Chirp is too long"}
		data, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return
	}

	newChirp.Body = util.CleanBody(newChirp.Body)

	chirpParams := database.CreateChirpParams{
		Body:   newChirp.Body,
		UserID: userId,
	}
	dbChirp, err := cfg.Db.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		w.WriteHeader(500)
		return
	}

	newChirp.MapDBChirp(dbChirp)
	data, err := json.Marshal(newChirp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
