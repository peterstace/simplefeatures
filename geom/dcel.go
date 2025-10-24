package geom

func newDCELFromGeometries(a, b Geometry) *doublyConnectedEdgeList {
	a, b, ghosts := prepareGeometriesForDCEL(a, b)
	return newDCELFromRenodedGeometries(a, b, ghosts)
}

// prepareGeometriesForDCEL pre-processes the input geometries (A and B) such
// that they can be used to create a DCEL. An additional "ghost"
// MultiLineString is also returned, which provides the appropriate connections
// such that A and B (when combined together) are fully connected.
func prepareGeometriesForDCEL(a, b Geometry) (Geometry, Geometry, MultiLineString) {
	// Phase 1: Initial renoding (without ghosts). This ensures that the two
	// input geometries only interact at control points before ghost edge
	// construction.
	a, b, _ = reNodeGeometries(a, b, MultiLineString{})

	// Phase 2: Create ghost edges to connect disjoint components.
	ghosts := createGhosts(a, b)

	// Phase 3: Final renoding (with ghosts). This handles the case where ghost
	// edges had to split input geometry edges.
	if !ghosts.IsEmpty() {
		a, b, ghosts = reNodeGeometries(a, b, ghosts)
	}

	return a, b, ghosts
}

func newDCELFromRenodedGeometries(a, b Geometry, ghosts MultiLineString) *doublyConnectedEdgeList {
	interactions := findInteractionPoints([]Geometry{a, b, ghosts.AsGeometry()})

	dcel := newDCEL()
	dcel.addVertices(interactions)
	dcel.addGhosts(ghosts, interactions)
	dcel.addGeometry(a, operandA, interactions)
	dcel.addGeometry(b, operandB, interactions)

	dcel.fixVertices()
	dcel.assignFaces()
	dcel.populateInSetLabels()

	return dcel
}

func newDCEL() *doublyConnectedEdgeList {
	return &doublyConnectedEdgeList{
		faces:     nil,
		halfEdges: make(map[[2]XY]*halfEdgeRecord),
		vertices:  make(map[XY]*vertexRecord),
	}
}

type doublyConnectedEdgeList struct {
	faces     []*faceRecord // only populated in the overlay
	halfEdges map[[2]XY]*halfEdgeRecord
	vertices  map[XY]*vertexRecord
}

type faceRecord struct {
	cycle *halfEdgeRecord

	// inSet encodes whether this face is part of the input geometry for each
	// operand.
	inSet [2]bool

	extracted bool
}

type halfEdgeRecord struct {
	origin     *vertexRecord
	twin       *halfEdgeRecord
	incident   *faceRecord // only populated in the overlay
	next, prev *halfEdgeRecord
	seq        Sequence

	// srcEdge encodes whether or not this edge is explicitly appears as part
	// of the input geometries.
	srcEdge [2]bool

	// srcFace encodes whether or not this edge explicitly borders onto a face
	// in the input geometries.
	srcFace [2]bool

	// inSet encodes whether or not this edge is (explicitly or implicitly)
	// part of the input geometry for each operand.
	inSet [2]bool

	extracted bool
}

type vertexRecord struct {
	coords    XY
	incidents map[*halfEdgeRecord]struct{}

	// src encodes whether on not this vertex explicitly appears in the input
	// geometries.
	src [2]bool

	// inSet encodes whether or not this vertex is part of each input geometry
	// (although it might not be explicitly encoded there).
	inSet [2]bool

	locations [2]location
	extracted bool
}

func forEachEdgeInCycle(start *halfEdgeRecord, fn func(*halfEdgeRecord)) {
	e := start
	for {
		fn(e)
		e = e.next
		if e == start {
			break
		}
	}
}

// operand represents either the first (A) or second (B) geometry in a binary
// operation (such as Union or Covers).
type operand int

const (
	operandA operand = 0
	operandB operand = 1
)

func forEachOperand(fn func(operand operand)) {
	fn(operandA)
	fn(operandB)
}

type location struct {
	interior bool
	boundary bool
}
