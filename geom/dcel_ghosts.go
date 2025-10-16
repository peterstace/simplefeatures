package geom

import (
	"fmt"
	"sort"

	"github.com/peterstace/simplefeatures/rtree"
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

	// Union endpoints of all edges to build connected components.
	all := NewGeometryCollection([]Geometry{a, b}).AsGeometry()
	lines := appendLines(nil, all)
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
// descending Y). This order is used for processing components during ghost
// edge construction.
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

// findMaxX returns the maximum X coordinate among all points.
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
	pointIndex indexedPoints,
	lineIndex indexedLines,
	allPoints []XY,
	allLines []line,
) rayHitResult {
	closestDist := float64(1<<63 - 1) // Max float64 approximation.
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

	aObstructed := isObstructed(origin, edge.a, allPoints, allLines)
	bObstructed := isObstructed(origin, edge.b, allPoints, allLines)

	if !aObstructed && !bObstructed {
		// Both endpoints unobstructed - choose the closer one.
		if origin.distanceSquaredTo(edge.a) <= origin.distanceSquaredTo(edge.b) {
			return line{origin, edge.a}.asLineString()
		}
		return line{origin, edge.b}.asLineString()
	}

	if !aObstructed {
		return line{origin, edge.a}.asLineString()
	}
	if !bObstructed {
		return line{origin, edge.b}.asLineString()
	}

	// Both endpoints obstructed - connect to intersection point.
	return line{origin, hitResult.hitPoint}.asLineString()
}

// createGhosts creates a MultiLineString that connects all components of the
// input Geometries using a ray-casting algorithm that minimizes crossings with
// input geometry.
func createGhosts(a, b Geometry) MultiLineString {
	// Get representative points for each component.
	representatives := findComponentRepresentatives(a, b)

	if len(representatives) <= 1 {
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

	// Calculate max X for vertical line fallback.
	maxX := findMaxX(allPoints)

	// Process each representative, casting rays rightward.
	var ghostEdges []LineString
	var verticalLineOrigins []XY

	for _, origin := range representatives {
		hitResult := findClosestRayIntersection(
			origin, pointIndex, lineIndex, allPoints, allLines,
		)

		if hitResult.hitType == hitNone {
			// No intersection - would need vertical line connection.
			verticalLineOrigins = append(verticalLineOrigins, origin)
			continue
		}

		// Can create a ghost edge to an actual component.
		ghostEdge := createGhostFromHit(
			origin, hitResult, allPoints, allLines,
		)
		ghostEdges = append(ghostEdges, ghostEdge)
	}

	// Only create vertical line connections if at least 2 components need it.
	if len(verticalLineOrigins) >= 2 {
		verticalLineX := maxX + 2
		for _, origin := range verticalLineOrigins {
			ghostEdge := line{origin, XY{verticalLineX, origin.Y}}.asLineString()
			ghostEdges = append(ghostEdges, ghostEdge)
		}
	}

	return NewMultiLineString(ghostEdges)
}

// spanningTree creates a near-minimum spanning tree (using the euclidean
// distance metric) over the supplied points. The tree will consist of N-1
// lines, where N is the number of _distinct_ xys supplied.
//
// It's a 'near' minimum spanning tree rather than a spanning tree, because we
// use a simple greedy algorithm rather than a proper minimum spanning tree
// algorithm.
func spanningTree(xys []XY) MultiLineString {
	if len(xys) <= 1 {
		return MultiLineString{}
	}

	// Load points into r-tree.
	xys = sortAndUniquifyXYs(xys)
	items := make([]rtree.BulkItem, len(xys))
	for i, xy := range xys {
		items[i] = rtree.BulkItem{Box: xy.box(), RecordID: i}
	}
	tree := rtree.BulkLoad(items)

	// The disjoint set keeps track of which points have been joined together
	// so far. Two entries in dset are in the same set iff they are connected
	// in the incrementally-built spanning tree.
	dset := newDisjointSet(len(xys))
	lss := make([]LineString, 0, len(xys)-1)

	for i, xyi := range xys {
		if i == len(xys)-1 {
			// Skip the last point, since a tree is formed from N-1 edges
			// rather than N edges. The last point will be included by virtue
			// of being the closest to another point.
			continue
		}
		tree.PrioritySearch(xyi.box(), func(j int) error {
			// We don't want to include a new edge in the spanning tree if it
			// would cause a cycle (i.e. the two endpoints are already in the
			// same tree). This is checked via dset.
			if i == j || dset.find(i) == dset.find(j) {
				return nil
			}
			dset.union(i, j)
			xyj := xys[j]
			lss = append(lss, line{xyi, xyj}.asLineString())
			return rtree.Stop
		})
	}

	return NewMultiLineString(lss)
}

func appendXYForPoint(xys []XY, pt Point) []XY {
	if xy, ok := pt.XY(); ok {
		xys = append(xys, xy)
	}
	return xys
}

func appendXYForLineString(xys []XY, ls LineString) []XY {
	return appendXYForPoint(xys, ls.StartPoint())
}

func appendXYsForPolygon(xys []XY, poly Polygon) []XY {
	xys = appendXYForLineString(xys, poly.ExteriorRing())
	n := poly.NumInteriorRings()
	for i := 0; i < n; i++ {
		xys = appendXYForLineString(xys, poly.InteriorRingN(i))
	}
	return xys
}

func appendComponentPoints(xys []XY, g Geometry) []XY {
	switch g.Type() {
	case TypePoint:
		return appendXYForPoint(xys, g.MustAsPoint())
	case TypeMultiPoint:
		mp := g.MustAsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			xys = appendXYForPoint(xys, mp.PointN(i))
		}
		return xys
	case TypeLineString:
		ls := g.MustAsLineString()
		return appendXYForLineString(xys, ls)
	case TypeMultiLineString:
		mls := g.MustAsMultiLineString()
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			ls := mls.LineStringN(i)
			xys = appendXYForLineString(xys, ls)
		}
		return xys
	case TypePolygon:
		poly := g.MustAsPolygon()
		return appendXYsForPolygon(xys, poly)
	case TypeMultiPolygon:
		mp := g.MustAsMultiPolygon()
		n := mp.NumPolygons()
		for i := 0; i < n; i++ {
			poly := mp.PolygonN(i)
			xys = appendXYsForPolygon(xys, poly)
		}
		return xys
	case TypeGeometryCollection:
		gc := g.MustAsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			xys = appendComponentPoints(xys, gc.GeometryN(i))
		}
		return xys
	default:
		panic(fmt.Sprintf("unknown geometry type: %v", g.Type()))
	}
}
