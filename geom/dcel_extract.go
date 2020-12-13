package geom

import "fmt"

// extractGeometry converts the DECL into a Geometry that represents it.
func (d *doublyConnectedEdgeList) extractGeometry(include func(uint8) bool) (Geometry, error) {
	areals, err := d.extractPolygons(include)
	if err != nil {
		return Geometry{}, err
	}
	linears, err := d.extractLineStrings(include)
	if err != nil {
		return Geometry{}, err
	}
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

func (d *doublyConnectedEdgeList) extractPolygons(include func(uint8) bool) ([]Polygon, error) {
	var polys []Polygon
	for _, face := range d.faces {
		// Skip any faces not selected to be include in the output geometry, or
		// any faces already extracted.
		if !include(face.label) || face.label&extracted != 0 {
			continue
		}

		// Find all faces that make up the polygon.
		facesInPoly := findFacesMakingPolygon(include, face)

		// Extract the Polygon boundaries from the edges forming the face cycles.
		var rings []LineString
		seen := make(map[*halfEdgeRecord]bool)
		for f := range facesInPoly {
			f.label |= extracted
			forEachEdge(f.cycle, func(edge *halfEdgeRecord) {

				// Mark all edges and vertices intersecting with the polygon as
				// being extracted.  This will prevent them being considered
				// during linear and point geometry extraction.
				edge.edgeLabel |= extracted
				edge.twin.edgeLabel |= extracted
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
			return nil, err
		}
		polys = append(polys, poly)
	}
	return polys, nil
}

func extractPolygonBoundary(faceSet map[*faceRecord]bool, start *halfEdgeRecord, seen map[*halfEdgeRecord]bool) Sequence {
	var coords []float64
	e := start
	for {
		seen[e] = true
		xy := e.origin.coords
		coords = append(coords, xy.X, xy.Y)
		for _, xy := range e.intermediate {
			coords = append(coords, xy.X, xy.Y)
		}

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

func (d *doublyConnectedEdgeList) extractLineStrings(include func(uint8) bool) ([]LineString, error) {
	var lss []LineString
	for _, e := range d.halfEdges {
		if shouldExtractLine(e, include) {
			e.edgeLabel |= extracted
			e.twin.edgeLabel |= extracted
			e.origin.label |= extracted
			e.twin.origin.label |= extracted

			coords := make([]float64, 4+2*len(e.intermediate))
			coords[0] = e.origin.coords.X
			coords[1] = e.origin.coords.Y
			for i, xy := range e.intermediate {
				coords[2+2*i] = xy.X
				coords[3+2*i] = xy.Y
			}
			coords[len(coords)-2] = e.twin.origin.coords.X
			coords[len(coords)-1] = e.twin.origin.coords.Y

			seq := NewSequence(coords, DimXY)
			ls, err := NewLineString(seq)
			if err != nil {
				return nil, err
			}
			lss = append(lss, ls)
		}
	}
	return lss, nil
}

func shouldExtractLine(e *halfEdgeRecord, include func(uint8) bool) bool {
	return (e.edgeLabel&extracted == 0) && include(e.edgeLabel) && !include(e.incident.label) && !include(e.twin.incident.label)
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
