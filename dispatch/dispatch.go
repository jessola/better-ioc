// Package dispatch provides mediator-style request dispatch backed by an IOC
// container.
package dispatch

import (
	"fmt"
	"reflect"

	"github.com/jessola/better-ioc"
)

// ErrCouldNotCast is returned when a function handler's result cannot be cast
// to its registered response type.
type ErrCouldNotCast struct {
}

// ErrNoHandler is returned when no handler is registered for a request type.
type ErrNoHandler struct {
	t reflect.Type
}

// Dispatcher dispatches requests to handlers registered in its container.
type Dispatcher struct {
	container *ioc.Container
}

// Handler handles a request and returns a response or an error.
type Handler[Req, Res any] interface {
	Handle(req Req) (Res, error)
}

// ErrHandler handles a request that has no response value.
type ErrHandler[Req any] interface {
	Handle(req Req) error
}

// HandlerFunc adapts a function to a Handler.
type HandlerFunc[Req, Res any] func(req Req) (Res, error)

// ErrHandlerFunc adapts a function to an ErrHandler.
type ErrHandlerFunc[Req any] func(req Req) error

func newDispatcher(c *ioc.Container) *Dispatcher {
	return &Dispatcher{container: c}
}

// Error returns the description of err.
func (err ErrCouldNotCast) Error() string {
	return "could not cast..." // TODO fix this
}

// Error returns the description of err.
func (err ErrNoHandler) Error() string {
	return fmt.Sprintf("no handler for type %s", err.t.String())
}

// AddDispatcher registers a scoped Dispatcher in c.
func AddDispatcher(c *ioc.Container) error {
	// Maybe err if you try to add another one?
	return c.AddScoped[*Dispatcher](newDispatcher)
}

// Handle invokes fn with req.
func (fn HandlerFunc[Req, Res]) Handle(req Req) (Res, error) {
	return fn(req)
}

// Handle invokes fn with req.
func (fn ErrHandlerFunc[Req]) Handle(req Req) error {
	return fn(req)
}

// TODO: Consider returning errors on duplicate as you should not really have
// multiple handlers for a given request type. This is not some generic message
// bus

// RegisterHandler registers handler to handle requests of type Req and return
// responses of type Res.
func RegisterHandler[Req, Res any](c *ioc.Container, handler Handler[Req, Res], overrides ...ioc.Override) error {
	return c.AddTransient[Handler[Req, Res]](ioc.Value(handler), overrides...)
}

// RegisterErrHandler registers handler to handle requests of type Req without
// returning a response value.
func RegisterErrHandler[Req any](c *ioc.Container, handler ErrHandler[Req], overrides ...ioc.Override) error {
	return c.AddTransient[ErrHandler[Req]](ioc.Value(handler), overrides...)
}

// RegisterHandlerFunc registers fn as a handler for Req with a response of Res.
// When dispatched, fn's parameters are resolved from the active container scope.
func RegisterHandlerFunc[Req, Res any](c *ioc.Container, fn any, overrides ...ioc.Override) error {
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		return ioc.NewErrNotAFunc(reflect.TypeOf(fn))
	}

	return c.AddTransient[Handler[Req, Res]](func(s *ioc.Container) HandlerFunc[Req, Res] {
		return func(req Req) (res Res, err error) {
			val, err := ioc.MustCallValErr(s, fn, overrides...)
			if err != nil {
				return
			}
			var ok bool
			if res, ok = val.(Res); !ok {
				err = ErrCouldNotCast{}
			}
			return
		}
	})
}

// RegisterErrHandlerFunc registers fn as a handler for Req without a response
// value. When dispatched, fn's parameters are resolved from the active
// container scope.
func RegisterErrHandlerFunc[Req any](c *ioc.Container, fn any, overrides ...ioc.Override) error {
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		return ioc.NewErrNotAFunc(reflect.TypeOf(fn))
	}

	return c.AddTransient[ErrHandler[Req]](func(s *ioc.Container) ErrHandlerFunc[Req] {
		return func(req Req) error {
			return ioc.MustCallErr(s, fn, overrides...)
		}
	})
}

// Send dispatches req to its Handler and returns the response or an error.
func (d *Dispatcher) Send[Res, Req any](req Req) (res Res, err error) {
	d.container.AddScoped[Req](ioc.Value(req))
	handler, ok := d.container.Get[Handler[Req, Res]]()
	if !ok {
		err = ErrNoHandler{t: reflect.TypeFor[Handler[Req, Res]]()}
		return
	}
	res, err = handler.Handle(req)
	return
}

// SendErr dispatches req to its ErrHandler and returns any handler error.
func (d *Dispatcher) SendErr[Req any](req Req) (err error) {
	d.container.AddScoped[Req](ioc.Value(req))
	handler, ok := d.container.Get[ErrHandler[Req]]()
	if !ok {
		err = ErrNoHandler{t: reflect.TypeFor[ErrHandler[Req]]()}
		return
	}
	err = handler.Handle(req)
	return
}
