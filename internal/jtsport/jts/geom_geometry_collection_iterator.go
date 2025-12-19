package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geom_GeometryCollectionIterator iterates over all Geometrys in a Geometry,
// (which may be either a collection or an atomic geometry). The iteration
// sequence follows a pre-order, depth-first traversal of the structure of the
// GeometryCollection (which may be nested). The original Geometry object is
// returned as well (as the first object), as are all sub-collections and atomic
// elements. It is simple to ignore the intermediate GeometryCollection objects
// if they are not needed.
type Geom_GeometryCollectionIterator struct {
	// The Geometry being iterated over.
	parent *Geom_Geometry
	// Indicates whether or not the first element (the root GeometryCollection)
	// has been returned.
	atStart bool
	// The number of Geometrys in the GeometryCollection.
	max int
	// The index of the Geometry that will be returned when Next is called.
	index int
	// The iterator over a nested Geometry, or nil if this
	// GeometryCollectionIterator is not currently iterating over a nested
	// GeometryCollection.
	subcollectionIterator *Geom_GeometryCollectionIterator
}

// Geom_NewGeometryCollectionIterator constructs an iterator over the given
// Geometry.
//
// Parameters:
//   - parent: the geometry over which to iterate; also, the first element
//     returned by the iterator.
func Geom_NewGeometryCollectionIterator(parent *Geom_Geometry) *Geom_GeometryCollectionIterator {
	return &Geom_GeometryCollectionIterator{
		parent:  parent,
		atStart: true,
		index:   0,
		max:     parent.GetNumGeometries(),
	}
}

// HasNext tests whether any geometry elements remain to be returned.
func (it *Geom_GeometryCollectionIterator) HasNext() bool {
	if it.atStart {
		return true
	}
	if it.subcollectionIterator != nil {
		if it.subcollectionIterator.HasNext() {
			return true
		}
		it.subcollectionIterator = nil
	}
	if it.index >= it.max {
		return false
	}
	return true
}

// Next gets the next geometry in the iteration sequence.
func (it *Geom_GeometryCollectionIterator) Next() *Geom_Geometry {
	// The parent GeometryCollection is the first object returned.
	if it.atStart {
		it.atStart = false
		if geom_GeometryCollectionIterator_isAtomic(it.parent) {
			it.index++
		}
		return it.parent
	}
	if it.subcollectionIterator != nil {
		if it.subcollectionIterator.HasNext() {
			return it.subcollectionIterator.Next()
		}
		it.subcollectionIterator = nil
	}
	if it.index >= it.max {
		panic("no such element")
	}
	obj := it.parent.GetGeometryN(it.index)
	it.index++
	if java.InstanceOf[*Geom_GeometryCollection](obj) {
		it.subcollectionIterator = Geom_NewGeometryCollectionIterator(obj)
		// There will always be at least one element in the sub-collection.
		return it.subcollectionIterator.Next()
	}
	return obj
}

func geom_GeometryCollectionIterator_isAtomic(geom *Geom_Geometry) bool {
	return !java.InstanceOf[*Geom_GeometryCollection](geom)
}
