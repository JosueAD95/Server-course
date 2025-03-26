package model

import (
	"time"

	db "github.com/JosueAD95/Server-course/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (c *Chirp) MapDBChirp(dbChirp db.Chirp) {
	c.Id = dbChirp.ID
	c.Body = dbChirp.Body
	c.UserId = dbChirp.UserID
	c.CreatedAt = dbChirp.CreatedAt
	c.UpdatedAt = dbChirp.UpdatedAt
}
