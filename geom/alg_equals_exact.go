package geom

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

func Tolerance(within float64) EqualsExactOption {
	return func(s *equalsExactOptionSet) {
		s.toleranceSq = within * within
	}
}

func (os equalsExactOptionSet) eq(a XY, b XY) bool {
	asb := a.Sub(b)
	return asb.Dot(asb) <= os.toleranceSq
}

var IgnoreOrder = EqualsExactOption(
	func(s *equalsExactOptionSet) {
		s.ignoreOrder = true
	},
)

type curve interface {
	NumPoints() int
	PointN(int) Point
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
			pt1 := c1.PointN(m1(i)).XY()
			pt2 := c2.PointN(m2(i)).XY()
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

func isRing(c curve) bool {
	ptA := c.PointN(0)
	ptB := c.PointN(c.NumPoints() - 1)
	return ptA.XY().Equals(ptB.XY())
}
