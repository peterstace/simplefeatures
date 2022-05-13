package geom

// Union returns a geometry that represents the parts from either geometry A or
// geometry B (or both). An error may be returned in pathological cases of
// numerical degeneracy.
func Union(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() && b.IsEmpty() {
		return Geometry{}, nil
	}
	if a.IsEmpty() {
		return b, nil
	}
	if b.IsEmpty() {
		return a, nil
	}
	g, err := setOp(a, b, selectUnion, false)
	return g, wrap(err, "executing union")
}

// Intersection returns a geometry that represents the parts that are common to
// both geometry A and geometry B. An error may be returned in pathological
// cases of numerical degeneracy.
func Intersection(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() || b.IsEmpty() {
		return Geometry{}, nil
	}
	g, err := setOp(a, b, selectIntersection, true)
	return g, wrap(err, "executing intersection")
}

// Difference returns a geometry that represents the parts of input geometry A
// that are not part of input geometry B. An error may be returned in cases of
// pathological cases of numerical degeneracy.
func Difference(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() {
		return Geometry{}, nil
	}
	if b.IsEmpty() {
		return a, nil
	}
	g, err := setOp(a, b, selectDifference, true)
	return g, wrap(err, "executing difference")
}

// SymmetricDifference returns a geometry that represents the parts of geometry
// A and B that are not in common. An error may be returned in pathological
// cases of numerical degeneracy.
func SymmetricDifference(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() && b.IsEmpty() {
		return Geometry{}, nil
	}
	if a.IsEmpty() {
		return b, nil
	}
	if b.IsEmpty() {
		return a, nil
	}
	g, err := setOp(a, b, selectSymmetricDifference, true)
	return g, wrap(err, "executing symmetric difference")
}

func setOp(a, b Geometry, include func([2]label) bool, mergeFirst bool) (Geometry, error) {
	if mergeFirst {
		// If a or b are GeometryCollections, take the union of their elements.
		// This is a workaround for an issue where if a GeometryCollection has
		// elements that overlap, some set operations don't behave well.
		var err error
		if gc, ok := a.AsGeometryCollection(); ok {
			if a, err = merge(gc); err != nil {
				return Geometry{}, wrap(err, "error merging GeometryCollection")
			}
		}
		if gc, ok := b.AsGeometryCollection(); ok {
			if b, err = merge(gc); err != nil {
				return Geometry{}, wrap(err, "error merging GeometryCollection")
			}
		}
	}
	overlay, err := createOverlay(a, b)
	if err != nil {
		return Geometry{}, wrap(err, "internal error creating overlay")
	}

	g, err := overlay.extractGeometry(include)
	if err != nil {
		return Geometry{}, wrap(err, "internal error extracting geometry")
	}
	return g, nil
}

func merge(gc GeometryCollection) (Geometry, error) {
	merged := Geometry{}
	for _, g := range gc.geoms {
		var err error
		merged, err = Union(merged, g)
		if err != nil {
			return Geometry{}, err
		}
	}
	return merged, nil
}
