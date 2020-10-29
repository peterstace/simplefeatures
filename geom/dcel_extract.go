package geom

import "fmt"

// extractGeometry converts the DECL into a Geometry that represents it.
func (d *doublyConnectedEdgeList) extractGeometry(include func(uint8) bool) (Geometry, error) {
	areals := d.extractPolygons(include)
	linears := d.extractLineStrings(include)
	points := d.extractPoints(include)

	switch {
	case len(areals) > 0 && len(linears) == 0 && len(points) == 0:
		if len(areals) == 1 {
			return areals[0].AsGeometry(), nil
		}
		mp, err := NewMultiPolygonFromPolygons(areals)
		if err != nil {
			return Geometry{}, fmt.Errorf("could not extract areal geometry from DCEL: %v", err)
		}
		return mp.AsGeometry(), nil
	case len(areals) == 0 && len(linears) > 0 && len(points) == 0:
		if len(linears) == 1 {
			return linears[0].AsGeometry(), nil
		}
		return NewMultiLineStringFromLineStrings(linears).AsGeometry(), nil
	case len(areals) == 0 && len(linears) == 0 && len(points) > 0:
		if len(points) == 1 {
			return NewPointFromXY(points[0]).AsGeometry(), nil
		}
		coords := make([]float64, 2*len(points))
		for i, xy := range points {
			coords[i*2+0] = xy.X
			coords[i*2+1] = xy.Y
		}
		return NewMultiPoint(NewSequence(coords, DimXY)).AsGeometry(), nil
	default:
		geoms := make([]Geometry, 0, len(areals)+len(linears)+len(points))
		for _, poly := range areals {
			geoms = append(geoms, poly.AsGeometry())
		}
		for _, ls := range linears {
			geoms = append(geoms, ls.AsGeometry())
		}
		for _, xy := range points {
			geoms = append(geoms, NewPointFromXY(xy).AsGeometry())
		}
		return NewGeometryCollection(geoms).AsGeometry(), nil
	}
}

func (d *doublyConnectedEdgeList) extractPolygons(include func(uint8) bool) []Polygon {
	var polys []Polygon
	for _, face := range d.faces {
		if !include(face.label) {
			continue
		}

		// Mark vertices internal to the face as already extracted so that
		// they're ignored during point extraction.
		for _, vert := range face.internalVertices {
			vert.label |= extracted
		}

		if (face.label & extracted) != 0 {
			continue
		}

		// Find all faces that make up the polygon.
		facesInPoly := findFacesMakingPolygon(include, face)

		// Find all edge cycles incident to the faces. Edges in these cycles
		// are are candidates to be part of the Polygon boundary.
		var components []*halfEdgeRecord
		for f := range facesInPoly {
			f.label |= extracted
			if cmp := f.outerComponent; cmp != nil {
				components = append(components, cmp)
			}
			components = append(components, f.innerComponents...)
		}

		// Extract the Polygon boundaries from the candidate edges.
		var rings []LineString
		seen := make(map[*halfEdgeRecord]bool)
		for _, cmp := range components {
			forEachEdge(cmp, func(edge *halfEdgeRecord) {

				// Mark all edges and vertices intersecting with the polygon as
				// being extracted.  This will prevent them being considered
				// during linear and point geometry extraction.
				edge.label |= extracted
				edge.twin.label |= extracted
				edge.origin.label |= extracted

				if seen[edge] {
					return
				}
				if include(edge.twin.incident.label) {
					// Adjacent face is in the polygon, so this edge cannot be part
					// of the boundary.
					seen[edge] = true
					return
				}
				seq := extractPolygonBoundary(facesInPoly, edge, seen)
				ring, err := NewLineString(seq)
				if err != nil {
					panic(fmt.Sprintf("could not create LineString: %v", err))
				}
				rings = append(rings, ring)
			})
		}

		// Construct the polygon.
		orderCCWRingFirst(rings)
		poly, err := NewPolygonFromRings(rings)
		if err != nil {
			panic(fmt.Sprintf("could not create Polygon: %v", err))
		}
		polys = append(polys, poly)
	}
	return polys
}

func extractPolygonBoundary(faceSet map[*faceRecord]bool, start *halfEdgeRecord, seen map[*halfEdgeRecord]bool) Sequence {
	var coords []float64
	e := start
	for {
		seen[e] = true
		xy := e.origin.coords
		coords = append(coords, xy.X, xy.Y)

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
	coords = append(coords, coords[:2]...)
	return NewSequence(coords, DimXY)
}

// findFacesMakingPolygon finds all faces that belong to the polygon that
// contains the start face (according to the given inclusion criteria).
func findFacesMakingPolygon(include func(uint8) bool, start *faceRecord) map[*faceRecord]bool {
	expanded := make(map[*faceRecord]bool)
	toExpand := make(map[*faceRecord]bool)
	toExpand[start] = true
	pop := func() *faceRecord {
		for f := range toExpand {
			delete(toExpand, f)
			return f
		}
		panic("could not pop")
	}

	for len(toExpand) > 0 {
		popped := pop()
		adj := adjacentFaces(popped)
		expanded[popped] = true
		for _, f := range adj {
			if !include(f.label) {
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

// orderCCWRingFirst reorders rings such that if it contains at least one CCW
// ring, then a CCW ring is the first element.
func orderCCWRingFirst(rings []LineString) {
	for i, r := range rings {
		if ccw := signedAreaOfLinearRing(r, nil) > 0; ccw {
			rings[i], rings[0] = rings[0], rings[i]
			return
		}
	}
}

// TODO: Line extraction isn't working too well at the moment. It's currently
// extracting each line individually, which isn't intended. It might be better
// to return a []line here, and then construct back into LineString and
// MultiLineString as a separate logical step since it seems tricky to do
// inline.

func (d *doublyConnectedEdgeList) extractLineStrings(include func(uint8) bool) []LineString {
	var lss []LineString
	for _, e := range d.halfEdges {
		if shouldExtractLine(e, include) {
			ls := extractLineString(e, include)
			lss = append(lss, ls)
		}
	}
	return lss
}

func extractLineString(e *halfEdgeRecord, include func(uint8) bool) LineString {
	u := e.origin.coords
	coords := []float64{u.X, u.Y}

	for {
		v := e.next.origin.coords
		coords = append(coords, v.X, v.Y)
		e.label |= extracted
		e.twin.label |= extracted
		e.origin.label |= extracted
		e.twin.origin.label |= extracted

		e = nextNoBranch(e, include)
		if e == nil {
			break
		}
	}

	seq := NewSequence(coords, DimXY)
	ls, err := NewLineString(seq)
	if err != nil {
		// Shouldn't ever happen, since we have at least one edge.
		panic(fmt.Sprintf("could not construct line string using %v: %v", coords, err))
	}
	return ls
}

func shouldExtractLine(e *halfEdgeRecord, include func(uint8) bool) bool {
	return (e.label&extracted == 0) && include(e.label) && !include(e.incident.label) && !include(e.twin.incident.label)
}

// nextNoBranch checks to see if the given edge has multiple next edges that it
// could use for linear extraction. If there are multiple edges, then nil is
// returned (this is called a 'branch'). If there is just one possible next
// edge, then that next edge is returned.
func nextNoBranch(edge *halfEdgeRecord, include func(uint8) bool) *halfEdgeRecord {
	e := edge.next
	var nextEdge *halfEdgeRecord

	// Find the first next edge.
	for {
		if e == edge.twin {
			// There are no linear branches that could be extracted.
			return nil
		}
		if shouldExtractLine(e, include) {
			nextEdge = e
			break
		}
		e = e.twin.next
	}

	// Check to see if there are additional next edges (i.e. a branch scenario).
	for {
		if e == edge.twin {
			// There is no branching.
			return nextEdge
		}
		if shouldExtractLine(e, include) {
			// There is branching, so indicate this by returning nil.
			return nil
		}
		e = e.twin.next
	}
}

// extractPoints extracts any vertices in the DCEL that should be part of the
// output geometry, but aren't yet represented as part of any previously
// extracted geometries.
func (d *doublyConnectedEdgeList) extractPoints(include func(uint8) bool) []XY {
	var xys []XY
	for _, vert := range d.vertices {
		if include(vert.label) && vert.label&extracted == 0 {
			vert.label |= extracted
			xys = append(xys, vert.coords)
		}
	}
	return xys
}
