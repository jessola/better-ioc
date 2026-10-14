package ioc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransient(t *testing.T) {
	assert := assert.New(t)
	c := New()

	c.AddTransient[*time.Time](func() *time.Time {
		t := time.Now()
		return &t
	})

	s := c.CreateScope()

	x, ok := s.Get[*time.Time]()
	assert.True(ok)

	y, ok := s.Get[*time.Time]()
	assert.True(ok)

	assert.NotEqual(x, y)
}

func TestSingleton(t *testing.T) {
	assert := assert.New(t)
	c := New()

	c.AddSingleton[*time.Time](func() *time.Time {
		t := time.Now()
		return &t
	})

	x, ok := c.Get[*time.Time]()
	assert.True(ok)

	s := c.CreateScope()

	y, ok := s.Get[*time.Time]()
	assert.True(ok)

	z, ok := s.Get[*time.Time]()
	assert.True(ok)

	assert.Equal(x, y)
	assert.Equal(y, z)
}

func TestScoped(t *testing.T) {
	assert := assert.New(t)
	c := New()

	c.AddScoped[*time.Time](func() *time.Time {
		t := time.Now()
		return &t
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

func TestRecursion(t *testing.T) {
	assert := assert.New(t)
	c := New()

	c.AddSingleton[int](func() int { return 4 })

	s := c.CreateScope().CreateScope().CreateScope().CreateScope()

	x, ok := s.Get[int]()
	assert.True(ok)
	assert.Equal(4, x)
}

func TestScopeOverrides(t *testing.T) {
	assert := assert.New(t)
	c := New()

	c.AddScoped[int](func() int { return 0 })
	c.AddScoped[int](func() int { return 42 })

	s1 := c.CreateScope()
	s1.AddScoped[int](func() int { return 100 })

	s2 := c.CreateScope()
	s2.AddScoped[int](func() int { return -5 })
	s2.AddScoped[int](func() int { return 9000 })

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

func TestGetSlice(t *testing.T) {
	assert := assert.New(t)
	c := New()

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
