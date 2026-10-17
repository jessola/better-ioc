package ioc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransient(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddTransient[*time.Time](func() *time.Time {
		now := time.Now()
		return &now
	})

	s := c.CreateScope()

	x, ok := s.Get[*time.Time]()
	assert.True(ok)

	y, ok := s.Get[*time.Time]()
	assert.True(ok)

	assert.NotEqual(x, y)
}

func TestNamedTransient(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddNamedTransient[string]("TEST NAME", Value("bar"))
	c.AddTransient[string](Value("foo"))

	s := c.CreateScope()

	x, ok := s.Get[string]()
	assert.True(ok)

	y, ok := s.GetNamed[string]("TEST NAME")
	assert.True(ok)

	assert.Equal("foo", x)
	assert.Equal("bar", y)
}
