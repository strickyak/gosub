package errors

type Error struct {
	message string
}

func New(message string) error {
	return &Error{message}
}
