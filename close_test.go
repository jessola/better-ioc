package ioc

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type closeTestValue struct {
	name   string
	closed *[]string
	err    error
}

func (v *closeTestValue) Close() error {
	*v.closed = append(*v.closed, v.name)
	return v.err
}

type closeTestExecer interface {
	Exec()
}

type closeTestDB struct {
	*closeTestValue
}

func (*closeTestDB) Exec() {}

func TestCloseClosesValuesInReverseCreationOrder(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()
	closed := []string{}
	firstErr := errors.New("first close failed")
	secondErr := errors.New("second close failed")
	first := &closeTestValue{name: "first", closed: &closed, err: firstErr}
	second := &closeTestValue{name: "second", closed: &closed, err: secondErr}

	assert.NoError(c.AddTransient[*closeTestValue](Value(first)))
	assert.NoError(c.AddTransient[*closeTestValue](Value(second)))
	_, ok := c.Get[[]*closeTestValue]()
	assert.True(ok)

	err := c.Close()

	assert.Equal([]string{"second", "first"}, closed)
	assert.ErrorIs(err, firstErr)
	assert.ErrorIs(err, secondErr)
	assert.NoError(c.Close())
	assert.Equal([]string{"second", "first"}, closed)
}

func TestCloseOnlyClosesSharedValueOnce(t *testing.T) {
	assert := assert.New(t)
	c := NewContainer()
	closed := []string{}
	value := &closeTestValue{name: "shared", closed: &closed}

	assert.NoError(c.AddScoped[*closeTestValue](Value(value)))
	assert.NoError(c.AddScoped[io.Closer](Value(io.Closer(value))))
	c.MustGet[*closeTestValue]()
	c.MustGet[io.Closer]()

	assert.NoError(c.Close())
	assert.Equal([]string{"shared"}, closed)
}

func TestCloseDoesNotCloseSingletonFromChildScope(t *testing.T) {
	assert := assert.New(t)
	root := NewContainer()
	child := root.CreateScope()
	closed := []string{}
	value := &closeTestValue{name: "singleton", closed: &closed}

	assert.NoError(root.AddSingleton[*closeTestValue](Value(value)))
	assert.Same(value, child.MustGet[*closeTestValue]())

	assert.NoError(child.Close())
	assert.Empty(closed)
	assert.Same(value, root.MustGet[*closeTestValue]())

	assert.NoError(root.Close())
	assert.Equal([]string{"singleton"}, closed)
}

func TestCloseDoesNotCloseSingletonReturnedByScopedProvider(t *testing.T) {
	assert := assert.New(t)
	root := NewContainer()
	child := root.CreateScope()
	closed := []string{}
	db := &closeTestDB{closeTestValue: &closeTestValue{name: "database", closed: &closed}}

	assert.NoError(root.AddSingleton[*closeTestDB](Value(db)))
	assert.NoError(root.AddScoped[closeTestExecer](func(db *closeTestDB) closeTestExecer {
		return db
	}))

	assert.Same(db, child.MustGet[closeTestExecer]())
	assert.NoError(child.Close())
	assert.Empty(closed)

	assert.NoError(root.Close())
	assert.Equal([]string{"database"}, closed)
}
