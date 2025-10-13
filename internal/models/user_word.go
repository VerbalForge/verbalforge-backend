package models

import (
	"time"
)

type UserWord struct {
	ID        string    `json:"id" bson:"_id"`
	UserID    string    `json:"user_id" bson:"user_id"`
	Known     []string  `json:"known" bson:"known"`
	Practice  []string  `json:"practice" bson:"practice"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
