package dto

type ErrorType struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}
