package ioc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultContainerConvenienceFunctions(t *testing.T) {
	assert := assert.New(t)
	original := DefaultContainer
	DefaultContainer = NewContainer()
	t.Cleanup(func() {
		DefaultContainer = original
	})

	next := 0
	assert.NoError(AddSingleton[int](Value(1)))
	assert.NoError(AddScoped[string](Value("scoped")))
	assert.NoError(AddTransient[*int](func() *int {
		next++
		value := next
		return &value
	}))
	assert.NoError(AddNamedSingleton[float64]("singleton", Value(1.5)))
	assert.NoError(AddNamedScoped[bool]("scoped", Value(true)))
	assert.NoError(AddNamedTransient[byte]("transient", Value(byte(2))))

	boundService := &defaultContainerBoundService{}
	assert.NoError(AddScoped[*defaultContainerBoundService](Value(boundService)))
	assert.NoError(BindSingleton[defaultContainerBound, *defaultContainerBoundService]())
	assert.NoError(BindNamedSingleton[defaultContainerBound, *defaultContainerBoundService]("singleton"))
	assert.NoError(BindScoped[defaultContainerBound, *defaultContainerBoundService]())
	assert.NoError(BindNamedScoped[defaultContainerBound, *defaultContainerBoundService]("scoped"))
	assert.NoError(BindTransient[defaultContainerBound, *defaultContainerBoundService]())
	assert.NoError(BindNamedTransient[defaultContainerBound, *defaultContainerBoundService]("transient"))

	s := CreateScope()

	assert.Equal(1, s.MustGet[int]())
	assert.Equal("scoped", s.MustGet[string]())
	assert.NotSame(s.MustGet[*int](), s.MustGet[*int]())
	assert.Equal(1.5, s.MustGetNamed[float64]("singleton"))
	assert.True(s.MustGetNamed[bool]("scoped"))
	assert.Equal(byte(2), s.MustGetNamed[byte]("transient"))
	assert.Same(boundService, s.MustGet[defaultContainerBound]())
	assert.Same(boundService, s.MustGetNamed[defaultContainerBound]("singleton"))
	assert.Same(boundService, s.MustGetNamed[defaultContainerBound]("scoped"))
	assert.Same(boundService, s.MustGetNamed[defaultContainerBound]("transient"))
}

type defaultContainerBound interface {
	isDefaultContainerBound()
}

type defaultContainerBoundService struct{}

func (*defaultContainerBoundService) isDefaultContainerBound() {}
