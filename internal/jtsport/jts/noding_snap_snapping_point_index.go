package jts

// NodingSnap_SnappingPointIndex is an index providing fast creation and lookup
// of snap points.
type NodingSnap_SnappingPointIndex struct {
	snapTolerance  float64
	snapPointIndex *IndexKdtree_KdTree
}

// NodingSnap_NewSnappingPointIndex creates a snap point index using a
// specified distance tolerance.
func NodingSnap_NewSnappingPointIndex(snapTolerance float64) *NodingSnap_SnappingPointIndex {
	return &NodingSnap_SnappingPointIndex{
		snapTolerance:  snapTolerance,
		snapPointIndex: IndexKdtree_NewKdTreeWithTolerance(snapTolerance),
	}
}

// Snap snaps a coordinate to an existing snap point, if it is within the snap
// tolerance distance. Otherwise adds the coordinate to the snap point index.
func (spi *NodingSnap_SnappingPointIndex) Snap(p *Geom_Coordinate) *Geom_Coordinate {
	// Inserting the coordinate snaps it to any existing one within tolerance,
	// or adds it if not.
	node := spi.snapPointIndex.Insert(p)
	return node.GetCoordinate()
}

// GetTolerance gets the snapping tolerance value for the index.
func (spi *NodingSnap_SnappingPointIndex) GetTolerance() float64 {
	return spi.snapTolerance
}

// Depth computes the depth of the index tree.
func (spi *NodingSnap_SnappingPointIndex) Depth() int {
	return spi.snapPointIndex.Depth()
}
