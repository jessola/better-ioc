package ioc

import (
	"io"
	"reflect"
)

func (c *Container) trackValue(v any) {
	c.uniqueValsMu.Lock()
	defer c.uniqueValsMu.Unlock()

	if _, ok := v.(io.Closer); !ok {
		return
	}
	if !reflect.TypeOf(v).Comparable() {
		return
	}
	// if _, ok := c.uniqueValMap[v]; ok {
	// 	return
	// }
	if !c.isValGloballyUnique(v) {
		return
	}

	c.uniqueValMap[v] = true
	c.uniqueVals = append(c.uniqueVals, v)
}

func (c *Container) isValGloballyUnique(v any) bool {
	if c.parent != nil && !c.parent.isValGloballyUnique(v) {
		return false
	}
	_, notUnique := c.uniqueValMap[v]
	return notUnique == false // easier to get than !notUnique
}
