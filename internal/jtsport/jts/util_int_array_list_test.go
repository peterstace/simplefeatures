package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestIntArrayListEmpty(t *testing.T) {
	iar := jts.Util_NewIntArrayList()
	junit.AssertEquals(t, 0, iar.Size())
}

func TestIntArrayListAddFew(t *testing.T) {
	iar := jts.Util_NewIntArrayList()
	iar.Add(1)
	iar.Add(2)
	iar.Add(3)
	junit.AssertEquals(t, 3, iar.Size())

	data := iar.ToArray()
	junit.AssertEquals(t, 3, len(data))
	junit.AssertEquals(t, 1, data[0])
	junit.AssertEquals(t, 2, data[1])
	junit.AssertEquals(t, 3, data[2])
}

func TestIntArrayListAddMany(t *testing.T) {
	iar := jts.Util_NewIntArrayListWithCapacity(20)

	max := 100
	for i := 0; i < max; i++ {
		iar.Add(i)
	}

	junit.AssertEquals(t, max, iar.Size())

	data := iar.ToArray()
	junit.AssertEquals(t, max, len(data))
	for j := 0; j < max; j++ {
		junit.AssertEquals(t, j, data[j])
	}
}

func TestIntArrayListAddAll(t *testing.T) {
	iar := jts.Util_NewIntArrayList()

	iar.AddAll(nil)
	iar.AddAll([]int{})
	iar.AddAll([]int{1, 2, 3})
	junit.AssertEquals(t, 3, iar.Size())

	data := iar.ToArray()
	junit.AssertEquals(t, 3, len(data))
	junit.AssertEquals(t, 1, data[0])
	junit.AssertEquals(t, 2, data[1])
	junit.AssertEquals(t, 3, data[2])
}
