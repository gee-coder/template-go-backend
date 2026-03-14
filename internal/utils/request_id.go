package utils

import "github.com/google/uuid"

// NewRequestID returns a new request ID.
func NewRequestID() string {
	return uuid.NewString()
}

