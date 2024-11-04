package handler

import (
	"encoding/json"
	"log"
	"net/http"

	model "github.com/JosueAD95/Server-course/models"
	util "github.com/JosueAD95/Server-course/utils"
)

func ValidateChirp(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	params := model.Parameters{}
	response := model.JsonResponse{}
	w.Header().Add("Content-type", "application/json")
	if err := decoder.Decode(&params); err != nil {
		response.Error = "Something went wrong"
		data, err := json.Marshal(response)
		w.WriteHeader(500)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			return
		}
		w.Write(data)
		return
	}

	if len(params.Body) > 140 {
		response.Error = "Chirp is too long"
		data, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(400)
		w.Write(data)
		return
	}

	response.CleanBody = util.CleanBody(params.Body)
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
