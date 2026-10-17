package ioc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecursion(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddSingleton[int](Value(4))

	s := c.CreateScope().CreateScope().CreateScope().CreateScope()

	x, ok := s.Get[int]()
	assert.True(ok)
	assert.Equal(4, x)
}

func TestScopeOverrides(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddScoped[int](Value(0))
	c.AddScoped[int](Value(42))

	s1 := c.CreateScope()
	s1.AddScoped[int](Value(100))

	s2 := c.CreateScope()
	s2.AddScoped[int](Value(-5))
	s2.AddScoped[int](Value(9000))

	x, ok := c.Get[int]()
	assert.True(ok)

	y, ok := s1.Get[int]()
	assert.True(ok)

	z, ok := s2.Get[int]()
	assert.True(ok)

	assert.Equal(42, x)
	assert.Equal(100, y)
	assert.Equal(9000, z)
}

func TestNamedDependenciesCanShareNameAcrossTypes(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	assert.NoError(c.AddNamedScoped[string]("settings", Value("production")))
	assert.NoError(c.AddNamedScoped[int]("settings", Value(3)))

	s := c.CreateScope()

	name, ok := s.GetNamed[string]("settings")
	assert.True(ok)
	assert.Equal("production", name)

	retries, ok := s.GetNamed[int]("settings")
	assert.True(ok)
	assert.Equal(3, retries)
}

func TestGetSlice(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddScoped[int](Value(1))
	c.AddScoped[int](Value(2))
	c.AddScoped[int](Value(3))

	sIntermediary := c.CreateScope()
	sIntermediary.AddScoped[int](Value(4))
	sIntermediary.AddScoped[int](Value(5))
	s1 := sIntermediary.CreateScope()
	s1.AddScoped[int](6)

	s2 := c.CreateScope()
	s2.AddScoped[[]int](Value([]int{0, 0, 0}))

	x, _ := s1.Get[[]int]()
	// assert.True(ok)

	_, ok := s1.Get[[]string]()
	assert.False(ok)

	y, ok := s2.Get[[]int]()
	assert.True(ok)

	assert.Equal([]int{1, 2, 3, 4, 5}, x)
	assert.Equal([]int{0, 0, 0}, y)
}
