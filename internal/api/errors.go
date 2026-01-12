package api

import "encoding/json"

// ErrorCode represents an error code
type ErrorCode string

const (
	ErrorCodePolicyDenied = "POLICY_DENIED"
	ErrorCodeValidation   = "VALIDATION"
	ErrorCodeNotFound     = "NOT_FOUND"
	ErrorCodeConflict     = "CONFLICT"
	ErrorCodeInternal     = "INTERNAL"
)

// Error represents an API error
type Error struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error Error `json:"error"`
}

// NewError creates a new error
func NewError(code ErrorCode, message string, details map[string]interface{}) *ErrorResponse {
	return &ErrorResponse{
		Error: Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// JSON returns the error as JSON bytes
func (e *ErrorResponse) JSON() ([]byte, error) {
	return json.Marshal(e)
}
