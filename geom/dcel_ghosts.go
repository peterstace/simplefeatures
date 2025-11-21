package geom

import (
	"math"
	"sort"

	"github.com/peterstace/simplefeatures/rtree"
)

// createGhosts creates a MultiLineString that connects all components of the
// input Geometries using a ray-casting algorithm.
func createGhosts(a, b Geometry) MultiLineString {
	// Get representative points for each component.
	representatives := findComponentRepresentatives(a, b)

	if len(representatives) <= 1 {
		// When there are either 0 or 1 connected components, then they don't
		// need connecting.
		return MultiLineString{}
	}

	// Sort right-to-left for processing.
	sort.Slice(representatives, func(i, j int) bool {
		return isMoreRightmost(representatives[i], representatives[j])
	})

	// Build spatial indexes and collect geometry data.
	allPoints := collectAllPoints(a, b)
	all := NewGeometryCollection([]Geometry{a, b}).AsGeometry()
	allLines := appendLines(nil, all)
	pointIndex := newIndexedPoints(allPoints)
	lineIndex := newIndexedLines(allLines)

	// Process each representative, casting rays rightward.
	var ghostLines []line
	var fallbackOrigins []XY

	for _, origin := range representatives {
		hitResult := findClosestRayIntersection(
			origin, pointIndex, lineIndex,
		)

		if hitResult.hitType == hitNone {
			// No intersection - would need vertical line connection.
			fallbackOrigins = append(fallbackOrigins, origin)
			continue
		}

		// Can create a ghost edge to an actual component.
		ghostLine := createGhostFromHit(origin, hitResult)
		ghostLines = append(ghostLines, ghostLine)
	}

	// Only create vertical line connections if at least 2 components need it.
	if len(fallbackOrigins) >= 2 {
		// Sort vertical line origins by Y coordinate.
		sort.Slice(fallbackOrigins, func(i, j int) bool {
			return fallbackOrigins[i].Y < fallbackOrigins[j].Y
		})

		// Calculate max X for vertical line fallback.
		maxX := findMaxX(allPoints)
		verticalLineX := math.Ceil(maxX) + 1

		// Create horizontal connections to the vertical line.
		for _, origin := range fallbackOrigins {
			edge := line{origin, XY{verticalLineX, origin.Y}}
			ghostLines = append(ghostLines, edge)
		}

		// Create vertical line segments connecting consecutive horizontal endpoints.
		for i := 0; i < len(fallbackOrigins)-1; i++ {
			from := XY{verticalLineX, fallbackOrigins[i].Y}
			to := XY{verticalLineX, fallbackOrigins[i+1].Y}
			verticalSegment := line{from, to}
			ghostLines = append(ghostLines, verticalSegment)
		}
	}

	// Convert lines to LineStrings.
	ghostEdges := make([]LineString, len(ghostLines))
	for i, ln := range ghostLines {
		ghostEdges[i] = ln.asLineString()
	}
	return NewMultiLineString(ghostEdges)
}

// findMaxX returns the maximum X coordinate among all points. If no points are
// supplied, then returns 0 as a special case.
func findMaxX(points []XY) float64 {
	if len(points) == 0 {
		return 0
	}
	maxX := points[0].X
	for i := range points {
		if points[i].X > maxX {
			maxX = points[i].X
		}
	}
	return maxX
}

// collectAllPoints collects all control points from both geometries and
// returns them deduplicated.
//
// TODO: The same logic is present at the start of
// findComponentRepresentatives. This should get cleaned up.
func collectAllPoints(a, b Geometry) []XY {
	var points []XY
	walkXY(a, func(xy XY) { points = append(points, xy) })
	walkXY(b, func(xy XY) { points = append(points, xy) })
	return sortAndUniquifyXYs(points)
}

// findComponentRepresentatives identifies connected components in the input
// geometries and returns the rightmost point from each component.
func findComponentRepresentatives(a, b Geometry) []XY {
	// Collect all control points from both geometries.
	var points []XY
	walkXY(a, func(xy XY) { points = append(points, xy) })
	walkXY(b, func(xy XY) { points = append(points, xy) })

	if len(points) == 0 {
		return nil
	}

	// Deduplicate points and create point-to-index mapping.
	points = sortAndUniquifyXYs(points)
	pointToIdx := make(map[XY]int, len(points))
	for i, pt := range points {
		pointToIdx[pt] = i
	}

	// Initialize union-find with all points as separate sets.
	dset := newDisjointSet(len(points))

	// Union endpoints of all edges (since edges are connected).
	walkLines(NewGeometryCollection([]Geometry{a, b}).AsGeometry(), func(ln line) {
		idxA, okA := pointToIdx[ln.a]
		idxB, okB := pointToIdx[ln.b]
		if okA && okB {
			dset.union(idxA, idxB)
		}
	})

	// Find the right-most point for each component (identified by its root in
	// the disjoint set).
	rootToRightmost := make(map[int]XY)
	for i, pt := range points {
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

// rayHitType represents the type of intersection found by ray casting.
type rayHitType int

const (
	hitNone rayHitType = iota
	hitVertex
	hitEdge
)

// rayHitResult contains information about a ray intersection.
type rayHitResult struct {
	hitType  rayHitType
	hitPoint XY
}

// findClosestRayIntersection casts a horizontal ray from origin in the +X
// direction and finds the closest intersection with any vertex or edge.
func findClosestRayIntersection(
	origin XY,
	pointIndex indexedPoints,
	lineIndex indexedLines,
) rayHitResult {
	closestDist := math.MaxFloat64
	result := rayHitResult{hitType: hitNone}

	// Create bounding box for the rightward horizontal ray.
	rayBox := rtree.Box{
		MinX: origin.X,
		MaxX: math.MaxFloat64,
		MinY: origin.Y,
		MaxY: origin.Y,
	}

	// Check for vertex intersections.
	pointIndex.tree.RangeSearch(rayBox, func(i int) error {
		pt := pointIndex.points[i]
		if pt.X <= origin.X || pt.Y != origin.Y {
			return nil
		}
		dist := pt.X - origin.X
		if dist < closestDist {
			closestDist = dist
			result = rayHitResult{
				hitType:  hitVertex,
				hitPoint: pt,
			}
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
			result = rayHitResult{
				hitType:  hitEdge,
				hitPoint: inter,
			}
		}
		return nil
	})

	return result
}

// createGhostFromHit creates a ghost edge from origin to the intersection
// found by ray casting. Always draws a horizontal ray to the hit point.
func createGhostFromHit(origin XY, hitResult rayHitResult) line {
	return line{origin, hitResult.hitPoint}
}
