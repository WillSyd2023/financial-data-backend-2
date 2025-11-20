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

	ErrInvalidLimit = NewCError(http.StatusBadRequest,
		"invalid 'limit' query parameter: must be a positive integer")

	ErrInvalidCursor = NewCError(http.StatusBadRequest,
		"invalid 'before' query parameter: must be a non-negative integer (Unix millisecond timestamp)")
)
