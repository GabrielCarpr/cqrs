// Package errors includes standard error helpers
package errors

var (
	// InternalServerError is an error that has been hidden from the port-interface
	InternalServerError = Error{500, "Internal server error"}
)

// Error is a port-interface visible error
// If an error is provided to a port and it's not an Error,
// it should be hidden
//
// Ports will interpret Error however they choose. Eg, CLI may just show the
// message, and HTTP may show the message and the error code. Or, HTTP
// may maintain a mapping from codes to HTTP statuses
type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return e.Message
}

// Block hides non-Error errors
func Block(e error) Error {
	err, ok := e.(Error)
	if !ok {
		return InternalServerError
	}
	return err
}
