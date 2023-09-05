package geos

/*
#cgo LDFLAGS: -lgeos_c
#cgo CFLAGS: -Wall
#include "geos_c.h"
*/
import "C"

import (
	"github.com/peterstace/simplefeatures/geom"
)

func binaryOpE(
	g1, g2 geom.Geometry,
	op func(*handle, *C.GEOSGeometry, *C.GEOSGeometry) error,
) error {
	// Not all versions of GEOS can handle Z and M geometries correctly. For
	// binary operations, we only need 2D geometries anyway.
	g1 = g1.Force2D()
	g2 = g2.Force2D()

	h, err := newHandle()
	if err != nil {
		return err
	}
	defer h.release()

	gh1, err := h.createGeometryHandle(g1)
	if err != nil {
		return wrap(err, "converting first geom argument")
	}
	defer C.GEOSGeom_destroy(gh1)

	gh2, err := h.createGeometryHandle(g2)
	if err != nil {
		return wrap(err, "converting second geom argument")
	}
	defer C.GEOSGeom_destroy(gh2)

	return op(h, gh1, gh2)
}

func binaryOpG(
	g1, g2 geom.Geometry,
	op func(C.GEOSContextHandle_t, *C.GEOSGeometry, *C.GEOSGeometry) *C.GEOSGeometry,
) (geom.Geometry, error) {
	var result geom.Geometry
	err := binaryOpE(g1, g2, func(h *handle, gh1, gh2 *C.GEOSGeometry) error {
		resultGH := op(h.context, gh1, gh2)
		if resultGH == nil {
			return h.err()
		}
		defer C.GEOSGeom_destroy(resultGH)
		var err error
		result, err = h.decode(resultGH)
		return wrap(err, "decoding result")
	})
	return result, err
}

func binaryOpB(
	g1, g2 geom.Geometry,
	op func(C.GEOSContextHandle_t, *C.GEOSGeometry, *C.GEOSGeometry) C.char,
) (bool, error) {
	var result bool
	err := binaryOpE(g1, g2, func(h *handle, gh1, gh2 *C.GEOSGeometry) error {
		var err error
		result, err = h.boolErr(op(h.context, gh1, gh2))
		return err
	})
	return result, err
}

func unaryOpE(g geom.Geometry, op func(*handle, *C.GEOSGeometry) error) error {
	// Not all versions of libgeos can handle Z and M geometries correctly. For
	// unary operations, we only need 2D geometries anyway.
	g = g.Force2D()

	h, err := newHandle()
	if err != nil {
		return err
	}
	defer h.release()

	gh, err := h.createGeometryHandle(g)
	if err != nil {
		return wrap(err, "converting geom argument")
	}
	defer C.GEOSGeom_destroy(gh)

	return op(h, gh)
}

func unaryOpG(
	g geom.Geometry,
	op func(C.GEOSContextHandle_t, *C.GEOSGeometry) *C.GEOSGeometry,
) (geom.Geometry, error) {
	var result geom.Geometry
	err := unaryOpE(g, func(h *handle, gh *C.GEOSGeometry) error {
		resultGH := op(h.context, gh)
		if resultGH == nil {
			return h.err()
		}
		if gh != resultGH {
			// gh and resultGH will be the same if op is the noop function that
			// just returns its input. We need to avoid destroying resultGH in
			// that case otherwise we will do a double-free.
			defer C.GEOSGeom_destroy(resultGH)
		}
		var err error
		result, err = h.decode(resultGH)
		return wrap(err, "decoding result")
	})
	return result, err
}
