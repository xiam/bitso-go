package bitso

import (
	"fmt"
)

// Envelope represents a common response envelope from Bitso API.
type Envelope struct {
	Success bool `json:"success"`
	Error   struct {
		Code    interface{} `json:"code"`
		Message string      `json:"message"`
	} `json:"error,omitempty"`
}

// Error represents an API error
type Error struct {
	code    int
	message string
}

// Error returns the error message.
func (e Error) Error() string {
	return fmt.Sprintf("Error %v: %s", e.code, e.message)
}

// Code returns the error code.
func (e Error) Code() int {
	return e.code
}

func apiError(code int, message string) error {
	return &Error{code: code, message: message}
}
