package geom

// Union returns a geometry that represents the parts from either geometry A or
// geometry B (or both). An error may be returned in pathological cases of
// numerical degeneracy.
func Union(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() && b.IsEmpty() {
		return Geometry{}, nil
	}
	if a.IsEmpty() {
		return UnaryUnion(b)
	}
	if b.IsEmpty() {
		return UnaryUnion(a)
	}
	g, err := setOp(a, or, b)
	return g, wrap(err, "executing union")
}

// Intersection returns a geometry that represents the parts that are common to
// both geometry A and geometry B. An error may be returned in pathological
// cases of numerical degeneracy.
func Intersection(a, b Geometry) (Geometry, error) {
	if a.IsEmpty() || b.IsEmpty() {
		return Geometry{}, nil
	}
	g, err := setOp(a, and, b)
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
		return UnaryUnion(a)
	}
	g, err := setOp(a, andNot, b)
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
		return UnaryUnion(b)
	}
	if b.IsEmpty() {
		return UnaryUnion(a)
	}
	g, err := setOp(a, xor, b)
	return g, wrap(err, "executing symmetric difference")
}

// UnaryUnion is a single input variant of the Union function, unioning
// together the components of the input geometry.
func UnaryUnion(g Geometry) (Geometry, error) {
	return setOp(g, or, Geometry{})
}

// UnionMany unions together the input geometries.
func UnionMany(gs []Geometry) (Geometry, error) {
	gc := NewGeometryCollection(gs)
	return UnaryUnion(gc.AsGeometry())
}

func setOp(a Geometry, include func([2]bool) bool, b Geometry) (Geometry, error) {
	overlay := newDCELFromGeometries(a, b)
	g, err := overlay.extractGeometry(include)
	if err != nil {
		return Geometry{}, wrap(err, "internal error extracting geometry")
	}
	if err := g.Validate(); err != nil {
		return Geometry{}, wrap(err, "invalid geometry produced by overlay")
	}
	return g, nil
}

func or(b [2]bool) bool     { return b[0] || b[1] }
func and(b [2]bool) bool    { return b[0] && b[1] }
func xor(b [2]bool) bool    { return b[0] != b[1] }
func andNot(b [2]bool) bool { return b[0] && !b[1] }
