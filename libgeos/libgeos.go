package libgeos

// Package libgeos provides a cgo wrapper around the GEOS (Geometry Engine,
// Open Source) library. To use this package, you will need to install the
// GEOS library.
//
// Its purpose is to provide functionality that has been implemented in GEOS,
// but is not yet available in the simplefeatures library.
//
// This package can be used in two ways:
//
// 1. Many GEOS non-threadsafe handles can be created, and functionality used
// via those handles. This is useful if many threads need to perform geometry
// operations concurrently (each thread should use its own handle). This method
// of accessing GEOS functionality allows parallelism, but is more difficult to
// use.
//
// 2. The Global functions can be used, which share an unexported global
// handle. Usage is serialised using a mutex, so these functions are safe to
// use concurrently. This method of accessing GEOS functionaliy is easier,
// although doesn't allow parallelism.
//
// The operations in this package ignore Z and M values if they are present.

import (
	"sync"

	"github.com/peterstace/simplefeatures/geom"
)

var (
	globalMutex  sync.Mutex
	globalHandle *Handle
)

func getGlobalHandle() (*Handle, error) {
	if globalHandle != nil {
		return globalHandle, nil
	}

	var err error
	globalHandle, err = NewHandle()
	if err != nil {
		return nil, err
	}
	return globalHandle, nil
}

func executeBinaryRelation(fn func(h *Handle) (bool, error)) (bool, error) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	h, err := getGlobalHandle()
	if err != nil {
		return false, err
	}
	return fn(h)
}

// Equals returns true if and only if the input geometries are spatially equal,
// i.e. they represent exactly the same set of points.
func Equals(g1, g2 geom.Geometry) (bool, error) {
	return executeBinaryRelation(func(h *Handle) (bool, error) {
		return h.Equals(g1, g2)
	})
}

// Disjoint returns true if and only if the input geometries have no points in
// common.
func Disjoint(g1, g2 geom.Geometry) (bool, error) {
	return executeBinaryRelation(func(h *Handle) (bool, error) {
		return h.Disjoint(g1, g2)
	})
}

// Touches returns true if and only if the geometries have at least 1 point in
// common, but their interiors don't intersect.
func Touches(g1, g2 geom.Geometry) (bool, error) {
	return executeBinaryRelation(func(h *Handle) (bool, error) {
		return h.Touches(g1, g2)
	})
}

// Contains returns true if and only if geometry A contains geometry B.
// Formally, the following two conditions must hold:
//
// 1. No points of B lie on the exterior of geometry A. That is, B must only be
// in the exterior or boundary of A.
//
// 2.At least one point of the interior of B lies on the interior of A. That
// is, they can't *only* intersect at their boundaries.
func Contains(a, b geom.Geometry) (bool, error) {
	return executeBinaryRelation(func(h *Handle) (bool, error) {
		return h.Contains(a, b)
	})
}

// TODO:
//
// -- Contains
// -- Covers
//
// -- Intersects
// -- Within
// -- CoveredBy
//
// -- Crosses
// -- Overlaps
