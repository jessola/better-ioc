package ioc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvide(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	assert.NoError(c.AddScoped[int](Value(42)))
	assert.NoError(c.AddScoped[bool](WithOverrides(func(value any) bool {
		return value.(int) == 42
	}, Provide[int](0))))

	assert.True(c.MustGet[bool]())
}

func TestProvideNamed(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	assert.NoError(c.AddNamedScoped[int]("answer", Value(42)))
	assert.NoError(c.AddScoped[bool](WithOverrides(func(value any) bool {
		return value.(int) == 42
	}, ProvideNamed[int](0, "answer"))))

	assert.True(c.MustGet[bool]())
}

func TestProvideValue(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	assert.NoError(c.AddScoped[bool](WithOverrides(func(value any) bool {
		return value.(int) == 42
	}, ProvideValue[int](0, 42))))

	assert.True(c.MustGet[bool]())
}

func TestProvideFactory(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	assert.NoError(c.AddScoped[string](Value("factory value")))
	assert.NoError(c.AddScoped[bool](WithOverrides(func(value any) bool {
		return value.(string) == "factory value"
	}, ProvideFactory(0, func(c *Container) any {
		return c.MustGet[string]()
	}))))

	assert.True(c.MustGet[bool]())
}

func TestWithOverridesMustCall(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	assert.NoError(c.AddScoped[int](Value(42)))
	out := MustCall(c, WithOverrides(func(value any) bool {
		return value.(int) == 42
	}, Provide[int](0)))

	assert.True(out[0].Bool())
}

func TestProvideInstructionsWithInferredDependencies(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	assert.NoError(c.AddScoped[string](Value("regular")))
	assert.NoError(c.AddScoped[int](Value(42)))
	assert.NoError(c.AddScoped[bool](Value(true)))
	assert.NoError(c.AddNamedScoped[float64]("ratio", Value(1.5)))
	assert.NoError(c.AddScoped[[]string](Value([]string{"tag"})))
	assert.NoError(c.AddScoped[struct{}](WithOverrides(func(
		label string,
		count any,
		enabled bool,
		ratio any,
		tags []string,
		message any,
	) struct{} {
		assert.Equal("regular", label)
		assert.Equal(42, count)
		assert.True(enabled)
		assert.Equal(1.5, ratio)
		assert.Equal([]string{"tag"}, tags)
		assert.Equal("inline", message)
		return struct{}{}
	},
		Provide[int](1),
		ProvideNamed[float64](3, "ratio"),
		ProvideValue(5, "inline"),
	)))

	c.MustGet[struct{}]()
}
