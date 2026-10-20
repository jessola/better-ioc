package dispatch

import (
	"errors"
	"reflect"
	"testing"

	ioc "github.com/jessola/better-ioc"
	"github.com/stretchr/testify/assert"
)

type dispatchTestRequest struct {
	Value string
}

type dispatchTestResponse struct {
	Value string
}

type dispatchTestHandler struct {
	err error
}

func (h dispatchTestHandler) Handle(req dispatchTestRequest) (dispatchTestResponse, error) {
	if h.err != nil {
		return dispatchTestResponse{}, h.err
	}
	return dispatchTestResponse{Value: req.Value}, nil
}

type dispatchTestCommand struct{}

type dispatchTestCommandHandler struct {
	handled *bool
	err     error
}

func (h dispatchTestCommandHandler) Handle(dispatchTestCommand) error {
	*h.handled = true
	return h.err
}

func TestSend(t *testing.T) {
	assert := assert.New(t)
	c := ioc.NewContainer()

	assert.NoError(AddDispatcher(c))
	assert.NoError(AddHandler(c, dispatchTestHandler{}))

	response, err := c.MustGet[*Dispatcher]().Send[dispatchTestResponse](dispatchTestRequest{Value: "sent"})

	assert.NoError(err)
	assert.Equal(dispatchTestResponse{Value: "sent"}, response)
}

func TestSendReturnsHandlerError(t *testing.T) {
	assert := assert.New(t)
	c := ioc.NewContainer()
	expected := errors.New("handler failed")

	assert.NoError(AddDispatcher(c))
	assert.NoError(AddHandler(c, dispatchTestHandler{err: expected}))

	response, err := c.MustGet[*Dispatcher]().Send[dispatchTestResponse](dispatchTestRequest{})

	assert.ErrorIs(err, expected)
	assert.Equal(dispatchTestResponse{}, response)
}

func TestAddHandlerFunc(t *testing.T) {
	assert := assert.New(t)
	c := ioc.NewContainer()

	assert.NoError(AddDispatcher(c))
	assert.NoError(c.AddScoped[string](ioc.Value("prefix")))
	assert.NoError(AddHandlerFunc[dispatchTestRequest, dispatchTestResponse](c, func(
		prefix string,
		request dispatchTestRequest,
	) (dispatchTestResponse, error) {
		return dispatchTestResponse{Value: prefix + ":" + request.Value}, nil
	}))

	response, err := c.MustGet[*Dispatcher]().Send[dispatchTestResponse](dispatchTestRequest{Value: "request"})

	assert.NoError(err)
	assert.Equal(dispatchTestResponse{Value: "prefix:request"}, response)
}

func TestAddHandlerFuncReturnsErrCouldNotCast(t *testing.T) {
	assert := assert.New(t)
	c := ioc.NewContainer()

	assert.NoError(AddDispatcher(c))
	assert.NoError(AddHandlerFunc[dispatchTestRequest, dispatchTestResponse](c, func() (string, error) {
		return "wrong response type", nil
	}))

	_, err := c.MustGet[*Dispatcher]().Send[dispatchTestResponse](dispatchTestRequest{})

	assert.ErrorAs(err, new(ErrCouldNotCast))
}

func TestSendErr(t *testing.T) {
	assert := assert.New(t)
	c := ioc.NewContainer()
	handled := false

	assert.NoError(AddDispatcher(c))
	assert.NoError(AddErrHandler(c, dispatchTestCommandHandler{handled: &handled}))

	err := c.MustGet[*Dispatcher]().SendErr(dispatchTestCommand{})

	assert.NoError(err)
	assert.True(handled)
}

func TestSendErrReturnsHandlerError(t *testing.T) {
	assert := assert.New(t)
	c := ioc.NewContainer()
	handled := false
	expected := errors.New("handler failed")

	assert.NoError(AddDispatcher(c))
	assert.NoError(AddErrHandler(c, dispatchTestCommandHandler{
		handled: &handled,
		err:     expected,
	}))

	err := c.MustGet[*Dispatcher]().SendErr(dispatchTestCommand{})

	assert.ErrorIs(err, expected)
	assert.True(handled)
}

func TestAddErrHandlerFunc(t *testing.T) {
	assert := assert.New(t)
	c := ioc.NewContainer()
	handled := false

	assert.NoError(AddDispatcher(c))
	assert.NoError(c.AddScoped[string](ioc.Value("dependency")))
	assert.NoError(AddErrHandlerFunc[dispatchTestCommand](c, func(
		dependency string,
		command dispatchTestCommand,
	) error {
		assert.Equal("dependency", dependency)
		assert.Equal(dispatchTestCommand{}, command)
		handled = true
		return nil
	}))

	err := c.MustGet[*Dispatcher]().SendErr(dispatchTestCommand{})

	assert.NoError(err)
	assert.True(handled)
}

func TestAddHandlerFuncsRequireFunctions(t *testing.T) {
	assert := assert.New(t)
	c := ioc.NewContainer()

	err := AddHandlerFunc[dispatchTestRequest, dispatchTestResponse](c, "not a function")
	var notAFunc ioc.ErrNotAFunc
	assert.ErrorAs(err, &notAFunc)

	err = AddErrHandlerFunc[dispatchTestCommand](c, "not a function")
	assert.ErrorAs(err, &notAFunc)
}

func TestSendReturnsErrNoHandler(t *testing.T) {
	assert := assert.New(t)
	c := ioc.NewContainer()

	assert.NoError(AddDispatcher(c))

	_, err := c.MustGet[*Dispatcher]().Send[dispatchTestResponse](dispatchTestRequest{})

	var noHandler ErrNoHandler
	assert.ErrorAs(err, &noHandler)
	assert.Equal(reflect.TypeFor[Handler[dispatchTestRequest, dispatchTestResponse]](), noHandler.t)
}

func TestDispatcherUsesActiveScope(t *testing.T) {
	assert := assert.New(t)
	root := ioc.NewContainer()
	child := root.CreateScope()

	assert.NoError(AddDispatcher(root))
	assert.NoError(AddHandler(child, dispatchTestHandler{}))

	response, err := child.MustGet[*Dispatcher]().Send[dispatchTestResponse](dispatchTestRequest{Value: "child"})
	assert.NoError(err)
	assert.Equal(dispatchTestResponse{Value: "child"}, response)

	_, err = root.MustGet[*Dispatcher]().Send[dispatchTestResponse](dispatchTestRequest{})
	assert.ErrorAs(err, new(ErrNoHandler))
}
