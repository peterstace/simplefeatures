package geom

// findInteractionPoints finds the interaction points (including
// self-interaction points) between a list of geometries.
//
// Assumptions:
//
//  1. Input geometries are correctly noded with respect to each other.
//
//  2. Input geometries don't have any repeated coordinates (e.g. in areal
//     rings or linear elements).
func findInteractionPoints(gs []Geometry) map[XY]struct{} {
	var sizeHint int
	for _, g := range gs {
		sizeHint += g.controlPoints()
	}
	interactions := make(map[XY]struct{}, sizeHint)

	// adjacents tracks the next and previous points relative to a middle point
	// for linear elements (i.e. the points adjacent to a middle point). It is
	// used to differentiate the cases where linear elements overlap (in which
	// case there ISN'T an interaction point) and cases where they are crossing
	// over each other (in which case there IS an interaction point).
	adjacents := make(map[XY]xyPair, sizeHint)

	for _, g := range gs {
		addGeometryInteractions(g, adjacents, interactions)
	}
	return interactions
}

// xyPair is a container for a pair of XYs. The semantics of the points aren't
// implied by this type itself (user of this type is to BYO semantics).
type xyPair struct {
	first, second XY
}

func addGeometryInteractions(g Geometry, adjacents map[XY]xyPair, interactions map[XY]struct{}) {
	switch g.Type() {
	case TypePoint:
		addPointInteractions(g.MustAsPoint(), interactions)
	case TypeMultiPoint:
		addMultiPointInteractions(g.MustAsMultiPoint(), interactions)
	case TypeLineString:
		addLineStringInteractions(g.MustAsLineString(), adjacents, interactions)
	case TypeMultiLineString:
		addMultiLineStringInteractions(g.MustAsMultiLineString(), adjacents, interactions)
	case TypePolygon:
		addMultiLineStringInteractions(g.MustAsPolygon().Boundary(), adjacents, interactions)
	case TypeMultiPolygon:
		addMultiLineStringInteractions(g.MustAsMultiPolygon().Boundary(), adjacents, interactions)
	case TypeGeometryCollection:
		addGeometryCollectionInteractions(g.MustAsGeometryCollection(), adjacents, interactions)
	default:
		panic("unknown geometry: " + g.Type().String())
	}
}

func addLineStringInteractions(ls LineString, adjacents map[XY]xyPair, interactions map[XY]struct{}) {
	if xy, ok := ls.StartPoint().XY(); ok {
		interactions[xy] = struct{}{}
	}
	if xy, ok := ls.EndPoint().XY(); ok {
		interactions[xy] = struct{}{}
	}

	seq := ls.Coordinates()
	n := seq.Length()
	for i := 1; i+1 < n; i++ {
		prev := seq.GetXY(i - 1)
		curr := seq.GetXY(i)
		next := seq.GetXY(i + 1)

		if prev == next {
			// LineString loops back on itself, so the reversal point is the
			// interaction point.
			interactions[curr] = struct{}{}
			continue
		}

		adj := xyPair{prev, next}
		if adj.second.Less(adj.first) {
			// Canonicalise the pair, since we don't care about directionality.
			adj.first, adj.second = adj.second, adj.first
		}

		xy := seq.GetXY(i)
		existing, ok := adjacents[xy]
		if ok && existing != adj {
			interactions[xy] = struct{}{}
		}
		if !ok {
			adjacents[xy] = adj
		}
	}
}

func addMultiLineStringInteractions(mls MultiLineString, adjacents map[XY]xyPair, interactions map[XY]struct{}) {
	for i := 0; i < mls.NumLineStrings(); i++ {
		ls := mls.LineStringN(i)
		addLineStringInteractions(ls, adjacents, interactions)
	}
}

func addPointInteractions(pt Point, interactions map[XY]struct{}) {
	if xy, ok := pt.XY(); ok {
		interactions[xy] = struct{}{}
	}
}

func addMultiPointInteractions(mp MultiPoint, interactions map[XY]struct{}) {
	n := mp.NumPoints()
	for i := 0; i < n; i++ {
		xy, ok := mp.PointN(i).XY()
		if ok {
			interactions[xy] = struct{}{}
		}
	}
}

func addGeometryCollectionInteractions(gc GeometryCollection, adjacents map[XY]xyPair, interactions map[XY]struct{}) {
	n := gc.NumGeometries()
	for i := 0; i < n; i++ {
		g := gc.GeometryN(i)
		addGeometryInteractions(g, adjacents, interactions)
	}
}
