package errors

type Error struct {
	message string
}

func New(message string) error {
	return &Error{message}
}

func (e *Error) Error() string {
	return "ERROR: " + message
}
