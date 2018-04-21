package mw

// Error interface describes errors that can contain HTTP status code.
type Error interface {
	error
	Code() int
}

// StatusError is a basic Error interface implementation.
type StatusError struct {
	error
	code int
}

// NewStatusError is a constructor func for StatusError.
func NewStatusError(code int, cause error) StatusError {
	return StatusError{error: cause, code: code}
}

// Code return HTTP status code to satisfy Error interface.
func (e StatusError) Code() int {
	return e.code
}

// Error returns the actual error message.
func (e StatusError) Error() string {
	return e.error.Error()
}
