package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationUnion_CascadedPolygonUnion_ClassicUnion is a union strategy that
// uses the classic JTS SnapIfNeededOverlayOp.
var OperationUnion_CascadedPolygonUnion_ClassicUnion OperationUnion_UnionStrategy = &classicUnionStrategy{}

type classicUnionStrategy struct{}

func (s *classicUnionStrategy) Union(g0, g1 *Geom_Geometry) *Geom_Geometry {
	// Try SnapIfNeededOverlayOp first, fall back to OverlayNGRobust on
	// TopologyException (matching Java's behavior).
	result, topologyErr := classicUnionStrategy_trySnapUnion(g0, g1)
	if topologyErr != nil {
		return OperationOverlayng_OverlayNGRobust_Overlay(g0, g1, OperationOverlayng_OverlayNG_UNION)
	}
	return result
}

// classicUnionStrategy_trySnapUnion attempts to union using
// SnapIfNeededOverlayOp. Returns the result and nil on success, or nil and the
// TopologyException on failure.
func classicUnionStrategy_trySnapUnion(g0, g1 *Geom_Geometry) (result *Geom_Geometry, topologyErr *Geom_TopologyException) {
	defer func() {
		if r := recover(); r != nil {
			if te, ok := r.(*Geom_TopologyException); ok {
				topologyErr = te
			} else {
				// Re-panic for non-TopologyException panics.
				panic(r)
			}
		}
	}()
	result = OperationOverlaySnap_SnapIfNeededOverlayOp_Union(g0, g1)
	return result, nil
}

func (s *classicUnionStrategy) IsFloatingPrecision() bool {
	return true
}

// OperationUnion_CascadedPolygonUnion provides an efficient method of unioning
// a collection of Polygonal geometries. The geometries are indexed using a
// spatial index, and unioned recursively in index order. For geometries with a
// high degree of overlap, this has the effect of reducing the number of
// vertices early in the process, which increases speed and robustness.
//
// This algorithm is faster and more robust than the simple iterated approach of
// repeatedly unioning each polygon to a result geometry.
type OperationUnion_CascadedPolygonUnion struct {
	inputPolys     []*Geom_Geometry
	geomFactory    *Geom_GeometryFactory
	unionFun       OperationUnion_UnionStrategy
	countRemainder int
	countInput     int
}

// operationUnion_CascadedPolygonUnion_STRtreeNodeCapacity is the effectiveness
// of the index is somewhat sensitive to the node capacity. Testing indicates
// that a smaller capacity is better. For an STRtree, 4 is probably a good
// number (since this produces 2x2 "squares").
const operationUnion_CascadedPolygonUnion_STRtreeNodeCapacity = 4

// OperationUnion_CascadedPolygonUnion_Union computes the union of a collection
// of Polygonal Geometries.
func OperationUnion_CascadedPolygonUnion_Union(polys []*Geom_Geometry) *Geom_Geometry {
	op := OperationUnion_NewCascadedPolygonUnion(polys)
	return op.Union()
}

// OperationUnion_CascadedPolygonUnion_UnionWithStrategy computes the union of a
// collection of Polygonal Geometries using the given union strategy.
func OperationUnion_CascadedPolygonUnion_UnionWithStrategy(polys []*Geom_Geometry, unionFun OperationUnion_UnionStrategy) *Geom_Geometry {
	op := OperationUnion_NewCascadedPolygonUnionWithStrategy(polys, unionFun)
	return op.Union()
}

// OperationUnion_NewCascadedPolygonUnion creates a new instance to union the
// given collection of Geometries.
func OperationUnion_NewCascadedPolygonUnion(polys []*Geom_Geometry) *OperationUnion_CascadedPolygonUnion {
	return OperationUnion_NewCascadedPolygonUnionWithStrategy(polys, OperationUnion_CascadedPolygonUnion_ClassicUnion)
}

// OperationUnion_NewCascadedPolygonUnionWithStrategy creates a new instance to
// union the given collection of Geometries using the given union strategy.
func OperationUnion_NewCascadedPolygonUnionWithStrategy(polys []*Geom_Geometry, unionFun OperationUnion_UnionStrategy) *OperationUnion_CascadedPolygonUnion {
	inputPolys := polys
	// Guard against nil input.
	if inputPolys == nil {
		inputPolys = make([]*Geom_Geometry, 0)
	}
	return &OperationUnion_CascadedPolygonUnion{
		inputPolys:     inputPolys,
		unionFun:       unionFun,
		countInput:     len(inputPolys),
		countRemainder: len(inputPolys),
	}
}

// Union computes the union of the input geometries. This method discards the
// input geometries as they are processed. In many input cases this reduces the
// memory retained as the operation proceeds. Optimal memory usage is achieved
// by disposing of the original input collection before calling this method.
//
// Returns the union of the input geometries, or nil if no input geometries were
// provided.
//
// Panics if this method is called more than once.
func (cpu *OperationUnion_CascadedPolygonUnion) Union() *Geom_Geometry {
	if cpu.inputPolys == nil {
		panic("union() method cannot be called twice")
	}
	if len(cpu.inputPolys) == 0 {
		return nil
	}
	cpu.geomFactory = cpu.inputPolys[0].GetFactory()

	// A spatial index to organize the collection into groups of close
	// geometries. This makes unioning more efficient, since vertices are more
	// likely to be eliminated on each round.
	index := IndexStrtree_NewSTRtreeWithCapacity(operationUnion_CascadedPolygonUnion_STRtreeNodeCapacity)
	for _, item := range cpu.inputPolys {
		index.Insert(item.GetEnvelopeInternal(), item)
	}
	// To avoiding holding memory remove references to the input geometries.
	cpu.inputPolys = nil

	itemTree := index.ItemsTree()
	unionAll := cpu.unionTree(itemTree)
	return unionAll
}

func (cpu *OperationUnion_CascadedPolygonUnion) unionTree(geomTree []any) *Geom_Geometry {
	// Recursively unions all subtrees in the list into single geometries.
	// The result is a list of Geometries only.
	geoms := cpu.reduceToGeometries(geomTree)
	union := cpu.binaryUnion(geoms)
	return union
}

// binaryUnion unions a list of geometries by treating the list as a flattened
// binary tree, and performing a cascaded union on the tree.
func (cpu *OperationUnion_CascadedPolygonUnion) binaryUnion(geoms []*Geom_Geometry) *Geom_Geometry {
	return cpu.binaryUnionRange(geoms, 0, len(geoms))
}

// binaryUnionRange unions a section of a list using a recursive binary union
// on each half of the section.
func (cpu *OperationUnion_CascadedPolygonUnion) binaryUnionRange(geoms []*Geom_Geometry, start, end int) *Geom_Geometry {
	if end-start <= 1 {
		g0 := operationUnion_CascadedPolygonUnion_getGeometry(geoms, start)
		return cpu.unionSafe(g0, nil)
	} else if end-start == 2 {
		return cpu.unionSafe(operationUnion_CascadedPolygonUnion_getGeometry(geoms, start), operationUnion_CascadedPolygonUnion_getGeometry(geoms, start+1))
	} else {
		// Recurse on both halves of the list.
		mid := (end + start) / 2
		g0 := cpu.binaryUnionRange(geoms, start, mid)
		g1 := cpu.binaryUnionRange(geoms, mid, end)
		return cpu.unionSafe(g0, g1)
	}
}

// operationUnion_CascadedPolygonUnion_getGeometry gets the element at a given
// list index, or nil if the index is out of range.
func operationUnion_CascadedPolygonUnion_getGeometry(list []*Geom_Geometry, index int) *Geom_Geometry {
	if index >= len(list) {
		return nil
	}
	return list[index]
}

// reduceToGeometries reduces a tree of geometries to a list of geometries by
// recursively unioning the subtrees in the list.
func (cpu *OperationUnion_CascadedPolygonUnion) reduceToGeometries(geomTree []any) []*Geom_Geometry {
	geoms := make([]*Geom_Geometry, 0)
	for _, o := range geomTree {
		var geom *Geom_Geometry
		switch v := o.(type) {
		case []any:
			geom = cpu.unionTree(v)
		case *Geom_Geometry:
			geom = v
		}
		geoms = append(geoms, geom)
	}
	return geoms
}

// unionSafe computes the union of two geometries, either or both of which may
// be nil.
func (cpu *OperationUnion_CascadedPolygonUnion) unionSafe(g0, g1 *Geom_Geometry) *Geom_Geometry {
	if g0 == nil && g1 == nil {
		return nil
	}
	if g0 == nil {
		return g1.Copy()
	}
	if g1 == nil {
		return g0.Copy()
	}

	cpu.countRemainder--

	union := cpu.unionActual(g0, g1)
	return union
}

// unionActual encapsulates the actual unioning of two polygonal geometries.
func (cpu *OperationUnion_CascadedPolygonUnion) unionActual(g0, g1 *Geom_Geometry) *Geom_Geometry {
	union := cpu.unionFun.Union(g0, g1)
	unionPoly := operationUnion_CascadedPolygonUnion_restrictToPolygons(union)
	return unionPoly
}

// operationUnion_CascadedPolygonUnion_restrictToPolygons computes a Geometry
// containing only Polygonal components. Extracts the Polygons from the input
// and returns them as an appropriate Polygonal geometry.
//
// If the input is already Polygonal, it is returned unchanged.
//
// A particular use case is to filter out non-polygonal components returned from
// an overlay operation.
func operationUnion_CascadedPolygonUnion_restrictToPolygons(g *Geom_Geometry) *Geom_Geometry {
	if java.InstanceOf[Geom_Polygonal](g) {
		return g
	}
	polygons := GeomUtil_PolygonExtracter_GetPolygons(g)
	if len(polygons) == 1 {
		return polygons[0].Geom_Geometry
	}
	return g.GetFactory().CreateMultiPolygonFromPolygons(polygons).Geom_GeometryCollection.Geom_Geometry
}
