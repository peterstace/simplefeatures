package geom

import (
	"cmp"
	"fmt"
	"slices"
)

func clipPolygonByRect(p Polygon, rect Envelope) Geometry {
	ctype := p.CoordinatesType()
	emptyPoly := NewPolygon(nil).ForceCoordinatesType(ctype).AsGeometry()

	if p.IsEmpty() {
		return emptyPoly
	}

	// Normalise to CCW exterior / CW holes. This ensures all clipped rings
	// have known winding, which the topology resolution depends on.
	p = p.ForceCCW()

	// Degenerate rect: point or line envelope → empty polygon.
	if !rect.IsRectangle() {
		return emptyPoly
	}

	min, max, ok := rect.MinMaxXYs()
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

	// Clip exterior ring.
	clippedExt := clipRingSH(p.ExteriorRing().Coordinates(), min, max)
	if len(clippedExt) == 0 {
		return emptyPoly
	}

	// Classify holes.
	var freeHoles []LineString
	var clippedHoles [][]Coordinates
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
		if rect.Covers(holeEnv) && !holeOnRectBoundary(hole.Coordinates(), min, max) {
			freeHoles = append(freeHoles, hole)
			continue
		}

		// Hole touches or crosses rect boundary — clip it.
		clippedHole := clipRingSH(hole.Coordinates(), min, max)
		if len(clippedHole) > 0 {
			clippedHoles = append(clippedHoles, clippedHole)
		}
	}

	return resolveClippedPolygon(clippedExt, freeHoles, clippedHoles, min, max, ctype)
}

// holeOnRectBoundary reports whether any vertex of the sequence lies on the
// rect boundary.
func holeOnRectBoundary(seq Sequence, min, max XY) bool {
	for i := 0; i < seq.Length(); i++ {
		xy := seq.GetXY(i)
		if xy.X == min.X || xy.X == max.X || xy.Y == min.Y || xy.Y == max.Y {
			return true
		}
	}
	return false
}

// resolveClippedPolygon takes the clipped exterior ring, classified holes, and
// the rect bounds, and produces the output Geometry (Polygon or MultiPolygon).
// The input polygon must have been normalised to CCW (exterior CCW, holes CW)
// before clipping, so that all clipped rings have known winding.
func resolveClippedPolygon(
	clippedExterior []Coordinates,
	freeHoles []LineString,
	clippedHoles [][]Coordinates,
	min, max XY,
	ctype CoordinatesType,
) Geometry {
	extArcs := extractInteriorArcs(clippedExterior, min, max)

	// Collect all arcs.
	var allArcs []interiorArc
	allArcs = append(allArcs, extArcs...)

	emptyPoly := NewPolygon(nil).ForceCoordinatesType(ctype).AsGeometry()
	for _, hole := range clippedHoles {
		arcs := extractInteriorArcs(hole, min, max)
		if len(arcs) == 0 {
			// Clipped hole is entirely on the rect boundary — it covers
			// the entire rect. No polygon area survives.
			return emptyPoly
		}
		allArcs = append(allArcs, arcs...)
	}

	var outputRings [][]Coordinates
	if len(allArcs) == 0 {
		// No interior arcs: the polygon contains the entire rect.
		// Output is the rect itself (closed ring).
		w := max.X - min.X
		h := max.Y - min.Y
		corners := [4]float64{0, w, w + h, 2*w + h}
		ring := make([]Coordinates, 5)
		for i, cp := range corners {
			ring[i] = Coordinates{XY: paramToXY(cp, min, max), Type: ctype}
		}
		ring[4] = ring[0] // close the ring
		outputRings = append(outputRings, ring)
	} else {
		outputRings = walkArcs(allArcs, min, max)
	}

	// Build output polygons. Each output ring is an exterior (already closed).
	var polys []Polygon
	for _, ring := range outputRings {
		seq := coordsToSeq(ring, ctype)
		exterior := NewLineString(seq)
		polys = append(polys, NewPolygon([]LineString{exterior}))
	}

	// Assign free holes to the correct exterior ring.
	for _, hole := range freeHoles {
		holeXY, ok := hole.StartPoint().XY()
		if !ok {
			continue
		}
		// TODO: Use R-Tree to speed up assignment to the correct exterior ring.
		for i, ring := range outputRings {
			if pointInRingXY(holeXY, ring) {
				existingRings := polys[i].DumpRings()
				existingRings = append(existingRings, hole)
				polys[i] = NewPolygon(existingRings)
				break
			}
		}
	}

	// Return Polygon or MultiPolygon.
	if len(polys) == 0 {
		return NewPolygon(nil).ForceCoordinatesType(ctype).AsGeometry()
	}
	if len(polys) == 1 {
		return polys[0].AsGeometry()
	}
	return NewMultiPolygon(polys).AsGeometry()
}

// walkArcs performs the topology resolution walk, producing output rings from
// the collected interior arcs. It pairs arc endpoints along the rect boundary
// (CCW) and traces complete rings.
func walkArcs(arcs []interiorArc, min, max XY) [][]Coordinates {
	w := max.X - min.X
	h := max.Y - min.Y
	perim := 2*w + 2*h

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
		bestDist := perim + 1
		for _, sp := range startParams {
			d := ccwDist(endParam, sp, perim)
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
	var rings [][]Coordinates

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

		var ring []Coordinates
		curIdx := firstIdx

		for {
			used[curIdx] = true
			a := arcs[curIdx]

			// Append arc coordinates.
			ring = append(ring, a.coords...)

			// Find next arc via boundary.
			nextStartParam, nextIdx := findNextStart(a.endParam)

			// Build boundary path from this arc's end to the next arc's start.
			endCoord := ring[len(ring)-1]
			startCoord := arcs[nextIdx].coords[0]

			bpath := buildBoundaryPath(a.endParam, nextStartParam, endCoord, startCoord, min, max)
			ring = append(ring, bpath...)

			if nextIdx == firstIdx {
				break // Ring complete.
			}
			curIdx = nextIdx
		}

		// Close the ring.
		ring = append(ring, ring[0])
		ring = removeDupConsecutiveCoords(ring)
		if len(ring) >= 4 {
			rings = append(rings, ring)
		}
	}

	return rings
}

// buildBoundaryPath returns the coordinates along the rect boundary going CCW
// from startParam to endParam. It includes rect corners between the two
// parameters but does NOT include the start or end points themselves (those are
// the last/first coordinates of the adjacent arcs). Z and M values at corners
// are linearly interpolated between startCoord and endCoord based on boundary
// distance.
func buildBoundaryPath(startParam, endParam float64, startCoord, endCoord Coordinates, min, max XY) []Coordinates {
	w := max.X - min.X
	h := max.Y - min.Y
	perim := 2*w + 2*h

	// Corner parameters in CCW order.
	corners := [4]float64{0, w, w + h, 2*w + h}

	// Find the first corner after startParam going CCW. The corners are in
	// ascending parameter order, so this is the first with cp > startParam.
	// If startParam is past the last corner (on the left edge), no corner
	// qualifies and firstIdx stays at 0 — the bottom-left corner is the
	// next one going CCW.
	firstIdx := 0
	for i, cp := range corners {
		if cp > startParam {
			firstIdx = i
			break
		}
	}

	// Iterate corners in CCW order from firstIdx, collecting those strictly
	// between startParam and endParam.
	totalDist := ccwDist(startParam, endParam, perim)
	var path []Coordinates
	for k := 0; k < 4; k++ {
		cp := corners[(firstIdx+k)%4]
		d := ccwDist(startParam, cp, perim)
		if d >= totalDist {
			break
		}
		frac := d / totalDist
		c := interpolateCoords(startCoord, endCoord, frac)
		c.XY = paramToXY(cp, min, max)
		path = append(path, c)
	}
	return path
}

// extractInteriorArcs decomposes a clipped ring into interior arcs. The ring
// must be explicitly closed (first == last). Each arc starts and ends at a
// point on the rect boundary. Returns an empty slice if the ring has no
// interior arcs (entirely on the boundary or entirely interior with no
// boundary contact).
func extractInteriorArcs(ring []Coordinates, min, max XY) []interiorArc {
	n := len(ring)
	if n < 4 {
		return nil
	}

	// There are n-1 edges in a closed ring (the last vertex == first vertex,
	// so edge n-1 from ring[n-1] to ring[0] is degenerate and skipped).
	numEdges := n - 1

	// Classify each edge as boundary (both endpoints on same rect edge) or interior.
	isBdry := make([]bool, numEdges)
	for i := 0; i < numEdges; i++ {
		isBdry[i] = isSameRectEdge(ring[i].XY, ring[i+1].XY, min, max)
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
		var arcCoords []Coordinates
		arcCoords = append(arcCoords, ring[i%numEdges])
		i++
		steps++
		for steps < numEdges && !isBdry[i%numEdges] {
			arcCoords = append(arcCoords, ring[i%numEdges])
			i++
			steps++
		}
		if steps < numEdges {
			// The arc ends at ring[i%numEdges] (the start of the next boundary edge).
			arcCoords = append(arcCoords, ring[i%numEdges])
		} else {
			// Wrapped around; the arc ends at ring[start] (where we began).
			arcCoords = append(arcCoords, ring[start])
		}

		sp := rectBoundaryParam(arcCoords[0].XY, min, max)
		ep := rectBoundaryParam(arcCoords[len(arcCoords)-1].XY, min, max)
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
// last elements of coords are on the rectangle boundary; everything in between
// is in the interior.
type interiorArc struct {
	coords     []Coordinates
	startParam float64 // boundary parameter of coords[0]
	endParam   float64 // boundary parameter of coords[len(coords)-1]
}

// clipRingSH clips a closed ring against an axis-aligned rectangle using the
// Sutherland-Hodgman algorithm. It returns the clipped ring as a closed slice
// of [Coordinates] (first == last), or nil if the ring is entirely outside the
// rectangle. The input ring must be explicitly closed.
func clipRingSH(seq Sequence, min, max XY) []Coordinates {
	coords := seqToCoords(seq)
	if len(coords) < 4 {
		return nil
	}

	// Clip against each of the 4 edges.
	coords = clipToEdge(coords,
		func(c Coordinates) bool { return c.X >= min.X },
		func(a, b Coordinates) Coordinates { return interpX(a, b, min.X) })
	coords = clipToEdge(coords,
		func(c Coordinates) bool { return c.X <= max.X },
		func(a, b Coordinates) Coordinates { return interpX(a, b, max.X) })
	coords = clipToEdge(coords,
		func(c Coordinates) bool { return c.Y >= min.Y },
		func(a, b Coordinates) Coordinates { return interpY(a, b, min.Y) })
	coords = clipToEdge(coords,
		func(c Coordinates) bool { return c.Y <= max.Y },
		func(a, b Coordinates) Coordinates { return interpY(a, b, max.Y) })

	coords = removeDupConsecutiveCoords(coords)
	if len(coords) < 4 {
		return nil
	}
	return coords
}

// clipToEdge performs one pass of the Sutherland-Hodgman algorithm, clipping a
// closed ring against a single half-plane defined by isInside. The intersect
// function computes the intersection of segment a→b with the clipping edge.
// The output is also an explicitly closed ring.
func clipToEdge(
	coords []Coordinates,
	isInside func(Coordinates) bool,
	intersect func(Coordinates, Coordinates) Coordinates,
) []Coordinates {
	if len(coords) == 0 {
		return nil
	}
	var output []Coordinates
	a := coords[len(coords)-1]
	for _, b := range coords {
		aIn := isInside(a)
		bIn := isInside(b)
		switch {
		case aIn && bIn:
			output = append(output, b)
		case aIn && !bIn:
			output = append(output, intersect(a, b))
		case !aIn && bIn:
			output = append(output, intersect(a, b))
			output = append(output, b)
		}
		a = b
	}
	// Ensure the output is explicitly closed. When the input's closing vertex
	// is outside the clip region, the degenerate closing edge emits nothing,
	// leaving the output open.
	if len(output) > 0 && output[0].XY != output[len(output)-1].XY {
		output = append(output, output[0])
	}
	return output
}

// interpX returns the intersection of segment a→b with the vertical line x=k.
// The result's X coordinate is set to x exactly to ensure reliable boundary
// detection. Z and M are interpolated via [interpolateCoords].
func interpX(a, b Coordinates, x float64) Coordinates {
	t := (x - a.X) / (b.X - a.X)
	c := interpolateCoords(a, b, t)
	c.XY.X = x
	return c
}

// interpY returns the intersection of segment a→b with the horizontal line
// y=k. The result's Y coordinate is set to y exactly to ensure reliable
// boundary detection. Z and M are interpolated via [interpolateCoords].
func interpY(a, b Coordinates, y float64) Coordinates {
	t := (y - a.Y) / (b.Y - a.Y)
	c := interpolateCoords(a, b, t)
	c.XY.Y = y
	return c
}

// rectBoundaryParam returns the CCW boundary parameter for a point on the rect
// boundary. The parameterisation starts at the bottom-left corner and goes
// counter-clockwise: bottom→right→top→left.
func rectBoundaryParam(xy, min, max XY) float64 {
	w := max.X - min.X
	h := max.Y - min.Y
	switch {
	case xy.Y == min.Y: // bottom edge
		return xy.X - min.X
	case xy.X == max.X: // right edge
		return w + (xy.Y - min.Y)
	case xy.Y == max.Y: // top edge
		return w + h + (max.X - xy.X)
	case xy.X == min.X: // left edge
		return 2*w + h + (max.Y - xy.Y)
	default:
		panic(fmt.Sprintf("point %v not on rect boundary [%v, %v]", xy, min, max))
	}
}

// paramToXY converts a CCW boundary parameter back to an XY coordinate.
func paramToXY(param float64, min, max XY) XY {
	w := max.X - min.X
	h := max.Y - min.Y
	switch {
	case param < w: // bottom edge
		return XY{X: min.X + param, Y: min.Y}
	case param < w+h: // right edge
		return XY{X: max.X, Y: min.Y + param - w}
	case param < 2*w+h: // top edge
		return XY{X: max.X - (param - w - h), Y: max.Y}
	case param < 2*w+2*h: // left edge
		return XY{X: min.X, Y: max.Y - (param - 2*w - h)}
	default:
		panic(fmt.Sprintf("boundary parameter %v out of range [0, %v)", param, 2*w+2*h))
	}
}

// ccwDist returns the CCW distance from param a to param b on a boundary with
// total perimeter perim.
func ccwDist(a, b, perim float64) float64 {
	d := b - a
	if d < 0 {
		d += perim
	}
	return d
}

// pointInRingXY returns true if xy is inside the ring defined by the given
// closed coordinates (first == last), using the ray-casting algorithm.
func pointInRingXY(xy XY, ring []Coordinates) bool {
	inside := false
	n := len(ring)
	for i := range ring {
		j := (i + 1) % n
		yi, yj := ring[i].Y, ring[j].Y
		xi, xj := ring[i].X, ring[j].X
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

// seqToCoords extracts all Coordinates from a Sequence into a slice.
func seqToCoords(seq Sequence) []Coordinates {
	n := seq.Length()
	coords := make([]Coordinates, n)
	for i := range coords {
		coords[i] = seq.Get(i)
	}
	return coords
}

// coordsToSeq converts a slice of Coordinates into a Sequence.
func coordsToSeq(coords []Coordinates, ctype CoordinatesType) Sequence {
	dim := ctype.Dimension()
	floats := make([]float64, 0, len(coords)*dim)
	for _, c := range coords {
		c.Type = ctype
		floats = c.appendFloat64s(floats)
	}
	return NewSequence(floats, ctype)
}

// removeDupConsecutiveCoords removes consecutive vertices with identical XY.
// For closed rings (first == last), the closing vertex is preserved.
func removeDupConsecutiveCoords(coords []Coordinates) []Coordinates {
	if len(coords) == 0 {
		return nil
	}
	out := coords[:1]
	for _, c := range coords[1:] {
		if c.XY != out[len(out)-1].XY {
			out = append(out, c)
		}
	}
	return out
}

// isSameRectEdge returns true if both points lie on the same edge of the
// rectangle.
func isSameRectEdge(a, b, min, max XY) bool {
	return (a.X == min.X && b.X == min.X) ||
		(a.X == max.X && b.X == max.X) ||
		(a.Y == min.Y && b.Y == min.Y) ||
		(a.Y == max.Y && b.Y == max.Y)
}
