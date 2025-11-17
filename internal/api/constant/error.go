package constant

import "net/http"

type CustomError struct {
	StatusCode int
	Message    string
}

func NewCError(StatusCode int, Message string) CustomError {
	return CustomError{StatusCode: StatusCode, Message: Message}
}

func (err CustomError) Error() string {
	return err.Message
}

var (
	ErrNoSymbol = NewCError(http.StatusBadRequest,
		"please provide symbol")
)
