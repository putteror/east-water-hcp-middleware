package common

import "errors"

// Standard domain and 3rd-party integration errors
var (
	ErrBadRequest         = errors.New("invalid request parameters")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrNotFound           = errors.New("resource not found")
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrThirdPartyTimeout  = errors.New("third party request timed out")
	ErrThirdPartyBadResp  = errors.New("third party returned an invalid response")
	ErrThirdPartyService  = errors.New("third party service unavailable or returned error")
	ErrInternalServer     = errors.New("internal server error")
)

// ThirdPartyError represents a structured error returned by an external API
type ThirdPartyError struct {
	StatusCode int    `json:"status_code"`
	ErrorCode  string `json:"error_code,omitempty"`
	Message    string `json:"message"`
	RawBody    string `json:"raw_body,omitempty"`
}

func (e *ThirdPartyError) Error() string {
	if e.ErrorCode != "" {
		return "[" + e.ErrorCode + "] " + e.Message
	}
	return e.Message
}
