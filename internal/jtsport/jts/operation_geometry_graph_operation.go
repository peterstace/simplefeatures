package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Operation_GeometryGraphOperation is the base class for operations that
// require GeometryGraphs.
type Operation_GeometryGraphOperation struct {
	child java.Polymorphic

	li                   *Algorithm_LineIntersector
	resultPrecisionModel *Geom_PrecisionModel

	// arg contains the operation args in an array so they can be accessed by
	// index.
	arg []*Geomgraph_GeometryGraph
}

// GetChild returns the immediate child in the type hierarchy chain.
func (ggo *Operation_GeometryGraphOperation) GetChild() java.Polymorphic {
	return ggo.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (ggo *Operation_GeometryGraphOperation) GetParent() java.Polymorphic {
	return nil
}

// Operation_NewGeometryGraphOperation creates a new GeometryGraphOperation for
// two geometries.
func Operation_NewGeometryGraphOperation(g0, g1 *Geom_Geometry) *Operation_GeometryGraphOperation {
	return Operation_NewGeometryGraphOperationWithBoundaryNodeRule(g0, g1, Algorithm_BoundaryNodeRule_OGC_SFS_BOUNDARY_RULE)
}

// Operation_NewGeometryGraphOperationWithBoundaryNodeRule creates a new
// GeometryGraphOperation for two geometries with a custom boundary node rule.
func Operation_NewGeometryGraphOperationWithBoundaryNodeRule(g0, g1 *Geom_Geometry, boundaryNodeRule Algorithm_BoundaryNodeRule) *Operation_GeometryGraphOperation {
	ggo := &Operation_GeometryGraphOperation{
		li: Algorithm_NewRobustLineIntersector().Algorithm_LineIntersector,
	}

	// Use the most precise model for the result.
	if g0.GetPrecisionModel().CompareTo(g1.GetPrecisionModel()) >= 0 {
		ggo.setComputationPrecision(g0.GetPrecisionModel())
	} else {
		ggo.setComputationPrecision(g1.GetPrecisionModel())
	}

	ggo.arg = make([]*Geomgraph_GeometryGraph, 2)
	ggo.arg[0] = Geomgraph_NewGeometryGraphWithBoundaryNodeRule(0, g0, boundaryNodeRule)
	ggo.arg[1] = Geomgraph_NewGeometryGraphWithBoundaryNodeRule(1, g1, boundaryNodeRule)
	return ggo
}

// Operation_NewGeometryGraphOperationSingle creates a new
// GeometryGraphOperation for a single geometry.
func Operation_NewGeometryGraphOperationSingle(g0 *Geom_Geometry) *Operation_GeometryGraphOperation {
	ggo := &Operation_GeometryGraphOperation{
		li: Algorithm_NewRobustLineIntersector().Algorithm_LineIntersector,
	}
	ggo.setComputationPrecision(g0.GetPrecisionModel())

	ggo.arg = make([]*Geomgraph_GeometryGraph, 1)
	ggo.arg[0] = Geomgraph_NewGeometryGraph(0, g0)
	return ggo
}

// GetArgGeometry returns the argument geometry at the given index.
func (ggo *Operation_GeometryGraphOperation) GetArgGeometry(i int) *Geom_Geometry {
	return ggo.arg[i].GetGeometry()
}

func (ggo *Operation_GeometryGraphOperation) setComputationPrecision(pm *Geom_PrecisionModel) {
	ggo.resultPrecisionModel = pm
	ggo.li.SetPrecisionModel(ggo.resultPrecisionModel)
}
