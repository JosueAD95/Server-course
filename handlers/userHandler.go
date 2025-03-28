package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	auth "github.com/JosueAD95/Server-course/internal/auth"
	db "github.com/JosueAD95/Server-course/internal/database"
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
	password, err := auth.HashPassword(newUser.Password)
	if err != nil {
		log.Printf("Error hashing the password: %s", err)
		w.WriteHeader(500)
		return
	}
	userParams := db.CreateUserParams{
		Email:          newUser.Email,
		HashedPassword: password,
	}
	rowUser, err := cfg.Db.CreateUser(r.Context(), userParams)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.WriteHeader(500)
		return
	}

	newUser.MapRowUser(rowUser)
	data, err := json.Marshal(newUser)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

func (cfg ApiConfig) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		model.User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	w.Header().Add("Content-type", "application/json")
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	u, err := cfg.Db.GetUserByEmail(r.Context(), params.Email)

	if err != nil {
		log.Printf("Error searching for user (%s): %s", params.Email, err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	err = auth.CheckPasswordHash(params.Password, u.HashedPassword)
	if err != nil {
		log.Printf("Error comparing password of user (%s): %s", params.Email, err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	refreshTokenParams := db.SaveRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		UserID:    u.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	}

	if _, err := cfg.Db.SaveRefreshToken(r.Context(), refreshTokenParams); err != nil {
		log.Printf("Error saving the refresh token: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	accessToken, err := auth.MakeJWT(
		u.ID,
		cfg.JWTSecret,
		time.Hour,
	)

	if err != nil {
		log.Printf("Couldn't create access JWT: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user := model.User{}
	user.MapDbUser(u)

	resp := response{
		User:         user,
		Token:        accessToken,
		RefreshToken: refreshTokenParams.Token,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (cfg ApiConfig) RefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Refesh token not found: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userRow, err := cfg.Db.GetUserIdFromRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("Error searching refresh token '%s': %s", token, err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if userRow.RevokedAt.Valid {
		log.Printf("The token was revoked at: %s", userRow.RevokedAt.Time)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	accessToken, err := auth.MakeJWT(
		userRow.UserID,
		cfg.JWTSecret,
		time.Hour,
	)

	if err != nil {
		log.Printf("Couldn't create access JWT: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type response struct {
		Token string `json:"token"`
	}
	resp := response{Token: accessToken}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (cfg ApiConfig) RevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Refesh token not found: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	result, err := cfg.Db.RevokeToken(r.Context(), token)
	rows, errorQuery := result.RowsAffected()
	if err != nil || errorQuery != nil || rows == 0 {
		log.Printf("Error revoking refresh token '%s': %s", token, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
