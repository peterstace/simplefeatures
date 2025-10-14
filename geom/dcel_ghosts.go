package geom

import (
	"math"
	"sort"

	"github.com/peterstace/simplefeatures/rtree"
)

// createGhosts creates a MultiLineString that connects all components of the
// input Geometries using a ray-casting algorithm.
func createGhosts(a, b Geometry) MultiLineString {
	ctrlPts := collectControlPoints(a, b)
	lines := appendLines(nil, NewGeometryCollection([]Geometry{a, b}).AsGeometry())

	// Find the right-most point for each connected component in the overlaid
	// input geometries.
	representatives := findConnectedComponentRepresentatives(ctrlPts, lines)
	if len(representatives) <= 1 {
		return MultiLineString{} // 0 or 1 components are trivially connected.
	}

	// Process in right-to-left order to so ghost lines don't interfere with
	// each other.
	sort.Slice(representatives, func(i, j int) bool {
		return isMoreRightmost(representatives[i], representatives[j])
	})

	// Process each representative, casting rays rightward. If rays hit other
	// components, then we create the ghost line immediately. If they don't hit
	// anything, then we just store the origin so it can be connected later.
	pointIndex := newIndexedPoints(ctrlPts)
	lineIndex := newIndexedLines(lines)
	var ghostLines []line
	var noHitOrigins []XY
	for _, origin := range representatives {
		hitLocation, hasHit := findClosestRayIntersection(origin, pointIndex, lineIndex)
		if !hasHit {
			noHitOrigins = append(noHitOrigins, origin)
			continue
		}
		ghostLine := line{origin, hitLocation}
		ghostLines = append(ghostLines, ghostLine)
	}

	// When there are multiple components whose rays didn't hit anything, we
	// connect them to a common vertical line that's to the right of all other
	// components.
	if len(noHitOrigins) >= 2 {
		// Create the common vertical line.
		sort.Slice(noHitOrigins, func(i, j int) bool {
			return noHitOrigins[i].Y < noHitOrigins[j].Y
		})
		verticalLineX := math.Ceil(findMaxX(ctrlPts)) + 1
		for i := 0; i < len(noHitOrigins)-1; i++ {
			from := XY{verticalLineX, noHitOrigins[i].Y}
			to := XY{verticalLineX, noHitOrigins[i+1].Y}
			ghostLines = append(ghostLines, line{from, to})
		}

		// Create horizontal connections to the vertical line.
		for _, origin := range noHitOrigins {
			edge := line{origin, XY{verticalLineX, origin.Y}}
			ghostLines = append(ghostLines, edge)
		}
	}

	return linesToMultiLineString(ghostLines)
}

func linesToMultiLineString(lines []line) MultiLineString {
	lss := make([]LineString, len(lines))
	for i, ln := range lines {
		lss[i] = ln.asLineString()
	}
	return NewMultiLineString(lss)
}

// findMaxX returns the maximum X coordinate among all points. Panics if no
// points supplied.
func findMaxX(points []XY) float64 {
	if len(points) == 0 {
		panic("no points supplied")
	}
	maxX := points[0].X
	for i := range points {
		if points[i].X > maxX {
			maxX = points[i].X
		}
	}
	return maxX
}

// collectControlPoints collects all control points from both geometries and
// returns them deduplicated.
func collectControlPoints(a, b Geometry) []XY {
	var points []XY
	walk(a, func(xy XY) { points = append(points, xy) })
	walk(b, func(xy XY) { points = append(points, xy) })
	return sortAndUniquifyXYs(points)
}

// findConnectedComponentRepresentatives identifies connected components in the input
// geometries and returns the rightmost point from each component.
func findConnectedComponentRepresentatives(ctrlPts []XY, lines []line) []XY {
	if len(ctrlPts) == 0 {
		return nil
	}

	// Create point-to-index mapping.
	pointToIdx := make(map[XY]int, len(ctrlPts))
	for i, pt := range ctrlPts {
		pointToIdx[pt] = i
	}

	// Initialize union-find with all points as separate sets.
	dset := newDisjointSet(len(ctrlPts))

	// Union endpoints of all edges (since edges are connected).
	for _, ln := range lines {
		idxA, okA := pointToIdx[ln.a]
		idxB, okB := pointToIdx[ln.b]
		if okA && okB {
			dset.union(idxA, idxB)
		}
	}

	// Find the right-most point for each component (identified by its root in
	// the disjoint set).
	rootToRightmost := make(map[int]XY)
	for i, pt := range ctrlPts {
		root := dset.find(i)
		current, exists := rootToRightmost[root]
		if !exists || isMoreRightmost(pt, current) {
			rootToRightmost[root] = pt
		}
	}

	// Collect representative points.
	representatives := make([]XY, 0, len(rootToRightmost))
	for _, pt := range rootToRightmost {
		representatives = append(representatives, pt)
	}
	return representatives
}

// isMoreRightmost returns true if p1 is more rightmost than p2.
// A point is more rightmost if it has a larger X coordinate, or the same X
// coordinate with a larger Y coordinate.
func isMoreRightmost(p1, p2 XY) bool {
	if p1.X > p2.X {
		return true
	}
	return p1.X == p2.X && p1.Y > p2.Y
}

// horizontalRayIntersection finds the intersection of a horizontal ray
// starting at origin and continuing infinitely to the right (increasing X
// value).
func horizontalRayIntersection(
	origin XY,
	edge line,
) (XY, bool) {
	if origin == edge.a || origin == edge.b {
		// Origin is exactly on a vertex.
		return XY{}, false
	}

	// Handle horizontal edge special case.
	if edge.a.Y == edge.b.Y {
		if edge.a.Y != origin.Y {
			return XY{}, false
		}
		if edge.a.X > edge.b.X {
			// Ensure that a is to the left of b.
			edge.a, edge.b = edge.b, edge.a
		}
		if origin.X < edge.a.X {
			return edge.a, true
		}
		if origin.X < edge.b.X {
			return origin, true
		}
		return XY{}, false // beyond edge.b.X
	}

	// Non-horizontal case.
	if edge.a.Y > edge.b.Y {
		// Ensure that a is below b.
		edge.a, edge.b = edge.b, edge.a
	}
	t := (origin.Y - edge.a.Y) / (edge.b.Y - edge.a.Y)
	if t < 0 || t > 1 {
		return XY{}, false
	}
	x := edge.a.X + t*(edge.b.X-edge.a.X)
	if x < origin.X {
		return XY{}, false
	}
	return XY{x, origin.Y}, true
}

// findClosestRayIntersection casts a horizontal ray from origin in the +X
// direction and finds the closest intersection with any vertex or edge.
func findClosestRayIntersection(
	origin XY,
	pointIndex indexedPoints,
	lineIndex indexedLines,
) (XY, bool) {
	// Create bounding box for the rightward horizontal ray.
	rayBox := rtree.Box{
		MinX: origin.X,
		MaxX: math.MaxFloat64,
		MinY: origin.Y,
		MaxY: origin.Y,
	}

	closestDist := math.MaxFloat64
	var hasHit bool
	var hitLocation XY

	// Check for vertex intersections.
	pointIndex.tree.RangeSearch(rayBox, func(i int) error {
		pt := pointIndex.points[i]
		if pt.X <= origin.X || pt.Y != origin.Y {
			return nil
		}
		dist := pt.X - origin.X
		if dist < closestDist {
			closestDist = dist
			hasHit = true
			hitLocation = pt
		}
		return nil
	})

	// Check for edge intersections.
	lineIndex.tree.RangeSearch(rayBox, func(i int) error {
		edge := lineIndex.lines[i]

		inter, ok := horizontalRayIntersection(origin, edge)
		if !ok {
			return nil
		}
		dist := inter.X - origin.X
		if dist < closestDist {
			closestDist = dist
			hasHit = true
			hitLocation = inter
		}
		return nil
	})

	return hitLocation, hasHit
}
