package apperror

type Code string

const (
	NotFound  Code = "NOT_FOUND"
	Duplicate Code = "DUPLICATE"
	Internal  Code = "INTERNAL"
	Invalid   Code = "INVALID"
)

type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}
