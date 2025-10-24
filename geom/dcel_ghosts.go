package geom

import (
	"math"
	"sort"
)

// findComponentRepresentatives identifies connected components in the input
// geometries and returns the rightmost point from each component. These
// representative points will be used for ghost edge construction.
func findComponentRepresentatives(a, b Geometry) []XY {
	// Collect all control points from both geometries.
	var points []XY
	walk(a, func(xy XY) { points = append(points, xy) })
	walk(b, func(xy XY) { points = append(points, xy) })

	if len(points) == 0 {
		return nil
	}

	// Deduplicate points and create point-to-index mapping.
	points = sortAndUniquifyXYs(points)
	pointToIdx := make(map[XY]int, len(points))
	for i, pt := range points {
		pointToIdx[pt] = i
	}

	// Initialize union-find with all points as separate components.
	dset := newDisjointSet(len(points))

	// Union endpoints of all edges (since edges are connected).
	all := NewGeometryCollection([]Geometry{a, b}).AsGeometry()
	lines := appendLines(nil, all) // TODO: Write a walkEdges function and use that instead.
	for _, ln := range lines {
		idxA, okA := pointToIdx[ln.a]
		idxB, okB := pointToIdx[ln.b]
		if okA && okB {
			dset.union(idxA, idxB)
		}
	}

	// Find the rightmost point for each component.
	rightmost := make(map[int]XY)
	for _, pt := range points {
		root := dset.find(pointToIdx[pt])
		current, exists := rightmost[root]
		if !exists || isMoreRightmost(pt, current) {
			rightmost[root] = pt
		}
	}

	// Collect representative points.
	representatives := make([]XY, 0, len(rightmost))
	for _, pt := range rightmost {
		representatives = append(representatives, pt)
	}
	// TODO: Sort representatives for consistent output?

	return representatives
}

// isMoreRightmost returns true if p1 is more rightmost than p2.
// A point is more rightmost if it has a larger X coordinate, or the same X
// coordinate with a larger Y coordinate.
func isMoreRightmost(p1, p2 XY) bool {
	if p1.X > p2.X {
		return true
	}
	if p1.X == p2.X && p1.Y > p2.Y {
		return true
	}
	return false
}

// sortRightmostFirst sorts points right-to-left (descending X, then
// descending Y).
func sortRightmostFirst(points []XY) {
	sort.Slice(points, func(i, j int) bool {
		if points[i].X != points[j].X {
			return points[i].X > points[j].X
		}
		return points[i].Y > points[j].Y
	})
}

// collectAllPoints collects all control points from both geometries and
// returns them deduplicated.
func collectAllPoints(a, b Geometry) []XY {
	var points []XY
	walk(a, func(xy XY) { points = append(points, xy) })
	walk(b, func(xy XY) { points = append(points, xy) })
	return sortAndUniquifyXYs(points)
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

// isObstructed checks if there is any control point or edge between origin and
// target. Returns true if the path from origin to target is obstructed.
func isObstructed(origin, target XY, allPoints []XY, allLines []line) bool {
	segment := line{origin, target}

	// TODO: Should use index structure here.

	// Check if any point lies on the segment (excluding endpoints).
	for _, pt := range allPoints {
		if pt == origin || pt == target {
			continue
		}
		if segment.intersectsXY(pt) {
			return true
		}
	}

	// Check if any edge intersects the segment.
	for _, edge := range allLines {
		inter := segment.intersectLine(edge)
		if inter.empty {
			continue
		}

		// Check if intersection is not just at the origin or target endpoints.
		if inter.ptA != origin && inter.ptA != target {
			return true
		}
		if inter.ptA != inter.ptB && inter.ptB != origin && inter.ptB != target {
			return true
		}
	}

	return false
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
	hitEdge  line
}

// findClosestRayIntersection casts a horizontal ray from origin in the +X
// direction and finds the closest intersection with any vertex or edge.
func findClosestRayIntersection(
	origin XY,
	pointIndex indexedPoints, // TODO: pointIndex and lineIndex are unused
	lineIndex indexedLines,
	allPoints []XY,
	allLines []line,
) rayHitResult {
	closestDist := math.MaxFloat64
	result := rayHitResult{hitType: hitNone}

	// Check for vertex intersections.
	for _, pt := range allPoints {
		if pt.X <= origin.X || pt.Y != origin.Y {
			continue
		}
		dist := pt.X - origin.X
		if dist < closestDist {
			closestDist = dist
			result = rayHitResult{
				hitType:  hitVertex,
				hitPoint: pt,
			}
		}
	}

	// Check for edge intersections.
	for _, edge := range allLines {
		// Create a ray as a very long horizontal line segment.
		ray := line{origin, XY{origin.X + 1e10, origin.Y}}
		inter := ray.intersectLine(edge)
		if inter.empty {
			continue
		}

		// Only consider intersections to the right of origin.
		if inter.ptA.X <= origin.X {
			continue
		}

		dist := inter.ptA.X - origin.X
		if dist < closestDist {
			closestDist = dist
			result = rayHitResult{
				hitType:  hitEdge,
				hitPoint: inter.ptA,
				hitEdge:  edge,
			}
		}
	}

	return result
}

// createGhostFromHit creates a ghost edge from origin to the intersection
// found by ray casting. Handles both vertex hits and edge hits.
func createGhostFromHit(
	origin XY,
	hitResult rayHitResult,
	allPoints []XY,
	allLines []line,
) LineString {
	if hitResult.hitType == hitVertex {
		// Case A: Ray hits a vertex directly.
		return line{origin, hitResult.hitPoint}.asLineString()
	}

	// Case B: Ray hits an edge - check endpoints for obstructions.
	edge := hitResult.hitEdge

	allowA := !isObstructed(origin, edge.a, allPoints, allLines) && edge.a.X > origin.X
	allowB := !isObstructed(origin, edge.b, allPoints, allLines) && edge.b.X > origin.X

	lineTo := func(to XY) LineString {
		return line{origin, to}.asLineString()
	}

	if allowA && allowB {
		// Both endpoints allowed. Choose the closer one. If they're same
		// distance, than choose the higher one.
		distA := origin.distanceSquaredTo(edge.a)
		distB := origin.distanceSquaredTo(edge.b)
		switch {
		case distA < distB:
			return lineTo(edge.a)
		case distA > distB:
			return lineTo(edge.b)
		default:
			if edge.a.Y >= edge.b.Y {
				return lineTo(edge.a)
			} else {
				return lineTo(edge.b)
			}
		}
		panic("unreachable")
	}

	if allowA {
		return lineTo(edge.a)
	}
	if allowB {
		return lineTo(edge.b)
	}

	// Both endpoints obstructed - connect to intersection point.
	return lineTo(hitResult.hitPoint)
}

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
	sortRightmostFirst(representatives)

	// Build spatial indexes and collect geometry data.
	allPoints := collectAllPoints(a, b)
	all := NewGeometryCollection([]Geometry{a, b}).AsGeometry()
	allLines := appendLines(nil, all)
	pointIndex := newIndexedPoints(allPoints)
	lineIndex := newIndexedLines(allLines)

	// Process each representative, casting rays rightward.
	var ghostEdges []LineString // TODO: collect ghostLines instead of ghostEdges (and convert to MultiLineString at the end).
	var fallbackOrigins []XY

	for _, origin := range representatives {
		hitResult := findClosestRayIntersection(
			origin, pointIndex, lineIndex, allPoints, allLines,
		)

		if hitResult.hitType == hitNone {
			// No intersection - would need vertical line connection.
			fallbackOrigins = append(fallbackOrigins, origin)
			continue
		}

		// Can create a ghost edge to an actual component.
		ghostEdge := createGhostFromHit(
			origin, hitResult, allPoints, allLines,
		)
		ghostEdges = append(ghostEdges, ghostEdge)
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
			edge := line{origin, XY{verticalLineX, origin.Y}}.asLineString()
			ghostEdges = append(ghostEdges, edge)
		}

		// Create vertical line segments connecting consecutive horizontal endpoints.
		for i := 0; i < len(fallbackOrigins)-1; i++ {
			from := XY{verticalLineX, fallbackOrigins[i].Y}
			to := XY{verticalLineX, fallbackOrigins[i+1].Y}
			verticalSegment := line{from, to}.asLineString()
			ghostEdges = append(ghostEdges, verticalSegment)
		}
	}

	return NewMultiLineString(ghostEdges)
}
