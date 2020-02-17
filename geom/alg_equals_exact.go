package geom

// EqualsExactOption allows the behaviour of the EqualsExact method in the
// Geometry interface to be modified.
type EqualsExactOption func(s *equalsExactOptionSet)

type equalsExactOptionSet struct {
	toleranceSq float64
	ignoreOrder bool
}

func newEqualsExactOptionSet(opts []EqualsExactOption) equalsExactOptionSet {
	var s equalsExactOptionSet
	for _, o := range opts {
		o(&s)
	}
	return s
}

// Tolerance modifies the behaviour of the EqualsExact method by allowing two
// geometry control points be be considered equal if they are within the given
// euclidean distance of each other.
func Tolerance(within float64) EqualsExactOption {
	return func(s *equalsExactOptionSet) {
		s.toleranceSq = within * within
	}
}

func (os equalsExactOptionSet) eq(a XY, b XY) bool {
	asb := a.Sub(b)
	return asb.Dot(asb) <= os.toleranceSq
}

// IgnoreOrder modifies the behaviour of the EqualsExact method by ignoring
// ordering that doesn't have a material impact on geometries.
//
// For Points, there is no ordering, so this option does nothing.
//
// For curves (Line, LineString, and LinearRing), the direction of the curve
// (start to end or end to start) is ignored. For curves that are rings (i.e.
// are simple and closed), the location of the start and end point of the ring
// is also ignored.
//
// For polygons the ordering between any interior rings is ignored, as is the
// ordering inside the rings themselves.
//
// For collections (MultiPoint, MultiLineString, MultiPolygon, and
// GeometryCollection), the ordering of constituent elements in the collection
// are ignored.
var IgnoreOrder = EqualsExactOption(
	func(s *equalsExactOptionSet) {
		s.ignoreOrder = true
	},
)

func ignoreOrder(opts []EqualsExactOption) bool {
	return newEqualsExactOptionSet(opts).ignoreOrder
}

// curve abstracts Line and LineString.
type curve interface {
	NumPoints() int
	PointN(int) Coordinates
}

func curvesExactEqual(c1, c2 curve, opts []EqualsExactOption) bool {
	// Must have the same number of points.
	n := c1.NumPoints()
	if n != c2.NumPoints() {
		return false
	}

	// Allow curves to be compared using a point index mapping, allowing
	// curves to be compared under a rotation or point reversal.
	os := newEqualsExactOptionSet(opts)
	type curveMapping func(int) int
	sameCurve := func(m1, m2 curveMapping) bool {
		for i := 0; i < n; i++ {
			pt1 := c1.PointN(m1(i)).XY
			pt2 := c2.PointN(m2(i)).XY
			if !os.eq(pt1, pt2) {
				return false
			}
		}
		return true
	}

	// First check the regular pointwise comparison. No accounting for
	// reversal or ring offsets.
	identity := func(i int) int { return i }
	if equal := sameCurve(identity, identity); equal || !os.ignoreOrder {
		return equal
	}

	// Next check if one ring is just the reversal of the other.
	reversed := func(i int) int { return n - i - 1 }
	areRings := isRing(c1) && isRing(c2)
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

// TODO: All curves should have an IsRing function. Once that exists, then this
// can be removed.
func isRing(c curve) bool {
	ptA := c.PointN(0)
	ptB := c.PointN(c.NumPoints() - 1)
	return ptA.XY == ptB.XY
}

func multiPointExactEqual(mp1, mp2 MultiPoint, opts []EqualsExactOption) bool {
	n := mp1.NumPoints()
	if mp2.NumPoints() != n {
		return false
	}
	os := newEqualsExactOptionSet(opts)
	ptsEq := func(i, j int) bool {
		xyA, okA := mp1.PointN(i).XY()
		xyB, okB := mp2.PointN(j).XY()
		if okA != okB {
			return false // one empty, but not the other
		}
		if !okA {
			return true // both empty
		}
		return os.eq(xyA, xyB)
	}
	return structureEqual(n, ptsEq, os.ignoreOrder)
}

func polygonExactEqual(p1, p2 Polygon, opts []EqualsExactOption) bool {
	n := p1.NumInteriorRings()
	if n != p2.NumInteriorRings() {
		return false
	}
	if !curvesExactEqual(p1.ExteriorRing(), p2.ExteriorRing(), opts) {
		return false
	}
	ringsEq := func(i, j int) bool {
		ringA := p1.InteriorRingN(i)
		ringB := p2.InteriorRingN(j)
		return curvesExactEqual(ringA, ringB, opts)
	}
	return structureEqual(n, ringsEq, ignoreOrder(opts))
}

func multiLineStringExactEqual(mls1, mls2 MultiLineString, opts []EqualsExactOption) bool {
	n := mls1.NumLineStrings()
	if n != mls2.NumLineStrings() {
		return false
	}
	lsEq := func(i, j int) bool {
		lsA := mls1.LineStringN(i)
		lsB := mls2.LineStringN(j)
		return curvesExactEqual(lsA, lsB, opts)
	}
	return structureEqual(n, lsEq, ignoreOrder(opts))
}

func multiPolygonExactEqual(mp1, mp2 MultiPolygon, opts []EqualsExactOption) bool {
	n := mp1.NumPolygons()
	if n != mp2.NumPolygons() {
		return false
	}
	polyEq := func(i, j int) bool {
		pA := mp1.PolygonN(i)
		pB := mp2.PolygonN(j)
		return polygonExactEqual(pA, pB, opts)
	}
	return structureEqual(n, polyEq, ignoreOrder(opts))
}

func geometryCollectionExactEqual(gc1, gc2 GeometryCollection, opts []EqualsExactOption) bool {
	n := gc1.NumGeometries()
	if n != gc2.NumGeometries() {
		return false
	}
	eq := func(i, j int) bool {
		gA := gc1.GeometryN(i)
		gB := gc2.GeometryN(j)
		return gA.EqualsExact(gB, opts...)
	}
	return structureEqual(n, eq, ignoreOrder(opts))
}

// structureEqual checks if the structure of two geometries each with n sub
// elements are equal. The eq function should check if sub element i from the
// first geometry is equal to sub element j from the second geometry.
func structureEqual(n int, eq func(i, j int) bool, ignoreOrder bool) bool {
	if ignoreOrder {
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
