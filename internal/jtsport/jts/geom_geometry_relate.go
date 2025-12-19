package jts

// geom_GeometryRelate provides internal helper functions for computing
// topological relationships between geometries.
//
// This is the Go equivalent of Java's GeometryRelate internal class.
// It supports switching between the original RelateOp algorithm and
// the modern RelateNG codebase via the SetRelateImpl function.

// geom_GeometryRelate_NG_DEFAULT indicates whether RelateNG is the default
// implementation. Currently the old relate implementation is the default.
const geom_GeometryRelate_NG_DEFAULT = false

// geom_geometryRelate_isRelateNG controls which relate implementation is used.
var geom_geometryRelate_isRelateNG = geom_GeometryRelate_NG_DEFAULT

// Geom_GeometryRelate_SetRelateImpl sets the relate implementation to use.
// Pass "ng" to use RelateNG, or "old" to use the original RelateOp.
// Other values are ignored and the default is retained.
// Note: It is not recommended to change this dynamically, since that may
// result in inconsistent relate behaviour.
func Geom_GeometryRelate_SetRelateImpl(relateImplCode string) {
	if relateImplCode == "" {
		return
	}
	// Reset to default since current value may not be default.
	geom_geometryRelate_isRelateNG = geom_GeometryRelate_NG_DEFAULT

	if relateImplCode == "ng" || relateImplCode == "NG" {
		geom_geometryRelate_isRelateNG = true
	}
}

func geom_GeometryRelate_Intersects(a, b *Geom_Geometry) bool {
	if geom_geometryRelate_isRelateNG {
		return OperationRelateng_RelateNG_Relate(a, b, OperationRelateng_RelatePredicate_Intersects())
	}
	if a.IsGeometryCollection() || b.IsGeometryCollection() {
		for i := 0; i < a.GetNumGeometries(); i++ {
			for j := 0; j < b.GetNumGeometries(); j++ {
				if a.GetGeometryN(i).Intersects(b.GetGeometryN(j)) {
					return true
				}
			}
		}
		return false
	}
	return OperationRelate_RelateOp_Relate(a, b).IsIntersects()
}

func geom_GeometryRelate_Contains(a, b *Geom_Geometry) bool {
	if geom_geometryRelate_isRelateNG {
		return OperationRelateng_RelateNG_Relate(a, b, OperationRelateng_RelatePredicate_Contains())
	}
	// Optimization - lower dimension cannot contain areas.
	if b.GetDimension() == 2 && a.GetDimension() < 2 {
		return false
	}
	// Optimization - P cannot contain a non-zero-length L.
	// Note that a point can contain a zero-length lineal geometry,
	// since the line has no boundary due to Mod-2 Boundary Rule.
	if b.GetDimension() == 1 && a.GetDimension() < 1 && b.GetLength() > 0.0 {
		return false
	}
	// Optimization - envelope test.
	if !a.GetEnvelopeInternal().ContainsEnvelope(b.GetEnvelopeInternal()) {
		return false
	}
	return OperationRelate_RelateOp_Relate(a, b).IsContains()
}

func geom_GeometryRelate_Covers(a, b *Geom_Geometry) bool {
	if geom_geometryRelate_isRelateNG {
		return OperationRelateng_RelateNG_Relate(a, b, OperationRelateng_RelatePredicate_Covers())
	}
	// Optimization - lower dimension cannot cover areas.
	if b.GetDimension() == 2 && a.GetDimension() < 2 {
		return false
	}
	// Optimization - P cannot cover a non-zero-length L.
	// Note that a point can cover a zero-length lineal geometry.
	if b.GetDimension() == 1 && a.GetDimension() < 1 && b.GetLength() > 0.0 {
		return false
	}
	// Optimization - envelope test.
	if !a.GetEnvelopeInternal().CoversEnvelope(b.GetEnvelopeInternal()) {
		return false
	}
	// Optimization for rectangle arguments.
	if a.IsRectangle() {
		// Since we have already tested that the test envelope is covered.
		return true
	}
	return OperationRelate_RelateOp_Relate(a, b).IsCovers()
}

func geom_GeometryRelate_CoveredBy(a, b *Geom_Geometry) bool {
	if geom_geometryRelate_isRelateNG {
		return OperationRelateng_RelateNG_Relate(a, b, OperationRelateng_RelatePredicate_CoveredBy())
	}
	return geom_GeometryRelate_Covers(b, a)
}

func geom_GeometryRelate_Crosses(a, b *Geom_Geometry) bool {
	if geom_geometryRelate_isRelateNG {
		return OperationRelateng_RelateNG_Relate(a, b, OperationRelateng_RelatePredicate_Crosses())
	}
	// Short-circuit test.
	if !a.GetEnvelopeInternal().IntersectsEnvelope(b.GetEnvelopeInternal()) {
		return false
	}
	return OperationRelate_RelateOp_Relate(a, b).IsCrosses(a.GetDimension(), b.GetDimension())
}

func geom_GeometryRelate_Disjoint(a, b *Geom_Geometry) bool {
	if geom_geometryRelate_isRelateNG {
		return OperationRelateng_RelateNG_Relate(a, b, OperationRelateng_RelatePredicate_Disjoint())
	}
	return !geom_GeometryRelate_Intersects(a, b)
}

func geom_GeometryRelate_EqualsTopo(a, b *Geom_Geometry) bool {
	if geom_geometryRelate_isRelateNG {
		return OperationRelateng_RelateNG_Relate(a, b, OperationRelateng_RelatePredicate_EqualsTopo())
	}
	if !a.GetEnvelopeInternal().Equals(b.GetEnvelopeInternal()) {
		return false
	}
	return OperationRelate_RelateOp_Relate(a, b).IsEquals(a.GetDimension(), b.GetDimension())
}

func geom_GeometryRelate_Overlaps(a, b *Geom_Geometry) bool {
	if geom_geometryRelate_isRelateNG {
		return OperationRelateng_RelateNG_Relate(a, b, OperationRelateng_RelatePredicate_Overlaps())
	}
	if !a.GetEnvelopeInternal().IntersectsEnvelope(b.GetEnvelopeInternal()) {
		return false
	}
	return OperationRelate_RelateOp_Relate(a, b).IsOverlaps(a.GetDimension(), b.GetDimension())
}

func geom_GeometryRelate_Touches(a, b *Geom_Geometry) bool {
	if geom_geometryRelate_isRelateNG {
		return OperationRelateng_RelateNG_Relate(a, b, OperationRelateng_RelatePredicate_Touches())
	}
	if !a.GetEnvelopeInternal().IntersectsEnvelope(b.GetEnvelopeInternal()) {
		return false
	}
	return OperationRelate_RelateOp_Relate(a, b).IsTouches(a.GetDimension(), b.GetDimension())
}

func geom_GeometryRelate_Within(a, b *Geom_Geometry) bool {
	if geom_geometryRelate_isRelateNG {
		return OperationRelateng_RelateNG_Relate(a, b, OperationRelateng_RelatePredicate_Within())
	}
	return geom_GeometryRelate_Contains(b, a)
}

func geom_GeometryRelate_Relate(a, b *Geom_Geometry) *Geom_IntersectionMatrix {
	if geom_geometryRelate_isRelateNG {
		return OperationRelateng_RelateNG_RelateMatrix(a, b)
	}
	Geom_Geometry_CheckNotGeometryCollection(a)
	Geom_Geometry_CheckNotGeometryCollection(b)
	return OperationRelate_RelateOp_Relate(a, b)
}

func geom_GeometryRelate_RelatePattern(a, b *Geom_Geometry, intersectionPattern string) bool {
	if geom_geometryRelate_isRelateNG {
		return OperationRelateng_RelateNG_RelatePattern(a, b, intersectionPattern)
	}
	Geom_Geometry_CheckNotGeometryCollection(a)
	Geom_Geometry_CheckNotGeometryCollection(b)
	return OperationRelate_RelateOp_Relate(a, b).MatchesPattern(intersectionPattern)
}
