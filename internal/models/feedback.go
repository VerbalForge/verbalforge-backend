package models

import "time"

// Feedback represents user feedback submission
type Feedback struct {
	ID        string    `json:"id" bson:"_id"`
	Email     string    `json:"email,omitempty" bson:"email,omitempty"`
	Type      string    `json:"type" bson:"type"` // suggestion, bug, general, praise
	Message   string    `json:"message" bson:"message"`
	Status    string    `json:"status" bson:"status"` // new, reviewed, archived
	CreatedAt time.Time `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updated_at"`
}

// CreateFeedbackRequest represents the request to create new feedback
type CreateFeedbackRequest struct {
	Email   string `json:"email" binding:"omitempty,email"`
	Type    string `json:"type" binding:"required,oneof=suggestion bug general praise"`
	Message string `json:"message" binding:"required,min=10,max=2000"`
}

// UpdateFeedbackStatusRequest represents the request to update feedback status
type UpdateFeedbackStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=new reviewed archived"`
}
