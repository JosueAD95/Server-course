package handler

import (
	"encoding/json"
	"log"
	"net/http"

	model "github.com/JosueAD95/Server-course/models"
)

func (cfg ApiConfig) AddUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	newUser := model.User{}
	w.Header().Add("Content-type", "application/json")
	if err := decoder.Decode(&newUser); err != nil {
		log.Printf("Error decoding JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	if newUser.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	dbUser, err := cfg.Db.CreateUser(r.Context(), newUser.Email)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.WriteHeader(500)
		return
	}
	newUser.MapDBUser(dbUser)
	data, err := json.Marshal(newUser)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}
