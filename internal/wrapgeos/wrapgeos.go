package wrapgeos

/*
#cgo LDFLAGS: -lgeos_c
#cgo CFLAGS: -Wall
#include "geos_c.h"

typedef GEOSGeometry* (*sf_binary_op_g_func) (GEOSContextHandle_t, GEOSGeometry*, GEOSGeometry*);

GEOSGeometry* exec_binary_op_g(GEOSContextHandle_t handle, GEOSGeometry* g1, GEOSGeometry* g2, sf_binary_op_g_func op) {
	return op(handle, g1, g2);
}
*/
import "C"
import (
	"github.com/peterstace/simplefeatures/geom"
)

// Version returns the GEOS version as (major, minor, patch).
func Version() (int, int, int) {
	return C.GEOS_VERSION_MAJOR, C.GEOS_VERSION_MINOR, C.GEOS_VERSION_PATCH
}

func BinaryOpG(
	g1, g2 geom.Geometry,
	//op func(_ C.GEOSContextHandle_t, _, _ *C.GEOSGeometry) *C.GEOSGeometry,
	op C.sf_binary_op_g_func,
) (geom.Geometry, error) {
	// Not all versions of GEOS can handle Z and M geometries correctly. For
	// binary operations, we only need 2D geometries anyway.
	g1 = g1.Force2D()
	g2 = g2.Force2D()

	h, err := newHandle()
	if err != nil {
		return geom.Geometry{}, err
	}
	defer h.release()

	gh1, err := h.createGeometryHandle(g1)
	if err != nil {
		return geom.Geometry{}, wrap(err, "converting first geom argument")
	}
	defer C.GEOSGeom_destroy(gh1)

	gh2, err := h.createGeometryHandle(g2)
	if err != nil {
		return geom.Geometry{}, wrap(err, "converting second geom argument")
	}
	defer C.GEOSGeom_destroy(gh2)

	resultGH := C.exec_binary_op_g(h.context, gh1, gh2, op)
	if resultGH == nil {
		return geom.Geometry{}, h.err()
	}
	defer C.GEOSGeom_destroy(resultGH)

	result, err := h.decode(resultGH)
	return result, wrap(err, "decoding result")
}

func Union(g1, g2 geom.Geometry) (geom.Geometry, error) {
	r, err := BinaryOpG(g1, g2, C.sf_binary_op_g_func(C.GEOSUnion_r))
	return r, wrap(err, "executing GEOSUnion_r")
}
