package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geom_GeometryOverlay is an internal type which encapsulates the runtime
// switch to use OverlayNG, and some additional extensions for optimization and
// GeometryCollection handling.
//
// This type allows the Geometry overlay methods to be switched between the
// original algorithm and the modern OverlayNG codebase via the
// SetOverlayImpl method.
type Geom_GeometryOverlay struct{}

const Geom_GeometryOverlay_PropertyName = "jts.overlay"
const Geom_GeometryOverlay_PropertyValueNG = "ng"
const Geom_GeometryOverlay_PropertyValueOld = "old"

// Geom_GeometryOverlay_OverlayNGDefault indicates whether OverlayNG is used by
// default. Currently the original JTS overlay implementation is the default.
const Geom_GeometryOverlay_OverlayNGDefault = false

var geom_GeometryOverlay_isOverlayNG = Geom_GeometryOverlay_OverlayNGDefault

// Geom_GeometryOverlay_SetOverlayImpl sets the overlay implementation to use.
// This function is provided primarily for unit testing. It is not recommended
// to use it dynamically, since that may result in inconsistent overlay
// behaviour.
func Geom_GeometryOverlay_SetOverlayImpl(overlayImplCode string) {
	if overlayImplCode == "" {
		return
	}
	// Set flag explicitly since current value may not be default.
	geom_GeometryOverlay_isOverlayNG = Geom_GeometryOverlay_OverlayNGDefault

	if overlayImplCode == Geom_GeometryOverlay_PropertyValueNG {
		geom_GeometryOverlay_isOverlayNG = true
	}
}

func geom_GeometryOverlay_overlay(a, b *Geom_Geometry, opCode int) *Geom_Geometry {
	if geom_GeometryOverlay_isOverlayNG {
		return OperationOverlayng_OverlayNGRobust_Overlay(a, b, opCode)
	}
	return OperationOverlaySnap_SnapIfNeededOverlayOp_OverlayOp(a, b, opCode)
}

// Geom_GeometryOverlay_Difference computes the difference of two geometries.
func Geom_GeometryOverlay_Difference(a, b *Geom_Geometry) *Geom_Geometry {
	// Special case: if A.isEmpty ==> empty; if B.isEmpty ==> A.
	if a.IsEmpty() {
		return OperationOverlay_OverlayOp_CreateEmptyResult(OperationOverlay_OverlayOp_Difference, a, b, a.GetFactory())
	}
	if b.IsEmpty() {
		return a.Copy()
	}

	Geom_Geometry_CheckNotGeometryCollection(a)
	Geom_Geometry_CheckNotGeometryCollection(b)

	return geom_GeometryOverlay_overlay(a, b, OperationOverlay_OverlayOp_Difference)
}

// Geom_GeometryOverlay_Intersection computes the intersection of two geometries.
func Geom_GeometryOverlay_Intersection(a, b *Geom_Geometry) *Geom_Geometry {
	// Special case: if one input is empty ==> empty.
	if a.IsEmpty() || b.IsEmpty() {
		return OperationOverlay_OverlayOp_CreateEmptyResult(OperationOverlay_OverlayOp_Intersection, a, b, a.GetFactory())
	}

	// Compute for GCs (an inefficient algorithm, but will work).
	if a.IsGeometryCollection() {
		g2 := b
		return GeomUtil_GeometryCollectionMapper_Map(
			java.Cast[*Geom_GeometryCollection](a),
			func(g *Geom_Geometry) *Geom_Geometry {
				return g.Intersection(g2)
			},
		)
	}

	return geom_GeometryOverlay_overlay(a, b, OperationOverlay_OverlayOp_Intersection)
}

// Geom_GeometryOverlay_SymDifference computes the symmetric difference of two
// geometries.
func Geom_GeometryOverlay_SymDifference(a, b *Geom_Geometry) *Geom_Geometry {
	// Handle empty geometry cases.
	if a.IsEmpty() || b.IsEmpty() {
		// Both empty - check dimensions.
		if a.IsEmpty() && b.IsEmpty() {
			return OperationOverlay_OverlayOp_CreateEmptyResult(OperationOverlay_OverlayOp_SymDifference, a, b, a.GetFactory())
		}

		// Special case: if either input is empty ==> result = other arg.
		if a.IsEmpty() {
			return b.Copy()
		}
		if b.IsEmpty() {
			return a.Copy()
		}
	}

	Geom_Geometry_CheckNotGeometryCollection(a)
	Geom_Geometry_CheckNotGeometryCollection(b)
	return geom_GeometryOverlay_overlay(a, b, OperationOverlay_OverlayOp_SymDifference)
}

// Geom_GeometryOverlay_Union computes the union of two geometries.
func Geom_GeometryOverlay_Union(a, b *Geom_Geometry) *Geom_Geometry {
	// Handle empty geometry cases.
	if a.IsEmpty() || b.IsEmpty() {
		if a.IsEmpty() && b.IsEmpty() {
			return OperationOverlay_OverlayOp_CreateEmptyResult(OperationOverlay_OverlayOp_Union, a, b, a.GetFactory())
		}

		// Special case: if either input is empty ==> other input.
		if a.IsEmpty() {
			return b.Copy()
		}
		if b.IsEmpty() {
			return a.Copy()
		}
	}

	Geom_Geometry_CheckNotGeometryCollection(a)
	Geom_Geometry_CheckNotGeometryCollection(b)

	return geom_GeometryOverlay_overlay(a, b, OperationOverlay_OverlayOp_Union)
}

// Geom_GeometryOverlay_UnionSelf computes the union of a single geometry.
func Geom_GeometryOverlay_UnionSelf(a *Geom_Geometry) *Geom_Geometry {
	if geom_GeometryOverlay_isOverlayNG {
		return OperationOverlayng_OverlayNGRobust_Union(a)
	}
	return OperationUnion_UnaryUnionOp_Union(a)
}
