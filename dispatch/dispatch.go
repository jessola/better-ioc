package dispatch

import (
	"fmt"
	"reflect"

	ioc "github.com/jessola/better-ioc"
)

type ErrCouldNotCast struct {
}

type ErrNoHandler struct {
	t reflect.Type
}

type Dispatcher struct {
	container *ioc.Container
}

type Handler[Req, Res any] interface {
	Handle(req Req) (Res, error)
}

type ErrHandler[Req any] interface {
	Handle(req Req) error
}

type HandlerFunc[Req, Res any] func(req Req) (Res, error)

type ErrHandlerFunc[Req any] func(req Req) error

func newDispatcher(c *ioc.Container) *Dispatcher {
	return &Dispatcher{container: c}
}

func (err ErrCouldNotCast) Error() string {
	return "could not cast..." // TODO fix this
}

func (err ErrNoHandler) Error() string {
	return fmt.Sprintf("no handler for type %s", err.t.String())
}

func AddDispatcher(c *ioc.Container) error {
	// Maybe err if you try to add another one?
	return c.AddScoped[*Dispatcher](newDispatcher)
}

func (fn HandlerFunc[Req, Res]) Handle(req Req) (Res, error) {
	return fn(req)
}

func (fn ErrHandlerFunc[Req]) Handle(req Req) error {
	return fn(req)
}

// TODO: Consider returning errors on duplicate as you should not really have
// multiple handlers for a given request type. This is not some generic message
// bus

func RegisterHandler[Req, Res any](c *ioc.Container, handler Handler[Req, Res], overrides ...ioc.Override) error {
	return c.AddTransient[Handler[Req, Res]](ioc.Value(handler), overrides...)
}

func RegisterErrHandler[Req any](c *ioc.Container, handler ErrHandler[Req], overrides ...ioc.Override) error {
	return c.AddTransient[ErrHandler[Req]](ioc.Value(handler), overrides...)
}

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
