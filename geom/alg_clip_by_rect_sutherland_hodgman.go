package geom

import (
	"cmp"
	"fmt"
	"slices"
)

// polygonClipper holds precomputed values for clipping a polygon against an
// axis-aligned rectangle.
type polygonClipper struct {
	lo, hi  XY
	w, h    float64
	perim   float64
	corners [4]float64 // CCW corner parameters: BL, BR, TR, TL
	ctype   CoordinatesType
	s       int // stride: floats per point (2 for XY, 3 for XYZ/XYM, 4 for XYZM)
}

func newPolygonClipper(lo, hi XY, ctype CoordinatesType) polygonClipper {
	w := hi.X - lo.X
	h := hi.Y - lo.Y
	return polygonClipper{
		lo:      lo,
		hi:      hi,
		w:       w,
		h:       h,
		perim:   2*w + 2*h,
		corners: [4]float64{0, w, w + h, 2*w + h},
		ctype:   ctype,
		s:       ctype.Dimension(),
	}
}

func clipPolygonByRect(p Polygon, rect Envelope) Geometry {
	ctype := p.CoordinatesType()
	emptyPoly := NewPolygon(nil).ForceCoordinatesType(ctype).AsGeometry()

	if p.IsEmpty() {
		return emptyPoly
	}

	// Degenerate rect: point or line envelope → empty polygon.
	if !rect.IsRectangle() {
		return emptyPoly
	}

	lo, hi, ok := rect.MinMaxXYs()
	if !ok {
		return emptyPoly
	}

	pEnv := p.Envelope()

	// Fast path: envelopes disjoint.
	if !rect.Intersects(pEnv) {
		return emptyPoly
	}

	// Fast path: polygon entirely inside rect.
	if rect.Covers(pEnv) {
		return p.AsGeometry()
	}

	// Normalise to CCW exterior / CW holes. This ensures all clipped rings
	// have known winding, which the topology resolution depends on.
	p = p.ForceCCW()

	c := newPolygonClipper(lo, hi, ctype)

	// Clip exterior ring.
	clippedExt := c.clipRingSH(p.ExteriorRing())
	if len(clippedExt) == 0 {
		return emptyPoly
	}

	// Classify holes.
	var freeHoles []LineString
	var clippedHoles [][]float64
	for i := 0; i < p.NumInteriorRings(); i++ {
		hole := p.InteriorRingN(i)
		holeEnv := hole.Envelope()

		if !rect.Intersects(holeEnv) {
			// Hole entirely outside rect → discard.
			continue
		}

		// A hole is "free" (entirely inside the rect with no boundary
		// contact) only if its envelope is strictly inside the rect. If any
		// vertex lies on the rect boundary, it must be clipped to avoid
		// producing invalid polygons with shared edges.
		if rect.Covers(holeEnv) && !c.ringTouchesRectBoundary(hole) {
			freeHoles = append(freeHoles, hole)
			continue
		}

		// Hole touches or crosses rect boundary — clip it.
		clippedHole := c.clipRingSH(hole)
		if len(clippedHole) > 0 {
			clippedHoles = append(clippedHoles, clippedHole)
		}
	}

	return c.resolveClippedPolygon(clippedExt, freeHoles, clippedHoles)
}

// ringTouchesRectBoundary reports whether any vertex of the ring lies on the rect
// boundary.
func (c *polygonClipper) ringTouchesRectBoundary(ring LineString) bool {
	seq := ring.Coordinates()
	for i := 0; i < seq.Length(); i++ {
		xy := seq.GetXY(i)
		if xy.X == c.lo.X || xy.X == c.hi.X || xy.Y == c.lo.Y || xy.Y == c.hi.Y {
			return true
		}
	}
	return false
}

// resolveClippedPolygon takes the clipped exterior ring, classified holes, and
// produces the output Geometry (Polygon or MultiPolygon). The input polygon
// must have been normalised to CCW (exterior CCW, holes CW) before clipping,
// so that all clipped rings have known winding.
func (c *polygonClipper) resolveClippedPolygon(
	clippedExterior []float64,
	freeHoles []LineString,
	clippedHoles [][]float64,
) Geometry {
	extArcs := c.extractInteriorArcs(clippedExterior)

	// Collect all arcs.
	var allArcs []interiorArc
	allArcs = append(allArcs, extArcs...)

	emptyPoly := NewPolygon(nil).ForceCoordinatesType(c.ctype).AsGeometry()
	for _, hole := range clippedHoles {
		arcs := c.extractInteriorArcs(hole)
		if len(arcs) == 0 {
			// Clipped hole is entirely on the rect boundary — it covers
			// the entire rect. No polygon area survives.
			return emptyPoly
		}
		allArcs = append(allArcs, arcs...)
	}

	var outputRings [][]float64
	if len(allArcs) == 0 {
		// No interior arcs: the polygon contains the entire rect.
		// Use the clipped exterior directly — it is the rect with
		// correctly interpolated Z/M values from S-H clipping.
		outputRings = append(outputRings, clippedExterior)
	} else {
		outputRings = c.walkArcs(allArcs)
	}

	// Build output polygons. Each output ring is an exterior (already closed).
	var polys []Polygon
	for _, ring := range outputRings {
		seq := NewSequence(ring, c.ctype)
		exterior := NewLineString(seq)
		polys = append(polys, NewPolygon([]LineString{exterior}))
	}

	// Assign free holes to the correct exterior ring.
	for _, hole := range freeHoles {
		holeXY, ok := hole.StartPoint().XY()
		if !ok {
			continue
		}
		for i, ring := range outputRings {
			if c.pointInRing(holeXY, ring) {
				existingRings := polys[i].DumpRings()
				existingRings = append(existingRings, hole)
				polys[i] = NewPolygon(existingRings)
				break
			}
		}
	}

	// Return Polygon or MultiPolygon.
	if len(polys) == 0 {
		return NewPolygon(nil).ForceCoordinatesType(c.ctype).AsGeometry()
	}
	if len(polys) == 1 {
		return polys[0].AsGeometry()
	}
	return NewMultiPolygon(polys).AsGeometry()
}

// walkArcs performs the topology resolution walk, producing output rings from
// the collected interior arcs. It pairs arc endpoints along the rect boundary
// (CCW) and traces complete rings.
func (c *polygonClipper) walkArcs(arcs []interiorArc) [][]float64 {
	// Build a sorted list of arc "end" events and a map from "end param" to
	// the index of the arc that ends there. Also build a map from
	// "start param" to arc index.
	type endpoint struct {
		param float64
		idx   int
	}
	var endPoints []endpoint // sorted by param
	startByParam := make(map[float64]int)
	for i, a := range arcs {
		endPoints = append(endPoints, endpoint{a.endParam, i})
		startByParam[a.startParam] = i
	}

	// Sort endpoints by param.
	slices.SortFunc(endPoints, func(a, b endpoint) int {
		return cmp.Compare(a.param, b.param)
	})

	// Collect all start params sorted.
	var startParams []float64
	for p := range startByParam {
		startParams = append(startParams, p)
	}
	slices.Sort(startParams)

	// For a given end param, find the next start param going CCW.
	findNextStart := func(endParam float64) (float64, int) {
		// Find the first start param that is strictly after endParam (CCW).
		best := -1.0
		bestDist := c.perim + 1
		for _, sp := range startParams {
			d := c.ccwDist(endParam, sp)
			if d > 0 && d < bestDist {
				bestDist = d
				best = sp
			}
		}
		if best < 0 {
			// TODO: Investigate whether this fallback is reachable and
			// whether it produces correct results.
			//
			// This branch is reached when every start param has ccwDist
			// == 0 from endParam — i.e., every arc's start is at the
			// exact same boundary parameter as the current arc's end.
			// The d > 0 filter above rejected them all.
			//
			// For valid clipped polygons this shouldn't happen: an
			// arc's end and the next arc's start should be at distinct
			// boundary positions (separated by at least one boundary
			// edge). Two arcs sharing the same boundary parameter would
			// mean two ring transitions at the exact same point, which
			// would imply a degenerate or self-touching input.
			//
			// If this is truly unreachable, it should be replaced with
			// a panic. If it is reachable, the behaviour of picking an
			// arbitrary arc at the same parameter needs validation —
			// it's unclear whether the walking algorithm produces
			// correct output in this case.
			for _, sp := range startParams {
				if sp == endParam {
					return sp, startByParam[sp]
				}
			}
		}
		return best, startByParam[best]
	}

	used := make([]bool, len(arcs))
	var rings [][]float64

	for {
		// Find first unused arc.
		firstIdx := -1
		for i := range arcs {
			if !used[i] {
				firstIdx = i
				break
			}
		}
		if firstIdx < 0 {
			break
		}

		var ring []float64
		curIdx := firstIdx

		for {
			used[curIdx] = true
			a := arcs[curIdx]

			// Append arc coordinates.
			ring = append(ring, a.coords...)

			// Find next arc via boundary.
			nextStartParam, nextIdx := findNextStart(a.endParam)

			// Build boundary path from this arc's end to the next arc's start.
			endPt := ring[len(ring)-c.s:]
			startPt := arcs[nextIdx].coords[:c.s]

			ring = c.appendBoundaryPath(ring, a.endParam, nextStartParam, endPt, startPt)

			if nextIdx == firstIdx {
				break // Ring complete.
			}
			curIdx = nextIdx
		}

		// Close the ring.
		ring = append(ring, ring[:c.s]...)
		ring = c.removeDupConsecutive(ring)
		if len(ring) >= 4*c.s {
			rings = append(rings, ring)
		}
	}

	return rings
}

// appendBoundaryPath appends coordinates along the rect boundary going CCW
// from startParam to endParam. It includes rect corners between the two
// parameters but does NOT include the start or end points themselves (those are
// the last/first coordinates of the adjacent arcs). Z and M values at corners
// are linearly interpolated between startPt and endPt based on boundary
// distance.
func (c *polygonClipper) appendBoundaryPath(dst []float64, startParam, endParam float64, startPt, endPt []float64) []float64 {
	// Find the first corner after startParam going CCW. The corners are in
	// ascending parameter order, so this is the first with cp > startParam.
	// If startParam is past the last corner (on the left edge), no corner
	// qualifies and firstIdx stays at 0 — the bottom-left corner is the
	// next one going CCW.
	firstIdx := 0
	for i, cp := range c.corners {
		if cp > startParam {
			firstIdx = i
			break
		}
	}

	// Iterate corners in CCW order from firstIdx, collecting those strictly
	// between startParam and endParam.
	totalDist := c.ccwDist(startParam, endParam)
	for k := 0; k < 4; k++ {
		cp := c.corners[(firstIdx+k)%4]
		d := c.ccwDist(startParam, cp)
		if d >= totalDist {
			break
		}
		frac := d / totalDist
		xy := c.paramToXY(cp)
		dst = append(dst, xy.X, xy.Y)
		for dim := 2; dim < c.s; dim++ {
			dst = append(dst, lerp(startPt[dim], endPt[dim], frac))
		}
	}
	return dst
}

// extractInteriorArcs decomposes a clipped ring into interior arcs. The ring
// must be explicitly closed (first == last). Each arc starts and ends at a
// point on the rect boundary. Returns an empty slice if the ring has no
// interior arcs (entirely on the boundary or entirely interior with no
// boundary contact).
func (c *polygonClipper) extractInteriorArcs(ring []float64) []interiorArc {
	s := c.s
	n := len(ring) / s
	if n < 4 {
		return nil
	}

	// There are n-1 edges in a closed ring (the last vertex == first vertex,
	// so edge n-1 from ring[n-1] to ring[0] is degenerate and skipped).
	numEdges := n - 1

	// Classify each edge as boundary (both endpoints on same rect edge) or interior.
	isBdry := make([]bool, numEdges)
	for i := 0; i < numEdges; i++ {
		ax, ay := ring[s*i], ring[s*i+1]
		bx, by := ring[s*(i+1)], ring[s*(i+1)+1]
		isBdry[i] = c.isSameRectEdge(XY{ax, ay}, XY{bx, by})
	}

	// Find a starting boundary edge so we can walk from there. If there are
	// no boundary edges, the entire ring is interior (no boundary contact).
	start := -1
	for i := 0; i < numEdges; i++ {
		if isBdry[i] {
			start = i
			break
		}
	}
	if start < 0 {
		return nil
	}

	// Walk around the ring starting from a boundary edge.
	var arcs []interiorArc
	i := start
	for steps := 0; steps < numEdges; {
		// Skip boundary edges.
		for steps < numEdges && isBdry[i%numEdges] {
			i++
			steps++
		}
		if steps >= numEdges {
			break
		}
		// Start of an interior arc at ring[i%numEdges].
		arcStart := (i % numEdges) * s
		var arcCoords []float64
		arcCoords = append(arcCoords, ring[arcStart:arcStart+s]...)
		i++
		steps++
		for steps < numEdges && !isBdry[i%numEdges] {
			off := (i % numEdges) * s
			arcCoords = append(arcCoords, ring[off:off+s]...)
			i++
			steps++
		}
		if steps < numEdges {
			// The arc ends at ring[i%numEdges] (the start of the next boundary edge).
			off := (i % numEdges) * s
			arcCoords = append(arcCoords, ring[off:off+s]...)
		} else {
			// Wrapped around; the arc ends at ring[start] (where we began).
			arcCoords = append(arcCoords, ring[start*s:start*s+s]...)
		}

		sp := c.rectBoundaryParam(XY{arcCoords[0], arcCoords[1]})
		ep := c.rectBoundaryParam(XY{arcCoords[len(arcCoords)-s], arcCoords[len(arcCoords)-s+1]})
		arcs = append(arcs, interiorArc{
			coords:     arcCoords,
			startParam: sp,
			endParam:   ep,
		})
	}
	return arcs
}

// interiorArc represents a portion of a clipped ring that passes through the
// interior of the clipping rectangle (not along its boundary). The first and
// last s floats (where s is the stride) are on the rectangle boundary;
// everything in between is in the interior.
type interiorArc struct {
	coords     []float64
	startParam float64 // boundary parameter of first point
	endParam   float64 // boundary parameter of last point
}

// clipRingSH clips a closed ring against an axis-aligned rectangle using the
// Sutherland-Hodgman algorithm. It returns the clipped ring as a closed
// []float64 (first point == last point), or nil if the ring is entirely
// outside the rectangle. The input [LineString] must be a closed ring.
func (c *polygonClipper) clipRingSH(ring LineString) []float64 {
	seq := ring.Coordinates()
	coords := seq.appendAllPoints(nil)
	if len(coords)/c.s < 4 {
		return nil
	}

	// Clip against each of the 4 edges.
	coords = c.clipToEdge(coords,
		func(x, y float64) bool { return x >= c.lo.X },
		func(dst, src []float64, ai, bi int) []float64 { return c.appendInterpX(dst, src, ai, bi, c.lo.X) })
	coords = c.clipToEdge(coords,
		func(x, y float64) bool { return x <= c.hi.X },
		func(dst, src []float64, ai, bi int) []float64 { return c.appendInterpX(dst, src, ai, bi, c.hi.X) })
	coords = c.clipToEdge(coords,
		func(x, y float64) bool { return y >= c.lo.Y },
		func(dst, src []float64, ai, bi int) []float64 { return c.appendInterpY(dst, src, ai, bi, c.lo.Y) })
	coords = c.clipToEdge(coords,
		func(x, y float64) bool { return y <= c.hi.Y },
		func(dst, src []float64, ai, bi int) []float64 { return c.appendInterpY(dst, src, ai, bi, c.hi.Y) })

	coords = c.removeDupConsecutive(coords)
	if len(coords)/c.s < 4 {
		return nil
	}
	return coords
}

// clipToEdge performs one pass of the Sutherland-Hodgman algorithm, clipping a
// closed ring against a single half-plane defined by isInside. The
// appendIntersection function appends the intersection point between points at
// indices ai and bi to dst. The output is also an explicitly closed ring.
func (c *polygonClipper) clipToEdge(
	coords []float64,
	isInside func(x, y float64) bool,
	appendIntersection func(dst, coords []float64, ai, bi int) []float64,
) []float64 {
	s := c.s
	n := len(coords) / s
	if n == 0 {
		return nil
	}
	var out []float64
	ai := n - 1
	for bi := 0; bi < n; bi++ {
		ax, ay := coords[s*ai], coords[s*ai+1]
		bx, by := coords[s*bi], coords[s*bi+1]
		aIn := isInside(ax, ay)
		bIn := isInside(bx, by)
		switch {
		case aIn && bIn:
			out = append(out, coords[s*bi:s*(bi+1)]...)
		case aIn && !bIn:
			out = appendIntersection(out, coords, ai, bi)
		case !aIn && bIn:
			out = appendIntersection(out, coords, ai, bi)
			out = append(out, coords[s*bi:s*(bi+1)]...)
		}
		ai = bi
	}
	// Ensure the output is explicitly closed. When the input's closing vertex
	// is outside the clip region, the degenerate closing edge emits nothing,
	// leaving the output open.
	nOut := len(out) / s
	if nOut > 0 && (out[0] != out[(nOut-1)*s] || out[1] != out[(nOut-1)*s+1]) {
		out = append(out, out[:s]...)
	}
	return out
}

// appendInterpX appends the intersection of the segment between points ai and
// bi with the vertical line x=k. The X coordinate is set exactly; Z and M are
// linearly interpolated.
func (c *polygonClipper) appendInterpX(dst, coords []float64, ai, bi int, x float64) []float64 {
	s := c.s
	ax := coords[s*ai]
	bx := coords[s*bi]
	t := (x - ax) / (bx - ax)
	for d := 0; d < s; d++ {
		dst = append(dst, lerp(coords[s*ai+d], coords[s*bi+d], t))
	}
	dst[len(dst)-s] = x // set exact X
	return dst
}

// appendInterpY appends the intersection of the segment between points ai and
// bi with the horizontal line y=k. The Y coordinate is set exactly; Z and M
// are linearly interpolated.
func (c *polygonClipper) appendInterpY(dst, coords []float64, ai, bi int, y float64) []float64 {
	s := c.s
	ay := coords[s*ai+1]
	by := coords[s*bi+1]
	t := (y - ay) / (by - ay)
	for d := 0; d < s; d++ {
		dst = append(dst, lerp(coords[s*ai+d], coords[s*bi+d], t))
	}
	dst[len(dst)-s+1] = y // set exact Y
	return dst
}

// rectBoundaryParam returns the CCW boundary parameter for a point on the rect
// boundary. The parameterisation starts at the bottom-left corner and goes
// counter-clockwise: bottom→right→top→left.
func (c *polygonClipper) rectBoundaryParam(xy XY) float64 {
	switch {
	case xy.Y == c.lo.Y: // bottom edge
		return xy.X - c.lo.X
	case xy.X == c.hi.X: // right edge
		return c.w + (xy.Y - c.lo.Y)
	case xy.Y == c.hi.Y: // top edge
		return c.w + c.h + (c.hi.X - xy.X)
	case xy.X == c.lo.X: // left edge
		return 2*c.w + c.h + (c.hi.Y - xy.Y)
	default:
		panic(fmt.Sprintf("point %v not on rect boundary [%v, %v]", xy, c.lo, c.hi))
	}
}

// paramToXY converts a CCW boundary parameter back to an XY coordinate.
func (c *polygonClipper) paramToXY(param float64) XY {
	switch {
	case param < c.w: // bottom edge
		return XY{X: c.lo.X + param, Y: c.lo.Y}
	case param < c.w+c.h: // right edge
		return XY{X: c.hi.X, Y: c.lo.Y + param - c.w}
	case param < 2*c.w+c.h: // top edge
		return XY{X: c.hi.X - (param - c.w - c.h), Y: c.hi.Y}
	case param < c.perim: // left edge
		return XY{X: c.lo.X, Y: c.hi.Y - (param - 2*c.w - c.h)}
	default:
		panic(fmt.Sprintf("boundary parameter %v out of range [0, %v)", param, c.perim))
	}
}

// ccwDist returns the CCW distance from param a to param b on the rect
// boundary.
func (c *polygonClipper) ccwDist(a, b float64) float64 {
	d := b - a
	if d < 0 {
		d += c.perim
	}
	return d
}

// isSameRectEdge returns true if both points lie on the same edge of the
// rectangle.
func (c *polygonClipper) isSameRectEdge(a, b XY) bool {
	return (a.X == c.lo.X && b.X == c.lo.X) ||
		(a.X == c.hi.X && b.X == c.hi.X) ||
		(a.Y == c.lo.Y && b.Y == c.lo.Y) ||
		(a.Y == c.hi.Y && b.Y == c.hi.Y)
}

// pointInRing returns true if xy is inside the ring defined by the given
// closed []float64 (first point == last point), using the ray-casting
// algorithm.
func (c *polygonClipper) pointInRing(xy XY, ring []float64) bool {
	s := c.s
	n := len(ring) / s
	inside := false
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		yi, yj := ring[s*i+1], ring[s*j+1]
		xi, xj := ring[s*i], ring[s*j]
		if (yi > xy.Y) != (yj > xy.Y) {
			slope := (xy.Y - yi) / (yj - yi)
			xIntersect := xi + slope*(xj-xi)
			if xy.X < xIntersect {
				inside = !inside
			}
		}
	}
	return inside
}

// removeDupConsecutive removes consecutive points with identical X,Y.
// For closed rings (first == last), the closing point is preserved.
func (c *polygonClipper) removeDupConsecutive(coords []float64) []float64 {
	s := c.s
	n := len(coords) / s
	if n == 0 {
		return nil
	}
	out := coords[:s] // keep first point
	for i := 1; i < n; i++ {
		prev := out[len(out)-s:]
		cur := coords[s*i : s*(i+1)]
		if cur[0] != prev[0] || cur[1] != prev[1] {
			out = append(out, cur...)
		}
	}
	return out
}
