package ioc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValue(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddScoped[string](Value("value"))

	assert.Equal("value", c.MustGet[string]())
}

func TestResolve(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()
	service := &syntacticSugarService{name: "default"}

	c.AddScoped[*syntacticSugarService](Value(service))
	c.AddScoped[syntacticSugarInterface](Resolve[*syntacticSugarService]())

	assert.Same(service, c.MustGet[syntacticSugarInterface]())
}

func TestResolveNamed(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()
	namedService := &syntacticSugarService{name: "named"}

	c.AddNamedScoped[*syntacticSugarService]("named", Value(namedService))
	c.AddScoped[*syntacticSugarService](Value(&syntacticSugarService{name: "default"}))
	c.AddScoped[syntacticSugarInterface](ResolveNamed[*syntacticSugarService]("named"))

	assert.Same(namedService, c.MustGet[syntacticSugarInterface]())
}

func TestBindSingleton(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddTransient[*syntacticSugarService](newSyntacticSugarService)
	assert.NoError(c.BindSingleton[syntacticSugarInterface, *syntacticSugarService]())

	first := c.MustGet[syntacticSugarInterface]()
	second := c.MustGet[syntacticSugarInterface]()

	assert.Same(first, second)
}

func TestBindNamedSingleton(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()
	service := &syntacticSugarService{name: "named"}

	c.AddScoped[*syntacticSugarService](Value(service))
	assert.NoError(c.BindNamedSingleton[syntacticSugarInterface, *syntacticSugarService]("service"))

	first := c.MustGetNamed[syntacticSugarInterface]("service")
	second := c.MustGetNamed[syntacticSugarInterface]("service")

	assert.Same(service, first)
	assert.Same(first, second)
	assert.Same(service, c.MustGet[syntacticSugarInterface]())
}

func TestBindScoped(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddTransient[*syntacticSugarService](newSyntacticSugarService)
	assert.NoError(c.BindScoped[syntacticSugarInterface, *syntacticSugarService]())

	rootFirst := c.MustGet[syntacticSugarInterface]()
	rootSecond := c.MustGet[syntacticSugarInterface]()
	child := c.CreateScope()
	childFirst := child.MustGet[syntacticSugarInterface]()
	childSecond := child.MustGet[syntacticSugarInterface]()

	assert.Same(rootFirst, rootSecond)
	assert.NotSame(rootFirst, childFirst)
	assert.Same(childFirst, childSecond)
}

func TestBindNamedScoped(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddTransient[*syntacticSugarService](newSyntacticSugarService)
	assert.NoError(c.BindNamedScoped[syntacticSugarInterface, *syntacticSugarService]("service"))

	root := c.MustGetNamed[syntacticSugarInterface]("service")
	child := c.CreateScope()
	childFirst := child.MustGetNamed[syntacticSugarInterface]("service")
	childSecond := child.MustGetNamed[syntacticSugarInterface]("service")

	assert.NotSame(root, childFirst)
	assert.Same(childFirst, childSecond)
}

func TestBindTransient(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddTransient[*syntacticSugarService](newSyntacticSugarService)
	assert.NoError(c.BindTransient[syntacticSugarInterface, *syntacticSugarService]())

	first := c.MustGet[syntacticSugarInterface]()
	second := c.MustGet[syntacticSugarInterface]()

	assert.NotSame(first, second)
}

func TestBindNamedTransient(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()

	c.AddTransient[*syntacticSugarService](newSyntacticSugarService)
	assert.NoError(c.BindNamedTransient[syntacticSugarInterface, *syntacticSugarService]("service"))

	first := c.MustGetNamed[syntacticSugarInterface]("service")
	second := c.MustGetNamed[syntacticSugarInterface]("service")

	assert.NotSame(first, second)
}

type syntacticSugarInterface interface {
	Name() string
}

type syntacticSugarService struct {
	name string
}

func (s *syntacticSugarService) Name() string {
	return s.name
}

func newSyntacticSugarService() *syntacticSugarService {
	return &syntacticSugarService{}
}
