package geom

import (
	"errors"
	"sort"
)

// extractGeometry converts the DECL into a Geometry that represents it.
func (d *doublyConnectedEdgeList) extractGeometry(include func([2]bool) bool) (Geometry, error) {
	areals, err := d.extractPolygons(include)
	if err != nil {
		return Geometry{}, err
	}
	linears := d.extractLineStrings(include)
	points := d.extractPoints(include)

	switch {
	case len(areals) > 0 && len(linears) == 0 && len(points) == 0:
		if len(areals) == 1 {
			return areals[0].AsGeometry(), nil
		}
		return NewMultiPolygon(areals).AsGeometry(), nil
	case len(areals) == 0 && len(linears) > 0 && len(points) == 0:
		if len(linears) == 1 {
			return linears[0].AsGeometry(), nil
		}
		return NewMultiLineString(linears).AsGeometry(), nil
	case len(areals) == 0 && len(linears) == 0 && len(points) > 0:
		if len(points) == 1 {
			return points[0].AsGeometry(), nil
		}
		return NewMultiPoint(points).AsGeometry(), nil
	default:
		geoms := make([]Geometry, 0, len(areals)+len(linears)+len(points))
		for _, poly := range areals {
			geoms = append(geoms, poly.AsGeometry())
		}
		for _, ls := range linears {
			geoms = append(geoms, ls.AsGeometry())
		}
		for _, pt := range points {
			geoms = append(geoms, pt.AsGeometry())
		}
		return NewGeometryCollection(geoms).AsGeometry(), nil
	}
}

func (d *doublyConnectedEdgeList) extractPolygons(include func([2]bool) bool) ([]Polygon, error) {
	var polys []Polygon
	for _, face := range d.faces {
		// Skip any faces not selected to be include in the output geometry, or
		// any faces already extracted.
		if !include(face.inSet) || face.extracted {
			continue
		}

		// Find all faces that make up the polygon.
		facesInPoly := findFacesMakingPolygon(include, face)

		// Extract the Polygon boundaries from the edges forming the face cycles.
		var rings []LineString
		seen := make(map[*halfEdgeRecord]bool)
		for f := range facesInPoly {
			f.extracted = true
			forEachEdgeInCycle(f.cycle, func(edge *halfEdgeRecord) {
				// Mark all edges and vertices intersecting with the polygon as
				// being extracted.  This will prevent them being considered
				// during linear and point geometry extraction.
				edge.extracted = true
				edge.twin.extracted = true
				edge.origin.extracted = true

				if seen[edge] {
					return
				}
				if include(edge.twin.incident.inSet) {
					// Adjacent face is in the polygon, so this edge cannot be part
					// of the boundary.
					seen[edge] = true
					return
				}
				ring := extractPolygonRing(facesInPoly, edge, seen)
				rings = append(rings, ring)
			})
		}

		if len(rings) == 0 {
			return nil, errors.New("no rings to extract")
		}

		// Construct the polygon.
		orderPolygonRings(rings)
		polys = append(polys, NewPolygon(rings))
	}

	sort.Slice(polys, func(i, j int) bool {
		seqI := polys[i].ExteriorRing().Coordinates()
		seqJ := polys[j].ExteriorRing().Coordinates()
		return seqI.less(seqJ)
	})
	return polys, nil
}

func extractPolygonRing(faceSet map[*faceRecord]bool, start *halfEdgeRecord, seen map[*halfEdgeRecord]bool) LineString {
	var seqs []Sequence
	e := start
	for {
		seen[e] = true
		seqs = append(seqs, e.seq)

		// Sweep through the edges around the vertex (in a counter-clockwise
		// order) until we find the next edge that is part of the polygon
		// boundary.
		e = e.twin.prev.twin
		for !faceSet[e.incident] {
			e = e.prev.twin
		}

		if e == start {
			break
		}
	}

	// Reorder seqs such that one of them comes first in a deterministic
	// manner. The sequences still have to form a ring, so they're rotated
	// rather than sorted.
	minI := 0
	for i := range seqs {
		if seqs[i].less(seqs[minI]) {
			minI = i
		}
	}
	rotateSeqs(seqs, len(seqs)-minI)

	return NewLineString(buildRingSequence(seqs))
}

func buildRingSequence(seqs []Sequence) Sequence {
	// Calculate desired capacity.
	var capacity int
	for _, seq := range seqs {
		capacity += 2 * seq.Length()
	}
	capacity -= 2 * len(seqs) // Account for shared point at start/end of each seq.
	capacity += 2             // Account for repeated start/end point of ring.

	// Build concatenated sequence.
	coords := make([]float64, 0, capacity)
	for _, seq := range seqs {
		coords = seq.appendAllPoints(coords)
		coords = coords[:len(coords)-2]
	}
	coords = append(coords, coords[:2]...)
	seq := NewSequence(coords, DimXY)
	seq.assertNoUnusedCapacity()
	return seq
}

// rotateSeqs moves each sequence rotRight places to the right, wrapping around
// the end of the slice to the start of the slice.
//
// TODO: use generics for this once we depend on Go 1.19.
func rotateSeqs(seqs []Sequence, rotRight int) {
	if rotRight == 0 || rotRight == len(seqs) {
		return // Nothing to do (optimisation).
	}
	reverseSeqs(seqs)
	reverseSeqs(seqs[:rotRight])
	reverseSeqs(seqs[rotRight:])
}

// reverseSeqs reverses the order of the input slice.
//
// TODO: use generics for this once we depend on Go 1.19.
func reverseSeqs(seqs []Sequence) {
	n := len(seqs)
	for i := 0; i < n/2; i++ {
		j := n - i - 1
		seqs[i], seqs[j] = seqs[j], seqs[i]
	}
}

// findFacesMakingPolygon finds all faces that belong to the polygon that
// contains the start face (according to the given inclusion criteria).
func findFacesMakingPolygon(include func([2]bool) bool, start *faceRecord) map[*faceRecord]bool {
	expanded := make(map[*faceRecord]bool)
	toExpand := make(map[*faceRecord]bool)
	toExpand[start] = true

	for len(toExpand) > 0 {
		var popped *faceRecord
		for f := range toExpand {
			delete(toExpand, f)
			popped = f
			break
		}

		adj := adjacentFaces(popped)
		expanded[popped] = true
		for _, f := range adj {
			if !include(f.inSet) {
				continue
			}
			if expanded[f] {
				continue
			}
			if toExpand[f] {
				continue
			}
			toExpand[f] = true
		}
	}
	return expanded
}

// orderPolygonRings reorders rings such that the outer (CCW) ring comes first,
// and any inner (CW) rings are ordered afterwards in a stable way.
func orderPolygonRings(rings []LineString) {
	for i, r := range rings {
		if ccw := signedAreaOfLinearRing(r, nil) > 0; ccw {
			rings[i], rings[0] = rings[0], rings[i]
			break
		}
	}
	inners := rings[1:]
	sort.Slice(inners, func(i, j int) bool {
		seqI := inners[i].Coordinates()
		seqJ := inners[j].Coordinates()
		return seqI.less(seqJ)
	})
}

func (d *doublyConnectedEdgeList) extractLineStrings(include func([2]bool) bool) []LineString {
	var lss []LineString
	for _, e := range d.halfEdges {
		if shouldExtractLine(e, include) {
			if e.twin.seq.less(e.seq) {
				e = e.twin // Extract in deterministic order.
			}
			e.extracted = true
			e.twin.extracted = true
			e.origin.extracted = true
			e.twin.origin.extracted = true

			lss = append(lss, NewLineString(e.seq))
		}
	}
	sort.Slice(lss, func(i, j int) bool {
		seqI := lss[i].Coordinates()
		seqJ := lss[j].Coordinates()
		return seqI.less(seqJ)
	})
	return lss
}

func shouldExtractLine(e *halfEdgeRecord, include func([2]bool) bool) bool {
	return !e.extracted &&
		include(e.inSet) &&
		!include(e.incident.inSet) &&
		!include(e.twin.incident.inSet)
}

// extractPoints extracts any vertices in the DCEL that should be part of the
// output geometry, but aren't yet represented as part of any previously
// extracted geometries.
func (d *doublyConnectedEdgeList) extractPoints(include func([2]bool) bool) []Point {
	xys := make([]XY, 0, len(d.vertices))
	for _, vert := range d.vertices {
		if include(vert.inSet) && !vert.extracted {
			vert.extracted = true
			xys = append(xys, vert.coords)
		}
	}

	sort.Slice(xys, func(i, j int) bool {
		return xys[i].Less(xys[j])
	})

	pts := make([]Point, 0, len(xys))
	for _, xy := range xys {
		pts = append(pts, xy.AsPoint())
	}
	return pts
}
