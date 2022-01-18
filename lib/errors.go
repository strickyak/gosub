package errors

type Error struct {
	message string
}

func New(message string) error {
	return &Error{message: message}
}

func (e *Error) Error() string {
	return "ERROR: " + e.message
}
func (e *Error) String() string {
	return e.Error()
}
