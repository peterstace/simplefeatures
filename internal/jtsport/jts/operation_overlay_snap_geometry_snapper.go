package jts

import (
	"math"
	"sort"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const operationOverlaySnap_geometrySnapper_SNAP_PRECISION_FACTOR = 1e-9

// OperationOverlaySnap_GeometrySnapper snaps the vertices and segments of a
// Geometry to another Geometry's vertices. A snap distance tolerance is used to
// control where snapping is performed. Snapping one geometry to another can
// improve robustness for overlay operations by eliminating nearly-coincident
// edges (which cause problems during noding and intersection calculation). It
// can also be used to eliminate artifacts such as narrow slivers, spikes and
// gores.
//
// Too much snapping can result in invalid topology being created, so the number
// and location of snapped vertices is decided using heuristics to determine
// when it is safe to snap. This can result in some potential snaps being
// omitted, however.
type OperationOverlaySnap_GeometrySnapper struct {
	srcGeom *Geom_Geometry
}

// OperationOverlaySnap_GeometrySnapper_ComputeOverlaySnapTolerance estimates the
// snap tolerance for a Geometry, taking into account its precision model.
func OperationOverlaySnap_GeometrySnapper_ComputeOverlaySnapTolerance(g *Geom_Geometry) float64 {
	snapTolerance := OperationOverlaySnap_GeometrySnapper_ComputeSizeBasedSnapTolerance(g)

	// Overlay is carried out in the precision model of the two inputs. If this
	// precision model is of type FIXED, then the snap tolerance must reflect
	// the precision grid size. Specifically, the snap tolerance should be at
	// least the distance from a corner of a precision grid cell to the centre
	// point of the cell.
	pm := g.GetPrecisionModel()
	if pm.GetType() == Geom_PrecisionModel_Fixed {
		fixedSnapTol := (1 / pm.GetScale()) * 2 / 1.415
		if fixedSnapTol > snapTolerance {
			snapTolerance = fixedSnapTol
		}
	}
	return snapTolerance
}

// OperationOverlaySnap_GeometrySnapper_ComputeSizeBasedSnapTolerance computes
// a size-based snap tolerance for a geometry.
func OperationOverlaySnap_GeometrySnapper_ComputeSizeBasedSnapTolerance(g *Geom_Geometry) float64 {
	env := g.GetEnvelopeInternal()
	minDimension := math.Min(env.GetHeight(), env.GetWidth())
	return minDimension * operationOverlaySnap_geometrySnapper_SNAP_PRECISION_FACTOR
}

// OperationOverlaySnap_GeometrySnapper_ComputeOverlaySnapToleranceFromTwo
// computes the snap tolerance for two geometries.
func OperationOverlaySnap_GeometrySnapper_ComputeOverlaySnapToleranceFromTwo(g0, g1 *Geom_Geometry) float64 {
	return math.Min(
		OperationOverlaySnap_GeometrySnapper_ComputeOverlaySnapTolerance(g0),
		OperationOverlaySnap_GeometrySnapper_ComputeOverlaySnapTolerance(g1),
	)
}

// OperationOverlaySnap_GeometrySnapper_Snap snaps two geometries together with
// a given tolerance.
func OperationOverlaySnap_GeometrySnapper_Snap(g0, g1 *Geom_Geometry, snapTolerance float64) []*Geom_Geometry {
	snapGeom := make([]*Geom_Geometry, 2)
	snapper0 := OperationOverlaySnap_NewGeometrySnapper(g0)
	snapGeom[0] = snapper0.SnapTo(g1, snapTolerance)

	// Snap the second geometry to the snapped first geometry (this strategy
	// minimizes the number of possible different points in the result).
	snapper1 := OperationOverlaySnap_NewGeometrySnapper(g1)
	snapGeom[1] = snapper1.SnapTo(snapGeom[0], snapTolerance)

	return snapGeom
}

// OperationOverlaySnap_GeometrySnapper_SnapToSelf snaps a geometry to itself.
// Allows optionally cleaning the result to ensure it is topologically valid
// (which fixes issues such as topology collapses in polygonal inputs).
//
// Snapping a geometry to itself can remove artifacts such as very narrow
// slivers, gores and spikes.
func OperationOverlaySnap_GeometrySnapper_SnapToSelf(geom *Geom_Geometry, snapTolerance float64, cleanResult bool) *Geom_Geometry {
	snapper0 := OperationOverlaySnap_NewGeometrySnapper(geom)
	return snapper0.SnapToSelf(snapTolerance, cleanResult)
}

// OperationOverlaySnap_NewGeometrySnapper creates a new snapper acting on the
// given geometry.
func OperationOverlaySnap_NewGeometrySnapper(srcGeom *Geom_Geometry) *OperationOverlaySnap_GeometrySnapper {
	return &OperationOverlaySnap_GeometrySnapper{
		srcGeom: srcGeom,
	}
}

// SnapTo snaps the vertices in the component LineStrings of the source geometry
// to the vertices of the given snap geometry.
func (gs *OperationOverlaySnap_GeometrySnapper) SnapTo(snapGeom *Geom_Geometry, snapTolerance float64) *Geom_Geometry {
	snapPts := gs.extractTargetCoordinates(snapGeom)

	snapTrans := operationOverlaySnap_newSnapTransformer(snapTolerance, snapPts)
	return snapTrans.Transform(gs.srcGeom)
}

// SnapToSelf snaps the vertices in the component LineStrings of the source
// geometry to the vertices of the same geometry. Allows optionally cleaning the
// result to ensure it is topologically valid (which fixes issues such as
// topology collapses in polygonal inputs).
func (gs *OperationOverlaySnap_GeometrySnapper) SnapToSelf(snapTolerance float64, cleanResult bool) *Geom_Geometry {
	snapPts := gs.extractTargetCoordinates(gs.srcGeom)

	snapTrans := operationOverlaySnap_newSnapTransformerSelfSnap(snapTolerance, snapPts, true)
	snappedGeom := snapTrans.Transform(gs.srcGeom)
	result := snappedGeom
	if cleanResult && java.InstanceOf[Geom_Polygonal](snappedGeom) {
		// TODO: use better cleaning approach.
		result = snappedGeom.Buffer(0)
	}
	return result
}

func (gs *OperationOverlaySnap_GeometrySnapper) extractTargetCoordinates(g *Geom_Geometry) []*Geom_Coordinate {
	// Use a map to track unique coordinates.
	ptSet := make(map[operationOverlaySnap_coordKey]*Geom_Coordinate)
	pts := g.GetCoordinates()
	for i := 0; i < len(pts); i++ {
		key := operationOverlaySnap_coordKey{x: pts[i].X, y: pts[i].Y}
		if _, exists := ptSet[key]; !exists {
			ptSet[key] = pts[i]
		}
	}

	// Convert map to slice and sort (TreeSet in Java maintains sorted order).
	result := make([]*Geom_Coordinate, 0, len(ptSet))
	for _, coord := range ptSet {
		result = append(result, coord)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CompareTo(result[j]) < 0
	})
	return result
}

type operationOverlaySnap_coordKey struct {
	x, y float64
}

func (gs *OperationOverlaySnap_GeometrySnapper) computeSnapTolerance(ringPts []*Geom_Coordinate) float64 {
	minSegLen := gs.computeMinimumSegmentLength(ringPts)
	// Use a small percentage of this to be safe.
	return minSegLen / 10
}

func (gs *OperationOverlaySnap_GeometrySnapper) computeMinimumSegmentLength(pts []*Geom_Coordinate) float64 {
	minSegLen := math.MaxFloat64
	for i := 0; i < len(pts)-1; i++ {
		segLen := pts[i].Distance(pts[i+1])
		if segLen < minSegLen {
			minSegLen = segLen
		}
	}
	return minSegLen
}

// operationOverlaySnap_SnapTransformer is a GeometryTransformer that snaps
// coordinates to a set of snap points.
type operationOverlaySnap_SnapTransformer struct {
	*GeomUtil_GeometryTransformer
	child         java.Polymorphic
	snapTolerance float64
	snapPts       []*Geom_Coordinate
	isSelfSnap    bool
}

func (st *operationOverlaySnap_SnapTransformer) GetChild() java.Polymorphic {
	return st.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (st *operationOverlaySnap_SnapTransformer) GetParent() java.Polymorphic {
	return st.GeomUtil_GeometryTransformer
}

func operationOverlaySnap_newSnapTransformer(snapTolerance float64, snapPts []*Geom_Coordinate) *operationOverlaySnap_SnapTransformer {
	return operationOverlaySnap_newSnapTransformerSelfSnap(snapTolerance, snapPts, false)
}

func operationOverlaySnap_newSnapTransformerSelfSnap(snapTolerance float64, snapPts []*Geom_Coordinate, isSelfSnap bool) *operationOverlaySnap_SnapTransformer {
	base := GeomUtil_NewGeometryTransformer()
	st := &operationOverlaySnap_SnapTransformer{
		GeomUtil_GeometryTransformer: base,
		snapTolerance:               snapTolerance,
		snapPts:                     snapPts,
		isSelfSnap:                  isSelfSnap,
	}
	base.child = st
	return st
}

// TransformCoordinates_BODY overrides the parent implementation to snap
// coordinates.
func (st *operationOverlaySnap_SnapTransformer) TransformCoordinates_BODY(coords Geom_CoordinateSequence, parent *Geom_Geometry) Geom_CoordinateSequence {
	srcPts := coords.ToCoordinateArray()
	newPts := st.snapLine(srcPts, st.snapPts)
	return st.factory.GetCoordinateSequenceFactory().CreateFromCoordinates(newPts)
}

func (st *operationOverlaySnap_SnapTransformer) snapLine(srcPts, snapPts []*Geom_Coordinate) []*Geom_Coordinate {
	snapper := OperationOverlaySnap_NewLineStringSnapperFromCoordinates(srcPts, st.snapTolerance)
	snapper.SetAllowSnappingToSourceVertices(st.isSelfSnap)
	return snapper.SnapTo(snapPts)
}
