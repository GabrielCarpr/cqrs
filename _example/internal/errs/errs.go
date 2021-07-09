package errs

import (
    "github.com/GabrielCarpr/cqrs/errors"
    stdErrors "errors"
)

var (
    UniqueEntityExists = stdErrors.New("UniqueEntityExists")
    EntityNotFound = stdErrors.New("EntityNotFound")
)

// ValidationError creates an error indicating validation of a
// message occurred
func ValidationError(message string) errors.Error {
    return errors.Error{Code: 400, Message: message}
}
