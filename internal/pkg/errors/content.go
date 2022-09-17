package errors

func ContentError(err error, content string) error {
	return &contentError{err, content}
}

type contentError struct {
	error
	Content string
}
