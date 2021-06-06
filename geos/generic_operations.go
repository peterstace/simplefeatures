package geos

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/peterstace/simplefeatures/geom"
)

// relatesAny checks if the two geometries are related using any of the masks.
func relatesAny(g1, g2 geom.Geometry, masks ...string) (bool, error) {
	for i, m := range masks {
		r, err := relate(g1, g2, m)
		if err != nil {
			return false, wrap(err, "could not relate mask %d of %d", i+1, len(masks))
		}
		if r {
			return true, nil
		}
	}
	return false, nil
}

// relate invokes the GEOS GEOSRelatePattern function, which checks if two
// geometries are related according to a DE-9IM 'relates' mask.
func relate(g1, g2 geom.Geometry, mask string) (bool, error) {
	if g1.IsGeometryCollection() || g2.IsGeometryCollection() {
		return false, errors.New("GeometryCollection not supported")
	}
	if len(mask) != 9 {
		return false, fmt.Errorf("mask has invalid length %d (must be 9)", len(mask))
	}

	// Not all versions of GEOS can handle Z and M geometries correctly. For
	// Relates, we only need 2D geometries anyway.
	g1 = g1.Force2D()
	g2 = g2.Force2D()

	// The bytes in cmask represent a NULL terminated string, hence it's 10
	// chars long (rather than 9) and the last char is left as 0.
	var cmask [10]byte
	copy(cmask[:], mask)

	var result bool
	err := binaryOpE(g1, g2, func(h *handle, gh1, gh2 *C.GEOSGeometry) error {
		var err error
		result, err = h.boolErr(C.GEOSRelatePattern_r(
			h.context, gh1, gh2, (*C.char)(unsafe.Pointer(&cmask)),
		))
		return wrap(err, "executing GEOSRelatePattern_r")
	})
	return result, err
}

func binaryOpE(
	g1, g2 geom.Geometry,
	op func(*handle, *C.GEOSGeometry, *C.GEOSGeometry) error,
) error {
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
	opts []geom.ConstructorOption,
	op func(C.GEOSContextHandle_t, *C.GEOSGeometry, *C.GEOSGeometry) *C.GEOSGeometry,
) (geom.Geometry, error) {
	// Not all versions of GEOS can handle Z and M geometries correctly. For
	// binary operations, we only need 2D geometries anyway.
	g1 = g1.Force2D()
	g2 = g2.Force2D()

	var result geom.Geometry
	err := binaryOpE(g1, g2, func(h *handle, gh1, gh2 *C.GEOSGeometry) error {
		resultGH := op(h.context, gh1, gh2)
		if resultGH == nil {
			return h.err()
		}
		defer C.GEOSGeom_destroy(resultGH)
		var err error
		result, err = h.decode(resultGH, opts)
		return wrap(err, "decoding result")
	})
	return result, err
}

func unaryOpE(g geom.Geometry, op func(*handle, *C.GEOSGeometry) error) error {
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
	opts []geom.ConstructorOption,
	op func(C.GEOSContextHandle_t, *C.GEOSGeometry) *C.GEOSGeometry,
) (geom.Geometry, error) {
	// Not all versions of libgeos can handle Z and M geometries correctly. For
	// unary operations, we only need 2D geometries anyway.
	g = g.Force2D()

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
		result, err = h.decode(resultGH, opts)
		return wrap(err, "decoding result")
	})
	return result, err
}
