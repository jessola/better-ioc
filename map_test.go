package ioc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendToMap(t *testing.T) {
	assert := assert.New(t)
	testMap := map[string][]int{
		"foo": {0},
	}

	assert.Len(testMap, 1)
	appendToSliceInMap(testMap, "bar", 8)
	appendToSliceInMap(testMap, "foo", 0, 0)
	assert.Len(testMap, 2)

	assert.Equal([]int{0, 0, 0}, testMap["foo"])
	assert.Equal([]int{8}, testMap["bar"])
}
