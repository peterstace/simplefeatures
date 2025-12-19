package jts

import "math"

// IndexStrtree_EnvelopeDistance contains functions for computing distances
// between Envelopes.

// IndexStrtree_EnvelopeDistance_MaximumDistance computes the maximum distance
// between the points defining two envelopes. It is equal to the length of the
// diagonal of the envelope containing both input envelopes. This is a coarse
// upper bound on the distance between geometries bounded by the envelopes.
func IndexStrtree_EnvelopeDistance_MaximumDistance(env1, env2 *Geom_Envelope) float64 {
	minx := math.Min(env1.GetMinX(), env2.GetMinX())
	miny := math.Min(env1.GetMinY(), env2.GetMinY())
	maxx := math.Max(env1.GetMaxX(), env2.GetMaxX())
	maxy := math.Max(env1.GetMaxY(), env2.GetMaxY())
	return indexStrtree_EnvelopeDistance_distance(minx, miny, maxx, maxy)
}

func indexStrtree_EnvelopeDistance_distance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Hypot(dx, dy)
}

// IndexStrtree_EnvelopeDistance_MinMaxDistance computes the Min-Max Distance
// between two Envelopes. It is equal to the minimum of the maximum distances
// between all pairs of edge segments from the two envelopes. This is the tight
// upper bound on the distance between geometric items bounded by the envelopes.
//
// Theoretically this bound can be used in the R-tree nearest-neighbour
// branch-and-bound search instead of MaximumDistance. However, little
// performance improvement is observed in practice.
func IndexStrtree_EnvelopeDistance_MinMaxDistance(a, b *Geom_Envelope) float64 {
	aminx := a.GetMinX()
	aminy := a.GetMinY()
	amaxx := a.GetMaxX()
	amaxy := a.GetMaxY()
	bminx := b.GetMinX()
	bminy := b.GetMinY()
	bmaxx := b.GetMaxX()
	bmaxy := b.GetMaxY()

	dist := indexStrtree_EnvelopeDistance_maxDistance(aminx, aminy, aminx, amaxy, bminx, bminy, bminx, bmaxy)
	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(aminx, aminy, aminx, amaxy, bminx, bminy, bmaxx, bminy))
	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(aminx, aminy, aminx, amaxy, bmaxx, bmaxy, bminx, bmaxy))
	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(aminx, aminy, aminx, amaxy, bmaxx, bmaxy, bmaxx, bminy))

	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(aminx, aminy, amaxx, aminy, bminx, bminy, bminx, bmaxy))
	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(aminx, aminy, amaxx, aminy, bminx, bminy, bmaxx, bminy))
	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(aminx, aminy, amaxx, aminy, bmaxx, bmaxy, bminx, bmaxy))
	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(aminx, aminy, amaxx, aminy, bmaxx, bmaxy, bmaxx, bminy))

	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(amaxx, amaxy, aminx, amaxy, bminx, bminy, bminx, bmaxy))
	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(amaxx, amaxy, aminx, amaxy, bminx, bminy, bmaxx, bminy))
	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(amaxx, amaxy, aminx, amaxy, bmaxx, bmaxy, bminx, bmaxy))
	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(amaxx, amaxy, aminx, amaxy, bmaxx, bmaxy, bmaxx, bminy))

	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(amaxx, amaxy, amaxx, aminy, bminx, bminy, bminx, bmaxy))
	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(amaxx, amaxy, amaxx, aminy, bminx, bminy, bmaxx, bminy))
	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(amaxx, amaxy, amaxx, aminy, bmaxx, bmaxy, bminx, bmaxy))
	dist = math.Min(dist, indexStrtree_EnvelopeDistance_maxDistance(amaxx, amaxy, amaxx, aminy, bmaxx, bmaxy, bmaxx, bminy))

	return dist
}

// indexStrtree_EnvelopeDistance_maxDistance computes the maximum distance
// between two line segments.
func indexStrtree_EnvelopeDistance_maxDistance(ax1, ay1, ax2, ay2, bx1, by1, bx2, by2 float64) float64 {
	dist := indexStrtree_EnvelopeDistance_distance(ax1, ay1, bx1, by1)
	dist = math.Max(dist, indexStrtree_EnvelopeDistance_distance(ax1, ay1, bx2, by2))
	dist = math.Max(dist, indexStrtree_EnvelopeDistance_distance(ax2, ay2, bx1, by1))
	dist = math.Max(dist, indexStrtree_EnvelopeDistance_distance(ax2, ay2, bx2, by2))
	return dist
}
