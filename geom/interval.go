package geom

// Interval represents the interval bound by two float64 values. The interval
// is closed, i.e. its endpoints are included. An interval typically has
// distinct endpoints (i.e. is non-degenerate). It may also be degenerate and
// contain no elements, or degenerate and contain a single element (i.e. the
// min and max bounds are the same). The zero value of Interval is the
// degenerate interval that contains no elements.
type Interval struct {
	min, max float64
	nonEmpty bool
}

// NewInterval returns a new non-empty Interval with the given bounds (which
// may be the same).
func NewInterval(boundA, boundB float64) Interval {
	if boundB < boundA {
		boundA, boundB = boundB, boundA
	}
	return Interval{boundA, boundB, true}
}

// MinMax returns the minimum and maximum bounds of the interval. The boolean
// return value indicates if the interval is non-empty (the minimum and maximum
// bounds should be ignored if false).
func (i Interval) MinMax() (float64, float64, bool) {
	return i.min, i.max, i.nonEmpty
}
