package jts

import "testing"

func TestOverlayGraph_Triangle(t *testing.T) {
	line1 := createLine(0, 0, 10, 10)
	line2 := createLine(10, 10, 0, 10)
	line3 := createLine(0, 10, 0, 0)

	graph := createGraphFromLines(line1, line2, line3)

	e1 := findEdge(graph, 0, 0, 10, 10)
	e2 := findEdge(graph, 10, 10, 0, 10)
	e3 := findEdge(graph, 0, 10, 0, 0)

	checkNodeValid(t, e1)
	checkNodeValid(t, e2)
	checkNodeValid(t, e3)

	checkNext(t, e1, e2)
	checkNext(t, e2, e3)
	checkNext(t, e3, e1)

	e1sym := findEdge(graph, 10, 10, 0, 0)
	e2sym := findEdge(graph, 0, 10, 10, 10)
	e3sym := findEdge(graph, 0, 0, 0, 10)

	if e1sym != e1.SymOE() {
		t.Errorf("e1sym != e1.SymOE()")
	}
	if e2sym != e2.SymOE() {
		t.Errorf("e2sym != e2.SymOE()")
	}
	if e3sym != e3.SymOE() {
		t.Errorf("e3sym != e3.SymOE()")
	}

	checkNext(t, e1sym, e3sym)
	checkNext(t, e2sym, e1sym)
	checkNext(t, e3sym, e2sym)
}

func TestOverlayGraph_Star(t *testing.T) {
	graph := OperationOverlayng_NewOverlayGraph()

	e1 := addEdgeToGraph(graph, 5, 5, 0, 0)
	e2 := addEdgeToGraph(graph, 5, 5, 0, 9)
	e3 := addEdgeToGraph(graph, 5, 5, 9, 9)

	checkNodeValid(t, e1)

	checkNext(t, e1, e1.SymOE())
	checkNext(t, e2, e2.SymOE())
	checkNext(t, e3, e3.SymOE())

	checkPrev(t, e1, e2.SymOE())
	checkPrev(t, e2, e3.SymOE())
	checkPrev(t, e3, e1.SymOE())
}

// TestOverlayGraph_CCWAfterInserts tests edge sorting after inserts.
// This test produced an error using the old HalfEdge sorting algorithm
// (in HalfEdge.insert).
func TestOverlayGraph_CCWAfterInserts(t *testing.T) {
	e1 := createLine(50, 39, 35, 42, 37, 30)
	e2 := createLine(50, 39, 50, 60, 20, 60)
	e3 := createLine(50, 39, 68, 35)

	graph := createGraphFromLines(e1, e2, e3)
	node := graph.GetNodeEdge(Geom_NewCoordinateWithXY(50, 39))
	checkNodeValid(t, node)
}

func TestOverlayGraph_CCWAfterInserts2(t *testing.T) {
	e1 := createLine(50, 200, 0, 200)
	e2 := createLine(50, 200, 190, 50, 50, 50)
	e3 := createLine(50, 200, 200, 200, 0, 200)

	graph := createGraphFromLines(e1, e2, e3)
	node := graph.GetNodeEdge(Geom_NewCoordinateWithXY(50, 200))
	checkNodeValid(t, node)
}

func checkNext(t *testing.T, e, eNext *OperationOverlayng_OverlayEdge) {
	t.Helper()
	if e.NextOE() != eNext {
		t.Errorf("expected next edge to be %v, got %v", eNext, e.NextOE())
	}
}

func checkPrev(t *testing.T, e, ePrev *OperationOverlayng_OverlayEdge) {
	t.Helper()
	if e.PrevOE() != ePrev {
		t.Errorf("expected prev edge to be %v, got %v", ePrev, e.PrevOE())
	}
}

func checkNodeValid(t *testing.T, e *OperationOverlayng_OverlayEdge) {
	t.Helper()
	isNodeValid := e.IsEdgesSorted()
	if !isNodeValid {
		t.Errorf("found non-sorted edges around node %s", e.ToStringNode())
	}
}

func findEdge(graph *OperationOverlayng_OverlayGraph, orgx, orgy, destx, desty float64) *OperationOverlayng_OverlayEdge {
	edges := graph.GetEdges()
	for _, e := range edges {
		if isEdgeOrgDest(e, orgx, orgy, destx, desty) {
			return e
		}
		if isEdgeOrgDest(e.SymOE(), orgx, orgy, destx, desty) {
			return e.SymOE()
		}
	}
	return nil
}

func isEdgeOrgDest(e *OperationOverlayng_OverlayEdge, orgx, orgy, destx, desty float64) bool {
	if !isEqualCoord(e.Orig(), orgx, orgy) {
		return false
	}
	if !isEqualCoord(e.Dest(), destx, desty) {
		return false
	}
	return true
}

func isEqualCoord(p *Geom_Coordinate, x, y float64) bool {
	return p.GetX() == x && p.GetY() == y
}

func createGraphFromLines(edges ...[]*Geom_Coordinate) *OperationOverlayng_OverlayGraph {
	graph := OperationOverlayng_NewOverlayGraph()
	for _, e := range edges {
		graph.AddEdge(e, OperationOverlayng_NewOverlayLabel())
	}
	return graph
}

func addEdgeToGraph(graph *OperationOverlayng_OverlayGraph, x1, y1, x2, y2 float64) *OperationOverlayng_OverlayEdge {
	pts := []*Geom_Coordinate{
		Geom_NewCoordinateWithXY(x1, y1),
		Geom_NewCoordinateWithXY(x2, y2),
	}
	return graph.AddEdge(pts, OperationOverlayng_NewOverlayLabel())
}

func createLine(ord ...float64) []*Geom_Coordinate {
	return toCoordinates(ord)
}

func toCoordinates(ord []float64) []*Geom_Coordinate {
	pts := make([]*Geom_Coordinate, len(ord)/2)
	for i := range pts {
		pts[i] = Geom_NewCoordinateWithXY(ord[2*i], ord[2*i+1])
	}
	return pts
}
