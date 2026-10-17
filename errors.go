package ioc

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrNameCannotBeBlank is returned when a named provider is registered with an
// empty or whitespace-only name.
var ErrNameCannotBeBlank = errors.New("name cannot be blank")

type ErrCouldNotCast struct {
}

// ErrNotAFunc is returned when a provider is not a function.
type ErrNotAFunc struct {
	t reflect.Type
}

// ErrNotFound is returned when no provider has been registered for a type.
type ErrNotFound struct {
	t reflect.Type
}

// Error returns the description of err.
func (err ErrNotAFunc) Error() string {
	return fmt.Sprintf("%s is not a func", err.t.String())
}

// Error returns the description of err.
func (err ErrNotFound) Error() string {
	return fmt.Sprintf("no service with type %s", err.t.String())
}

// Error returns the description of err.
func (err ErrCouldNotCast) Error() string {
	return "could not cast..." // TODO fix this
}

func NewErrNotAFunc(t reflect.Type) ErrNotAFunc {
	return ErrNotAFunc{t: t}
}
