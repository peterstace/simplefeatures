package geom

// ExactEqualsOption allows the behaviour of the ExactEquals method in the
// Geometry interface to be modified.
type ExactEqualsOption func(exactEqualsComparator) exactEqualsComparator

type exactEqualsComparator struct {
	toleranceSq float64
	ignoreOrder bool
}

func newExactEqualsComparator(opts []ExactEqualsOption) exactEqualsComparator {
	var c exactEqualsComparator
	for _, o := range opts {
		c = o(c)
	}
	return c
}

// ToleranceXY modifies the behaviour of the ExactEquals method by allowing two
// geometry control points be considered equal if their XY coordinates are
// within the given euclidean distance of each other.
func ToleranceXY(within float64) ExactEqualsOption {
	return func(c exactEqualsComparator) exactEqualsComparator {
		c.toleranceSq = within * within
		return c
	}
}

func (c exactEqualsComparator) eq(a, b Coordinates) bool {
	if a.Type != b.Type {
		return false
	}
	asb := a.XY.Sub(b.XY)
	if asb.lengthSq() > c.toleranceSq {
		return false
	}
	if a.Type.Is3D() && a.Z != b.Z {
		return false
	}
	if a.Type.IsMeasured() && a.M != b.M {
		return false
	}
	return true
}

// IgnoreOrder is an ExactEqualsOption that modifies the behaviour of the
// ExactEquals method by ignoring ordering that doesn't have a material impact
// on geometries.
//
// For Points, there is no ordering, so this option does nothing.
//
// For LineStrings, the direction of the curve (start to end or end to start)
// is ignored. If the LineStrings are rings, (i.e. are simple and closed), the
// location of the start and end point of the ring is also ignored.
//
// For polygons the ordering between any interior rings is ignored, as is the
// ordering inside the rings themselves.
//
// For collections (MultiPoint, MultiLineString, MultiPolygon, and
// GeometryCollection), the ordering of constituent elements in the collection
// are ignored.
func IgnoreOrder(c exactEqualsComparator) exactEqualsComparator { //nolint:revive
	c.ignoreOrder = true
	return c
}

// ExactEquals checks if two geometries are equal from a structural pointwise
// equality perspective. Geometries that are structurally equal are defined by
// exactly same control points in the same order. Note that even if two
// geometries are spatially equal (i.e.  represent the same point set), they
// may not be defined by exactly the same way. Ordering differences and numeric
// tolerances can be accounted for using options.
func ExactEquals(g1, g2 Geometry, opts ...ExactEqualsOption) bool {
	c := newExactEqualsComparator(opts)
	return c.geometriesEq(g1, g2)
}

func (c exactEqualsComparator) geometriesEq(g1, g2 Geometry) bool {
	if g1.Type() != g2.Type() {
		return false
	}
	switch typ := g1.Type(); typ {
	case TypePoint:
		return c.pointsEq(g1.MustAsPoint(), g2.MustAsPoint())
	case TypeMultiPoint:
		return c.multiPointsEq(g1.MustAsMultiPoint(), g2.MustAsMultiPoint())
	case TypeLineString:
		return c.lineStringsEq(g1.MustAsLineString(), g2.MustAsLineString())
	case TypeMultiLineString:
		return c.multiLineStringsEq(g1.MustAsMultiLineString(), g2.MustAsMultiLineString())
	case TypePolygon:
		return c.polygonsEq(g1.MustAsPolygon(), g2.MustAsPolygon())
	case TypeMultiPolygon:
		return c.multiPolygonsEq(g1.MustAsMultiPolygon(), g2.MustAsMultiPolygon())
	case TypeGeometryCollection:
		return c.geometryCollectionsEq(g1.MustAsGeometryCollection(), g2.MustAsGeometryCollection())
	default:
		panic("unknown geometry type: " + typ.String())
	}
}

func (c exactEqualsComparator) lineStringsEq(ls1, ls2 LineString) bool {
	c1 := ls1.Coordinates()
	c2 := ls2.Coordinates()

	// Must have the same number of points and be of the same coordinate type.
	n := c1.Length()
	if n != c2.Length() {
		return false
	}
	if c1.CoordinatesType() != c2.CoordinatesType() {
		return false
	}

	// Allow curves to be compared using a point index mapping, allowing
	// curves to be compared under a rotation or point reversal.
	type curveMapping func(int) int
	sameCurve := func(m1, m2 curveMapping) bool {
		for i := 0; i < n; i++ {
			c1 := c1.Get(m1(i))
			c2 := c2.Get(m2(i))
			if !c.eq(c1, c2) {
				return false
			}
		}
		return true
	}

	// First check the regular pointwise comparison. No accounting for
	// reversal or ring offsets.
	identity := func(i int) int { return i }
	if equal := sameCurve(identity, identity); equal || !c.ignoreOrder {
		return equal
	}

	// Next check if one ring is just the reversal of the other.
	reversed := func(i int) int { return n - i - 1 }
	areRings := ls1.IsRing() && ls2.IsRing()
	if revEq := sameCurve(identity, reversed); revEq || !areRings {
		return revEq
	}

	// Finally, check if the rings are the same once rotated.
	for o := 1; o < n; o++ {
		offset := func(i int) int {
			return (i + o) % (n - 1)
		}
		if sameCurve(identity, offset) || sameCurve(reversed, offset) {
			return true
		}
	}
	return false
}

func (c exactEqualsComparator) pointsEq(p1, p2 Point) bool {
	if p1.IsEmpty() && p2.IsEmpty() {
		return p1.CoordinatesType() == p2.CoordinatesType()
	}
	c1, ok1 := p1.Coordinates()
	c2, ok2 := p2.Coordinates()
	return ok1 && ok2 && c.eq(c1, c2)
}

func (c exactEqualsComparator) multiPointsEq(mp1, mp2 MultiPoint) bool {
	n := mp1.NumPoints()
	if mp2.NumPoints() != n {
		return false
	}
	if mp1.CoordinatesType() != mp2.CoordinatesType() {
		return false
	}
	ptsEq := func(i, j int) bool {
		cA, okA := mp1.PointN(i).Coordinates()
		cB, okB := mp2.PointN(j).Coordinates()
		if okA != okB {
			return false // one empty, but not the other
		}
		if !okA {
			return true // both empty
		}
		return c.eq(cA, cB)
	}
	return c.structureEq(n, ptsEq)
}

func (c exactEqualsComparator) polygonsEq(p1, p2 Polygon) bool {
	n := p1.NumInteriorRings()
	if n != p2.NumInteriorRings() {
		return false
	}
	if !c.lineStringsEq(p1.ExteriorRing(), p2.ExteriorRing()) {
		return false
	}
	ringsEq := func(i, j int) bool {
		ringA := p1.InteriorRingN(i)
		ringB := p2.InteriorRingN(j)
		return c.lineStringsEq(ringA, ringB)
	}
	return c.structureEq(n, ringsEq)
}

func (c exactEqualsComparator) multiLineStringsEq(mls1, mls2 MultiLineString) bool {
	n := mls1.NumLineStrings()
	if n != mls2.NumLineStrings() {
		return false
	}
	if mls1.CoordinatesType() != mls2.CoordinatesType() {
		return false
	}
	lsEq := func(i, j int) bool {
		lsA := mls1.LineStringN(i)
		lsB := mls2.LineStringN(j)
		return c.lineStringsEq(lsA, lsB)
	}
	return c.structureEq(n, lsEq)
}

func (c exactEqualsComparator) multiPolygonsEq(mp1, mp2 MultiPolygon) bool {
	n := mp1.NumPolygons()
	if n != mp2.NumPolygons() {
		return false
	}
	if mp1.CoordinatesType() != mp2.CoordinatesType() {
		return false
	}
	polyEq := func(i, j int) bool {
		pA := mp1.PolygonN(i)
		pB := mp2.PolygonN(j)
		return c.polygonsEq(pA, pB)
	}
	return c.structureEq(n, polyEq)
}

func (c exactEqualsComparator) geometryCollectionsEq(gc1, gc2 GeometryCollection) bool {
	n := gc1.NumGeometries()
	if n != gc2.NumGeometries() {
		return false
	}
	if gc1.CoordinatesType() != gc2.CoordinatesType() {
		return false
	}
	eq := func(i, j int) bool {
		gA := gc1.GeometryN(i)
		gB := gc2.GeometryN(j)
		return c.geometriesEq(gA, gB)
	}
	return c.structureEq(n, eq)
}

// structureEq checks if the structure of two geometries each with n sub
// elements are equal. The eq function should check if sub element i from the
// first geometry is equal to sub element j from the second geometry.
func (c exactEqualsComparator) structureEq(n int, eq func(i, j int) bool) bool {
	if c.ignoreOrder {
		return validPermutation(n, eq)
	}
	for i := 0; i < n; i++ {
		if !eq(i, i) {
			return false
		}
	}
	return true
}

// validPermutation tests if there is a permutation of 0, 1, ... n-1 such that
// eq is always true pairwise across permuted and unpermuted values.
func validPermutation(n int, eq func(i, j int) bool) bool {
	choices := make([]int, n)
	for i := 0; i < n; i++ {
		choices[i] = i
	}

	var recurse func(int) bool
	recurse = func(level int) bool {
		if len(choices) == 0 {
			return true
		}
		for i, c := range choices {
			if !eq(level, c) {
				continue
			}

			// Recurse using all choices _except_ for c by swapping c with the
			// last choice, shortening the slice by 1, then recursing. The
			// original choices state is restored after recursing.
			choices[i], choices[len(choices)-1] = choices[len(choices)-1], choices[i]
			choices = choices[:len(choices)-1]
			if recurse(level + 1) {
				return true
			}
			choices = choices[:len(choices)+1]
			choices[i], choices[len(choices)-1] = choices[len(choices)-1], choices[i]
		}
		return false
	}
	return recurse(0)
}
