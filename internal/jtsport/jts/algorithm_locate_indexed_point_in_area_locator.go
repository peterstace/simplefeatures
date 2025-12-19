package jts

import (
	"math"
	"sync"
)

// Compile-time interface check.
var _ AlgorithmLocate_PointOnGeometryLocator = (*AlgorithmLocate_IndexedPointInAreaLocator)(nil)

// AlgorithmLocate_IndexedPointInAreaLocator determines the Location of
// Coordinates relative to an areal geometry, using indexing for efficiency.
// This algorithm is suitable for use in cases where many points will be tested
// against a given area.
//
// The Location is computed precisely, in that points located on the geometry
// boundary or segments will return Geom_Location_Boundary.
//
// Polygonal and LinearRing geometries are supported.
//
// The index is lazy-loaded, which allows creating instances even if they are
// not used.
//
// Thread-safe and immutable.
type AlgorithmLocate_IndexedPointInAreaLocator struct {
	geom  *Geom_Geometry
	index *algorithmLocate_IntervalIndexedGeometry
	mu    sync.Mutex
}

// IsAlgorithmLocate_PointOnGeometryLocator is a marker method for the interface.
func (l *AlgorithmLocate_IndexedPointInAreaLocator) IsAlgorithmLocate_PointOnGeometryLocator() {}

// AlgorithmLocate_NewIndexedPointInAreaLocator creates a new locator for a
// given Geometry. Geometries containing Polygons and LinearRing geometries are
// supported.
func AlgorithmLocate_NewIndexedPointInAreaLocator(g *Geom_Geometry) *AlgorithmLocate_IndexedPointInAreaLocator {
	return &AlgorithmLocate_IndexedPointInAreaLocator{
		geom: g,
	}
}

// Locate determines the Location of a point in an areal Geometry.
func (l *AlgorithmLocate_IndexedPointInAreaLocator) Locate(p *Geom_Coordinate) int {
	// Avoid calling synchronized method improves performance.
	if l.index == nil {
		l.createIndex()
	}

	rcc := Algorithm_NewRayCrossingCounter(p)

	visitor := algorithmLocate_newSegmentVisitor(rcc)
	l.index.query(p.Y, p.Y, visitor)

	return rcc.GetLocation()
}

// createIndex creates the indexed geometry, creating it if necessary.
func (l *AlgorithmLocate_IndexedPointInAreaLocator) createIndex() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.index == nil {
		l.index = algorithmLocate_newIntervalIndexedGeometry(l.geom)
		// No need to hold onto geom.
		l.geom = nil
	}
}

// algorithmLocate_SegmentVisitor is a visitor for segments in the index.
type algorithmLocate_SegmentVisitor struct {
	counter *Algorithm_RayCrossingCounter
}

var _ Index_ItemVisitor = (*algorithmLocate_SegmentVisitor)(nil)

func (v *algorithmLocate_SegmentVisitor) IsIndex_ItemVisitor() {}

func algorithmLocate_newSegmentVisitor(counter *Algorithm_RayCrossingCounter) *algorithmLocate_SegmentVisitor {
	return &algorithmLocate_SegmentVisitor{counter: counter}
}

func (v *algorithmLocate_SegmentVisitor) VisitItem(item any) {
	seg := item.(*Geom_LineSegment)
	v.counter.CountSegment(seg.GetCoordinate(0), seg.GetCoordinate(1))
}

// algorithmLocate_IntervalIndexedGeometry is an internal class for indexing a
// geometry by its segments' Y-coordinate intervals.
type algorithmLocate_IntervalIndexedGeometry struct {
	isEmpty bool
	index   *IndexIntervalrtree_SortedPackedIntervalRTree
}

func algorithmLocate_newIntervalIndexedGeometry(geom *Geom_Geometry) *algorithmLocate_IntervalIndexedGeometry {
	iig := &algorithmLocate_IntervalIndexedGeometry{
		index: IndexIntervalrtree_NewSortedPackedIntervalRTree(),
	}
	if geom.IsEmpty() {
		iig.isEmpty = true
	} else {
		iig.isEmpty = false
		iig.init(geom)
	}
	return iig
}

func (iig *algorithmLocate_IntervalIndexedGeometry) init(geom *Geom_Geometry) {
	lines := GeomUtil_LinearComponentExtracter_GetLines(geom)
	for _, line := range lines {
		// Only include rings of Polygons or LinearRings.
		if !line.IsClosed() {
			continue
		}
		pts := line.GetCoordinates()
		iig.addLine(pts)
	}
}

func (iig *algorithmLocate_IntervalIndexedGeometry) addLine(pts []*Geom_Coordinate) {
	for i := 1; i < len(pts); i++ {
		seg := Geom_NewLineSegmentFromCoordinates(pts[i-1], pts[i])
		minY := math.Min(seg.P0.Y, seg.P1.Y)
		maxY := math.Max(seg.P0.Y, seg.P1.Y)
		iig.index.Insert(minY, maxY, seg)
	}
}

func (iig *algorithmLocate_IntervalIndexedGeometry) queryToList(minY, maxY float64) []any {
	if iig.isEmpty {
		return nil
	}
	visitor := Index_NewArrayListVisitor()
	iig.index.Query(minY, maxY, visitor)
	return visitor.GetItems()
}

func (iig *algorithmLocate_IntervalIndexedGeometry) query(minY, maxY float64, visitor *algorithmLocate_SegmentVisitor) {
	if iig.isEmpty {
		return
	}
	iig.index.Query(minY, maxY, visitor)
}
