package jts

import "math"

// IndexStrtree_Interval is a contiguous portion of 1D-space. Used internally by
// SIRtree.
type IndexStrtree_Interval struct {
	min float64
	max float64
}

// IndexStrtree_NewInterval creates a new Interval with the given min and max.
func IndexStrtree_NewInterval(min, max float64) *IndexStrtree_Interval {
	Util_Assert_IsTrue(min <= max)
	return &IndexStrtree_Interval{
		min: min,
		max: max,
	}
}

// IndexStrtree_NewIntervalFromInterval creates a new Interval as a copy of
// another.
func IndexStrtree_NewIntervalFromInterval(other *IndexStrtree_Interval) *IndexStrtree_Interval {
	return IndexStrtree_NewInterval(other.min, other.max)
}

// GetCentre returns the centre of the interval.
func (i *IndexStrtree_Interval) GetCentre() float64 {
	return (i.min + i.max) / 2
}

// ExpandToInclude expands this interval to include the other interval. Returns
// this.
func (i *IndexStrtree_Interval) ExpandToInclude(other *IndexStrtree_Interval) *IndexStrtree_Interval {
	i.max = math.Max(i.max, other.max)
	i.min = math.Min(i.min, other.min)
	return i
}

// Intersects tests whether this interval intersects the other interval.
func (i *IndexStrtree_Interval) Intersects(other *IndexStrtree_Interval) bool {
	return !(other.min > i.max || other.max < i.min)
}

// Equals tests whether this interval is equal to the other.
func (i *IndexStrtree_Interval) Equals(other *IndexStrtree_Interval) bool {
	if other == nil {
		return false
	}
	return i.min == other.min && i.max == other.max
}
