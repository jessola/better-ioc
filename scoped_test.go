package ioc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScoped(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddScoped[*time.Time](func() *time.Time {
		now := time.Now()
		return &now
	})

	x, ok := c.Get[*time.Time]()
	assert.True(ok)

	s := c.CreateScope()

	y, ok := s.Get[*time.Time]()
	assert.True(ok)

	z, ok := s.Get[*time.Time]()
	assert.True(ok)

	assert.NotEqual(x, y)
	assert.Equal(y, z)
}

func TestNamedScoped(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddNamedScoped[string]("TEST NAME", Value("bar"))
	c.AddScoped[string](Value("foo"))

	s := c.CreateScope()

	x, ok := s.Get[string]()
	assert.True(ok)

	y, ok := s.GetNamed[string]("TEST NAME")
	assert.True(ok)

	assert.Equal("foo", x)
	assert.Equal("bar", y)
}
