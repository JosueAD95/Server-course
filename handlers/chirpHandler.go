package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"

	"github.com/google/uuid"

	"github.com/JosueAD95/Server-course/internal/auth"
	database "github.com/JosueAD95/Server-course/internal/database"
	model "github.com/JosueAD95/Server-course/models"
	util "github.com/JosueAD95/Server-course/utils"
)

func (cfg ApiConfig) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Query().Get("author_id")
	sortType := r.URL.Query().Get("sort")
	var dbChirps []database.Chirp
	var err error
	if authorId != "" {
		id, _ := uuid.Parse(authorId)
		dbChirps, err = cfg.Db.GetChirpsByUserId(r.Context(), id)
	} else {
		dbChirps, err = cfg.Db.GetChirps(r.Context())
	}

	if sortType == "desc" {
		sort.Slice(dbChirps, func(i, j int) bool { return dbChirps[i].CreatedAt.After(dbChirps[j].CreatedAt) })
	} else {
		sort.Slice(dbChirps, func(i, j int) bool { return dbChirps[i].CreatedAt.Before(dbChirps[j].CreatedAt) })
	}

	if err != nil {
		log.Printf("Error retriaving all chirps: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
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
	w.Header().Add("Content-type", "application/json")
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
		w.WriteHeader(http.StatusNotFound)
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

func (cfg ApiConfig) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error parsing chirpId parameter: %s", err)
		return
	}

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

	dbChirp, err := cfg.Db.GetChirpById(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error retriaving chirp '%s': %s", chirpID.String(), err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if dbChirp.UserID != userId {
		log.Printf("User Id from Chirp and user id from token are not the same")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = cfg.Db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		log.Printf("Couldn't delete chirp '%s': %s", chirpID.String(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
