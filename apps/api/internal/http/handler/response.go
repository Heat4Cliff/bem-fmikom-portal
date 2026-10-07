package handler

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Envelope is the common API response shape. Endpoint-specific business
// payloads are intentionally not introduced in Phase 1.
type Envelope struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Code    string              `json:"code,omitempty"`
	Data    any                 `json:"data,omitempty"`
	Errors  map[string][]string `json:"errors,omitempty"`
	Meta    *Meta               `json:"meta,omitempty"`
}

func Success(message string, data any) Envelope {
	return Envelope{Success: true, Message: message, Data: data}
}

func Failure(message string, errors map[string][]string) Envelope {
	return Envelope{Success: false, Message: message, Errors: errors}
}

func FailureCode(code, message string, errors map[string][]string) Envelope {
	return Envelope{Success: false, Message: message, Code: code, Errors: errors}
}
