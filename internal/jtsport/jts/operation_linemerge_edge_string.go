package jts

// OperationLinemerge_EdgeString is a sequence of LineMergeDirectedEdges forming one of the lines
// that will be output by the line-merging process.
type OperationLinemerge_EdgeString struct {
	factory       *Geom_GeometryFactory
	directedEdges []*OperationLinemerge_LineMergeDirectedEdge
	coordinates   []*Geom_Coordinate
}

// OperationLinemerge_NewEdgeString constructs an EdgeString with the given factory used to
// convert this EdgeString to a LineString.
func OperationLinemerge_NewEdgeString(factory *Geom_GeometryFactory) *OperationLinemerge_EdgeString {
	return &OperationLinemerge_EdgeString{
		factory:       factory,
		directedEdges: make([]*OperationLinemerge_LineMergeDirectedEdge, 0),
	}
}

// Add adds a directed edge which is known to form part of this line.
func (es *OperationLinemerge_EdgeString) Add(directedEdge *OperationLinemerge_LineMergeDirectedEdge) {
	es.directedEdges = append(es.directedEdges, directedEdge)
}

// getCoordinates returns the coordinates of this EdgeString.
func (es *OperationLinemerge_EdgeString) getCoordinates() []*Geom_Coordinate {
	if es.coordinates == nil {
		forwardDirectedEdges := 0
		reverseDirectedEdges := 0
		coordinateList := Geom_NewCoordinateList()
		for _, directedEdge := range es.directedEdges {
			if directedEdge.GetEdgeDirection() {
				forwardDirectedEdges++
			} else {
				reverseDirectedEdges++
			}
			lineMergeEdge := directedEdge.GetEdge().GetChild().(*OperationLinemerge_LineMergeEdge)
			coordinateList.AddCoordinatesWithDirection(
				lineMergeEdge.GetLine().GetCoordinates(),
				false,
				directedEdge.GetEdgeDirection(),
			)
		}
		es.coordinates = coordinateList.ToCoordinateArray()
		if reverseDirectedEdges > forwardDirectedEdges {
			Geom_CoordinateArrays_Reverse(es.coordinates)
		}
	}
	return es.coordinates
}

// ToLineString converts this EdgeString into a LineString.
func (es *OperationLinemerge_EdgeString) ToLineString() *Geom_LineString {
	return es.factory.CreateLineStringFromCoordinates(es.getCoordinates())
}
