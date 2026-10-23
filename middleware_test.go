package ioc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithMiddleware(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()
	order := []string{}

	assert.NoError(c.AddScoped[int](Value(42)))
	handler := WithOverrides(func(value any) string {
		assert.Equal(42, value)
		order = append(order, "handler")
		return "handler"
	}, Provide[int](0))

	outer := MiddlewareFunc(func(next any) any {
		return func() string {
			order = append(order, "outer before")
			value := MustCall(c, next)[0].Interface().(string)
			order = append(order, "outer after")
			return "outer(" + value + ")"
		}
	})
	inner := MiddlewareFunc(func(next any) any {
		return func() string {
			order = append(order, "inner before")
			value := MustCall(c, next)[0].Interface().(string)
			order = append(order, "inner after")
			return "inner(" + value + ")"
		}
	})

	effective := Must(WithMiddleware(handler, outer, inner))
	result := MustCall(c, effective)[0].Interface().(string)

	assert.Equal("outer(inner(handler))", result)
	assert.Equal([]string{
		"outer before",
		"inner before",
		"handler",
		"inner after",
		"outer after",
	}, order)
}
