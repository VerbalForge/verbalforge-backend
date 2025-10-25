package models

import "time"

// SupportTicket represents a support ticket from a user
type SupportTicket struct {
	ID        string    `json:"id" bson:"_id"`
	UserID    string    `json:"userId" bson:"user_id"`
	Name      string    `json:"name" bson:"name"`
	Email     string    `json:"email" bson:"email"`
	Subject   string    `json:"subject" bson:"subject"`
	Message   string    `json:"message" bson:"message"`
	Status    string    `json:"status" bson:"status"`     // open, in-progress, resolved, closed
	Priority  string    `json:"priority" bson:"priority"` // low, medium, high, urgent
	CreatedAt time.Time `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updated_at"`
}

// CreateSupportTicketRequest represents the request to create a new support ticket
type CreateSupportTicketRequest struct {
	Name    string `json:"name" binding:"required,min=2"`
	Email   string `json:"email" binding:"required,email"`
	Subject string `json:"subject" binding:"required,min=5,max=200"`
	Message string `json:"message" binding:"required,min=10,max=5000"`
}

// UpdateSupportTicketRequest represents the request to update a support ticket
type UpdateSupportTicketRequest struct {
	Status   string `json:"status" binding:"omitempty,oneof=open in-progress resolved closed"`
	Priority string `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
}
