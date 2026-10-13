package ioc

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddingToScopeLoadsParentProviders(t *testing.T) {
	assert := assert.New(t)
	c := New()

	intType := reflect.TypeFor[int]()

	c.add(intType, Scoped, func() int {
		return 1
	})
	c.add(intType, Scoped, func() int {
		return 2
	})

	assert.Len(c.services[intType], 2)

	s := c.CreateScope()
	s2 := c.CreateScope()
	s.add(intType, Scoped, func() int {
		return 3
	})

	s2.get(intType) // Load the services for the type, so we can count them

	assert.Len(c.services[intType], 2)
	assert.Len(s.services[intType], 3)
	assert.Len(s2.services[intType], 2)
}

func TestTransient(t *testing.T) {
	assert := assert.New(t)
	c := New()

	timePtr := reflect.TypeFor[*time.Time]()

	c.add(timePtr, Transient, func() *time.Time {
		t := time.Now()
		return &t
	})

	s := c.CreateScope()

	t1, ok := c.get(timePtr)
	assert.True(ok)

	t2, ok := s.get(timePtr)
	assert.True(ok)

	t3, ok := s.get(timePtr)
	assert.True(ok)

	assert.NotNil(t1)
	assert.NotNil(t2)
	assert.NotNil(t3)

	assert.NotEqual(t1, t2)
	assert.NotEqual(t2, t3)
	assert.NotEqual(t3, t1)
}

func TestSingleton(t *testing.T) {
	assert := assert.New(t)
	c := New()

	timePtr := reflect.TypeFor[*time.Time]()

	c.add(timePtr, Singleton, func() *time.Time {
		t := time.Now()
		return &t
	})

	s := c.CreateScope()

	t1, ok := c.get(timePtr)
	assert.True(ok)

	t2, ok := c.get(timePtr)
	assert.True(ok)

	t3, ok := s.get(timePtr)
	assert.True(ok)

	t4, ok := s.get(timePtr)
	assert.True(ok)

	assert.NotNil(t1)
	assert.NotNil(t2)
	assert.NotNil(t3)
	assert.NotNil(t4)

	assert.Equal(t1, t2, "t1,t2")
	assert.Equal(t1, t3, "t1,t3")
	assert.Equal(t1, t4, "t1,t4")
	assert.Equal(t2, t3, "t2,t3")
	assert.Equal(t2, t4, "t2,t4")
	assert.Equal(t3, t4, "t3,t4")
}

func TestScoped(t *testing.T) {
	assert := assert.New(t)
	c := New()

	timePtr := reflect.TypeFor[*time.Time]()

	c.add(timePtr, Scoped, func() *time.Time {
		t := time.Now()
		return &t
	})

	s := c.CreateScope()

	t1, ok := c.get(timePtr)
	assert.True(ok)

	t2, ok := c.get(timePtr)
	assert.True(ok)

	t3, ok := s.get(timePtr)
	assert.True(ok)

	t4, ok := s.get(timePtr)
	assert.True(ok)

	assert.NotNil(t1)
	assert.NotNil(t2)
	assert.NotNil(t3)
	assert.NotNil(t4)

	assert.Equal(t1, t2, "t1 should equal t2")
	assert.Equal(t3, t4, "t3 should equal t4")
	assert.NotEqual(t1, t3, "t1 should NOT equal t3")
}

func TestGetSlice(t *testing.T) {
	assert := assert.New(t)
	c := New()

	intType := reflect.TypeFor[int]()
	intSliceType := reflect.TypeFor[[]int]()

	c.add(intType, Singleton, func() int { return 1 })
	c.add(intType, Scoped, func() int { return 2 })
	c.add(intType, Transient, func() int { return 3 })

	s1 := c.CreateScope()
	s2 := c.CreateScope()

	s2.add(intSliceType, Singleton, func() []int {
		return []int{0, 0, 0, 0}
	})

	i1, ok := s1.get(intType)
	assert.True(ok)

	i2, ok := s2.get(intType)
	assert.True(ok)

	assert.Len(c.services, 1)

	slice1, ok := s1.get(intSliceType)
	assert.True(ok)

	slice2, ok := s2.get(intSliceType)
	assert.True(ok)

	assert.Equal([]any{1, 2, 3}, slice1, "slice1 should equal []int{1,2,3}")
	assert.Equal([]int{0, 0, 0, 0}, slice2, "slice2 should equal []int{0,0,0,0}")
	assert.Equal(i1, i2)
}
