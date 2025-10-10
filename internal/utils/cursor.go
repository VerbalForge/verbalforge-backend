package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// CursorData represents the data stored in a continuation token
type CursorData struct {
	LastID        string `json:"last_id"`
	LastCreatedAt string `json:"last_created_at"`
}

// EncodeCursor creates a base64-encoded continuation token
func EncodeCursor(data CursorData) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor data: %w", err)
	}
	return base64.URLEncoding.EncodeToString(jsonData), nil
}

// DecodeCursor decodes a base64-encoded continuation token
func DecodeCursor(cursor string) (*CursorData, error) {
	if cursor == "" {
		return nil, nil
	}

	jsonData, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cursor: %w", err)
	}

	var data CursorData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cursor data: %w", err)
	}

	return &data, nil
}
