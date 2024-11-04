package model

import (
	"time"

	db "github.com/JosueAD95/Server-course/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (u *User) MapDBUser(dbUser db.User) {
	u.ID = dbUser.ID
	u.Email = dbUser.Email
	u.CreatedAt = dbUser.CreatedAt
	u.UpdatedAt = dbUser.UpdatedAt
}
