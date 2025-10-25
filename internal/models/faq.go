package models

import "time"

// FAQ represents a frequently asked question submission
type FAQ struct {
	ID        string    `json:"id" bson:"_id"`
	Email     string    `json:"email" bson:"email"`
	Question  string    `json:"question" bson:"question"`
	Status    string    `json:"status" bson:"status"` // pending, answered, archived
	Answer    string    `json:"answer,omitempty" bson:"answer,omitempty"`
	CreatedAt time.Time `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updated_at"`
}

// CreateFAQRequest represents the request to create a new FAQ
type CreateFAQRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Question string `json:"question" binding:"required,min=10,max=1000"`
}

// UpdateFAQStatusRequest represents the request to update FAQ status
type UpdateFAQStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending answered archived"`
	Answer string `json:"answer"`
}
