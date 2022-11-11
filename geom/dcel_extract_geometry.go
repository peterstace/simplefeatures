package geom

import "fmt"

// extractGeometry converts the DECL into a Geometry that represents it.
func (d *doublyConnectedEdgeList) extractGeometry(include func([2]bool) bool) (Geometry, error) {
	areals, err := d.extractPolygons(include)
	if err != nil {
		return Geometry{}, err
	}
	linears, err := d.extractLineStrings(include)
	if err != nil {
		return Geometry{}, err
	}
	points, err := d.extractPoints(include)
	if err != nil {
		return Geometry{}, err
	}

	switch {
	case len(areals) > 0 && len(linears) == 0 && len(points) == 0:
		if len(areals) == 1 {
			return areals[0].AsGeometry(), nil
		}
		mp, err := NewMultiPolygon(areals)
		if err != nil {
			return Geometry{}, wrap(err, "could not extract areal geometry from DCEL")
		}
		return mp.AsGeometry(), nil
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
		poly, err := NewPolygon(rings)
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
		for i := 0; i < e.seq.Length()-1; i++ {
			xy := e.seq.GetXY(i)
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

func (d *doublyConnectedEdgeList) extractLineStrings(include func([2]bool) bool) ([]LineString, error) {
	var lss []LineString
	for _, e := range d.halfEdges {
		if shouldExtractLine(e, include) {
			e.extracted = true
			e.twin.extracted = true
			e.origin.extracted = true
			e.twin.origin.extracted = true

			ls, err := NewLineString(e.seq)
			if err != nil {
				return nil, err
			}
			lss = append(lss, ls)
		}
	}
	return lss, nil
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
func (d *doublyConnectedEdgeList) extractPoints(include func([2]bool) bool) ([]Point, error) {
	var pts []Point
	for _, vert := range d.vertices {
		if include(vert.inSet) && !vert.extracted {
			vert.extracted = true
			pt, err := vert.coords.AsPoint()
			if err != nil {
				return nil, err
			}
			pts = append(pts, pt)
		}
	}
	return pts, nil
}
