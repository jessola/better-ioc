package ioc

import (
	"errors"
	"fmt"
	"reflect"
)

var ErrNameCannotBeBlank = errors.New("name cannot be blank")

type ErrNotAFunc struct {
	t reflect.Type
}

func (err ErrNotAFunc) Error() string {
	return fmt.Sprintf("%s is not a func", err.t.String())
}
