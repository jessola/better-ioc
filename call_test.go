package ioc

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type makeScopedFunc func(string) string

func TestMustCall(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	assert.NoError(c.AddScoped[string](Value("resolved")))

	out := MustCall(c, func(value string) (string, int) {
		return value, len(value)
	})

	assert.Len(out, 2)
	assert.Equal("resolved", out[0].Interface())
	assert.Equal(8, out[1].Interface())
}

func TestMustCallErr(t *testing.T) {
	t.Run("returns nil", func(t *testing.T) {
		assert := assert.New(t)
		c := NewContainer()

		assert.NoError(c.AddScoped[string](Value("resolved")))
		err := MustCallErr(c, func(value string) error {
			assert.Equal("resolved", value)
			return nil
		})

		assert.NoError(err)
	})

	t.Run("returns error", func(t *testing.T) {
		assert := assert.New(t)
		c := NewContainer()
		expected := errors.New("call failed")

		err := MustCallErr(c, func() error {
			return expected
		})

		assert.ErrorIs(err, expected)
	})
}

func TestMustCallValErr(t *testing.T) {
	t.Run("returns value", func(t *testing.T) {
		assert := assert.New(t)
		c := NewContainer()

		assert.NoError(c.AddScoped[int](Value(42)))
		value, err := MustCallValErr(c, func(number int) (string, error) {
			return "answer", nil
		})

		assert.NoError(err)
		assert.Equal("answer", value)
	})

	t.Run("returns error", func(t *testing.T) {
		assert := assert.New(t)
		c := NewContainer()
		expected := errors.New("call failed")

		value, err := MustCallValErr(c, func() (string, error) {
			return "", expected
		})

		assert.Nil(value)
		assert.ErrorIs(err, expected)
	})
}

func TestMakeScopedFunc(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()
	scopes := []*Container{}

	assert.NoError(c.AddScoped[int](Value(42)))
	fn, err := MakeScopedFunc[makeScopedFunc](c, func(request string, value int, scope *Container) string {
		assert.Equal(42, value)
		scopes = append(scopes, scope)
		return request
	})
	assert.NoError(err)

	assert.Equal("first", fn("first"))
	assert.Equal("second", fn("second"))
	assert.Len(scopes, 2)
	assert.NotSame(c, scopes[0])
	assert.NotSame(scopes[0], scopes[1])
}

func TestMustMakeScopedFunc(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	fn := MustMakeScopedFunc[makeScopedFunc](c, func(request string) string {
		return request
	})

	assert.Equal("request", fn("request"))
	assert.PanicsWithValue(NewErrNotAFunc(reflect.TypeOf("not a function")), func() {
		MustMakeScopedFunc[makeScopedFunc](c, "not a function")
	})
}
