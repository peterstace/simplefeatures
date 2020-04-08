package geos

// Package geos provides a cgo wrapper around the GEOS (Geometry Engine,
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

func executeBoolOp(fn func(h *Handle) (bool, error)) (bool, error) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	h, err := getGlobalHandle()
	if err != nil {
		return false, err
	}
	return fn(h)
}

func executeGeomOp(fn func(h *Handle) (geom.Geometry, error)) (geom.Geometry, error) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	h, err := getGlobalHandle()
	if err != nil {
		return geom.Geometry{}, err
	}
	return fn(h)
}

// Equals returns true if and only if the input geometries are spatially equal,
// i.e. they represent exactly the same set of points.
func Equals(g1, g2 geom.Geometry) (bool, error) {
	return executeBoolOp(func(h *Handle) (bool, error) {
		return h.Equals(g1, g2)
	})
}

// Disjoint returns true if and only if the input geometries have no points in
// common.
func Disjoint(g1, g2 geom.Geometry) (bool, error) {
	return executeBoolOp(func(h *Handle) (bool, error) {
		return h.Disjoint(g1, g2)
	})
}

// Touches returns true if and only if the geometries have at least 1 point in
// common, but their interiors don't intersect.
func Touches(g1, g2 geom.Geometry) (bool, error) {
	return executeBoolOp(func(h *Handle) (bool, error) {
		return h.Touches(g1, g2)
	})
}

// Contains returns true if and only if geometry A contains geometry B.
// See the global Contains function for details.
// Formally, the following two conditions must hold:
//
// 1. No points of B lies on the exterior of geometry A. That is, B must only be
// in the exterior or boundary of A.
//
// 2 .At least one point of the interior of B lies on the interior of A. That
// is, they can't *only* intersect at their boundaries.
func Contains(a, b geom.Geometry) (bool, error) {
	return executeBoolOp(func(h *Handle) (bool, error) {
		return h.Contains(a, b)
	})
}

// Covers returns true if and only if geometry A covers geometry B. Formally,
// the following two conditions must hold:
//
// 1. No points of B lies on the exterior of geometry A. That is, B must only be
// in the exterior or boundary of A.
//
// 2. At least one point of B lies on A (either its interior or boundary).
func Covers(a, b geom.Geometry) (bool, error) {
	return executeBoolOp(func(h *Handle) (bool, error) {
		return h.Covers(a, b)
	})
}

// Intersects returns true if and only if the geometries share at least one
// point in common.
func Intersects(a, b geom.Geometry) (bool, error) {
	return executeBoolOp(func(h *Handle) (bool, error) {
		return h.Intersects(a, b)
	})
}

// Within returns true if and only if geometry A is completely within geometry
// B. Formally, the following two conditions must hold:
//
// 1. No points of A lies on the exterior of geometry B. That is, A must only be
// in the exterior or boundary of B.
//
// 2.At least one point of the interior of A lies on the interior of B. That
// is, they can't *only* intersect at their boundaries.
func Within(a, b geom.Geometry) (bool, error) {
	return executeBoolOp(func(h *Handle) (bool, error) {
		return h.Within(a, b)
	})
}

// CoveredBy returns true if and only if geometry A is covered by geometry B.
// Formally, the following two conditions must hold:
//
// 1. No points of A lies on the exterior of geometry B. That is, A must only be
// in the exterior or boundary of B.
//
// 2. At least one point of A lies on B (either its interior or boundary).
func CoveredBy(a, b geom.Geometry) (bool, error) {
	return executeBoolOp(func(h *Handle) (bool, error) {
		return h.CoveredBy(a, b)
	})
}

// Crosses returns true if and only if geometry A and B cross each other.
// Formally, the following conditions must hold:
//
// 1. The geometries must have some but not all interior points in common.
//
// 2. The dimensionality of the intersection must be less than the maximum
// dimension of the input geometries.
//
// 3. The intersection must not equal either of the input geometries.
func Crosses(a, b geom.Geometry) (bool, error) {
	return executeBoolOp(func(h *Handle) (bool, error) {
		return h.Crosses(a, b)
	})
}

// Overlaps returns true if and only if geometry A and B overlap with each
// other. Formally, the following conditions must hold:
//
// 1. The geometries must have the same dimension.
//
// 2. The geometries must have some but not all points in common.
//
// 3. The intersection of the geometries must have the same dimension as the
// geometries themselves.
func Overlaps(a, b geom.Geometry) (bool, error) {
	return executeBoolOp(func(h *Handle) (bool, error) {
		return h.Overlaps(a, b)
	})
}

// Union returns a geometry that that is the union of the input geometries.
// Formally, the returned geometry will contain a particular point X if and
// only if X is present in either geometry (or both).
func Union(a, b geom.Geometry) (geom.Geometry, error) {
	return executeGeomOp(func(h *Handle) (geom.Geometry, error) {
		return h.Union(a, b)
	})
}

// Intersection returns a geometry that is the intersection of the input
// geometries. Formally, the returned geometry will contain a particular point
// X if and only if X is present in both geometries.
func Intersection(a, b geom.Geometry) (geom.Geometry, error) {
	return executeGeomOp(func(h *Handle) (geom.Geometry, error) {
		return h.Intersection(a, b)
	})
}

// Buffer returns a geometry that contains all points within the given radius
// of the input geometry.
func Buffer(g geom.Geometry, radius float64) (geom.Geometry, error) {
	return executeGeomOp(func(h *Handle) (geom.Geometry, error) {
		return h.Buffer(g, radius)
	})
}

// Simplify creates a simplified version of a geometry using the
// Douglas-Peucker algorithm. Topological invariants may not be maintained,
// e.g. polygons can collapse into linestrings, and holes in polygons may be
// lost.
func Simplify(g geom.Geometry, tolerance float64) (geom.Geometry, error) {
	return executeGeomOp(func(h *Handle) (geom.Geometry, error) {
		return h.Simplify(g, tolerance)
	})
}
