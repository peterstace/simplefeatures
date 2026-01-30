package geom

import "github.com/peterstace/simplefeatures/internal/jtsport/jts"

// hasGC determines if either argument is a GeometryCollection. The JTS port
// doesn't have full support for GeometryCollections, so we need to handle them
// specially.
//
// Specifically, the JTS port has the following restrictions for binary overlay
// operations:
//
// 1. All input elements must have the same dimension.
//
// 2. Polygons in a GeometryCollection must not overlap.
func hasGC(a, b Geometry) bool {
	return a.IsGeometryCollection() || b.IsGeometryCollection()
}

// Union returns a geometry that represents the parts from either geometry A or
// geometry B (or both). An error may be returned in pathological cases of
// numerical degeneracy.
func Union(a, b Geometry) (Geometry, error) {
	if hasGC(a, b) {
		return gcAwareUnion(a, b)
	}
	return jtsOverlayOp(a, b, jts.OperationOverlayng_OverlayNG_UNION)
}

func gcAwareUnion(a, b Geometry) (Geometry, error) {
	// UnaryUnion supports arbitrary GeometryCollections.
	gc := NewGeometryCollection([]Geometry{a, b})
	return UnaryUnion(gc.AsGeometry())
}

// Intersection returns a geometry that represents the parts that are common to
// both geometry A and geometry B. An error may be returned in pathological
// cases of numerical degeneracy.
func Intersection(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() || b.IsEmpty() {
		return Geometry{}, nil
	}
	if hasGC(a, b) {
		return gcAwareIntersection(a, b)
	}
	return jtsOverlayOp(a, b, jts.OperationOverlayng_OverlayNG_INTERSECTION)
}

func gcAwareIntersection(a, b Geometry) (Geometry, error) {
	partsA, partsB, err := prepareOverlayInputParts(a, b)
	if err != nil {
		return Geometry{}, err
	}

	// The total result is the union of the intersections across the Cartesian
	// product of parts.
	var results []Geometry
	for _, partA := range partsA {
		for _, partB := range partsB {
			result, err := jtsOverlayOp(partA, partB, jts.OperationOverlayng_OverlayNG_INTERSECTION)
			if err != nil {
				return Geometry{}, err
			}
			results = append(results, result)
		}
	}
	return UnaryUnion(NewGeometryCollection(results).AsGeometry())
}

func explodeGeometryCollections(dst []Geometry, g Geometry) []Geometry {
	if gc, ok := g.AsGeometryCollection(); ok {
		for i := 0; i < gc.NumGeometries(); i++ {
			dst = explodeGeometryCollections(dst, gc.GeometryN(i))
		}
		return dst
	}
	return append(dst, g)
}

func prepareOverlayInputParts(a, b Geometry) ([]Geometry, []Geometry, error) {
	// Normalize GC inputs by unioning their parts.
	if a.IsGeometryCollection() {
		var err error
		a, err = UnaryUnion(a)
		if err != nil {
			return nil, nil, err
		}
	}
	if b.IsGeometryCollection() {
		var err error
		b, err = UnaryUnion(b)
		if err != nil {
			return nil, nil, err
		}
	}

	// Extract non-GC parts from each input.
	partsA := explodeGeometryCollections(nil, a)
	partsB := explodeGeometryCollections(nil, b)
	return partsA, partsB, nil
}

// Difference returns a geometry that represents the parts of input geometry A
// that are not part of input geometry B. An error may be returned in cases of
// pathological cases of numerical degeneracy.
func Difference(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() {
		return Geometry{}, nil
	}
	if hasGC(a, b) {
		return gcAwareDifference(a, b)
	}
	return jtsOverlayOp(a, b, jts.OperationOverlayng_OverlayNG_DIFFERENCE)
}

func gcAwareDifference(a, b Geometry) (Geometry, error) {
	partsA, partsB, err := prepareOverlayInputParts(a, b)
	if err != nil {
		return Geometry{}, err
	}

	// The total result is the union of each part of A after each part of B has
	// been removed (sequentially).
	var results []Geometry
	for _, partA := range partsA {
		result := partA
		for _, partB := range partsB {
			var err error
			result, err = jtsOverlayOp(result, partB, jts.OperationOverlayng_OverlayNG_DIFFERENCE)
			if err != nil {
				return Geometry{}, err
			}
			if result.IsEmpty() {
				break
			}
		}
		results = append(results, result)
	}
	return UnaryUnion(NewGeometryCollection(results).AsGeometry())
}

// jtsOverlayOp invokes the JTS port's overlay operation with the given opCode.
func jtsOverlayOp(a, b Geometry, opCode int) (Geometry, error) {
	var result Geometry
	err := catch(func() error {
		wkbReader := jts.Io_NewWKBReader()
		jtsA, err := wkbReader.ReadBytes(a.AsBinary())
		if err != nil {
			return wrap(err, "converting geometry A to JTS")
		}
		jtsB, err := wkbReader.ReadBytes(b.AsBinary())
		if err != nil {
			return wrap(err, "converting geometry B to JTS")
		}
		jtsResult := jts.OperationOverlayng_OverlayNGRobust_Overlay(jtsA, jtsB, opCode)
		wkbWriter := jts.Io_NewWKBWriter()
		result, err = UnmarshalWKB(wkbWriter.Write(jtsResult), NoValidate{})
		return wrap(err, "converting JTS overlay result to simplefeatures")
	})
	return result, err
}

// SymmetricDifference returns a geometry that represents the parts of geometry
// A and B that are not in common. An error may be returned in pathological
// cases of numerical degeneracy.
func SymmetricDifference(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() && b.IsEmpty() {
		return Geometry{}, nil
	}
	if a.IsEmpty() {
		return UnaryUnion(b)
	}
	if b.IsEmpty() {
		return UnaryUnion(a)
	}

	if hasGC(a, b) {
		return gcAwareSymmetricDifference(a, b)
	}
	return jtsOverlayOp(a, b, jts.OperationOverlayng_OverlayNG_SYMDIFFERENCE)
}

func gcAwareSymmetricDifference(a, b Geometry) (Geometry, error) {
	diffAB, err := Difference(a, b)
	if err != nil {
		return Geometry{}, err
	}
	diffBA, err := Difference(b, a)
	if err != nil {
		return Geometry{}, err
	}
	return Union(diffAB, diffBA)
}

// UnaryUnion is a single input variant of the Union function, unioning
// together the components of the input geometry.
func UnaryUnion(g Geometry) (Geometry, error) {
	if g.IsEmpty() {
		return Geometry{}, nil
	}
	return jtsUnaryUnion(g)
}

// UnionMany unions together the input geometries.
func UnionMany(gs []Geometry) (Geometry, error) {
	gc := NewGeometryCollection(gs)
	return UnaryUnion(gc.AsGeometry())
}

// jtsUnaryUnion invokes the JTS port's unary union operation.
func jtsUnaryUnion(g Geometry) (Geometry, error) {
	var result Geometry
	err := catch(func() error {
		wkbReader := jts.Io_NewWKBReader()
		jtsG, err := wkbReader.ReadBytes(g.AsBinary())
		if err != nil {
			return wrap(err, "converting geometry to JTS")
		}
		jtsResult := jts.OperationOverlayng_OverlayNGRobust_Union(jtsG)
		wkbWriter := jts.Io_NewWKBWriter()
		result, err = UnmarshalWKB(wkbWriter.Write(jtsResult), NoValidate{})
		return wrap(err, "converting JTS union result to simplefeatures")
	})
	return result, err
}
