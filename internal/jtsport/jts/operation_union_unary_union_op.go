package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationUnion_UnaryUnionOp unions a Collection of Geometrys or a single
// Geometry (which may be a GeometryCollection) together. By using this
// special-purpose operation over a collection of geometries it is possible to
// take advantage of various optimizations to improve performance.
// Heterogeneous GeometryCollections are fully supported.
//
// The result obeys the following contract:
//   - Unioning a set of Polygons has the effect of merging the areas (i.e. the
//     same effect as iteratively unioning all individual polygons together).
//   - Unioning a set of LineStrings has the effect of noding and dissolving the
//     input linework. In this context "fully noded" means that there will be an
//     endpoint or node in the result for every endpoint or line segment crossing
//     in the input. "Dissolved" means that any duplicate (i.e. coincident) line
//     segments or portions of line segments will be reduced to a single line
//     segment in the result. This is consistent with the semantics of the
//     Geometry.Union operation. If merged linework is required, the LineMerger
//     class can be used.
//   - Unioning a set of Points has the effect of merging all identical points
//     (producing a set with no duplicates).
//
// UnaryUnion always operates on the individual components of MultiGeometries.
// So it is possible to use it to "clean" invalid self-intersecting
// MultiPolygons (although the polygon components must all still be individually
// valid.)
type OperationUnion_UnaryUnionOp struct {
	geomFact      *Geom_GeometryFactory
	extracter     *operationUnion_InputExtracter
	unionFunction OperationUnion_UnionStrategy
}

// OperationUnion_UnaryUnionOp_UnionCollection computes the geometric union of a
// Collection of Geometrys.
//
// Returns the union of the geometries, or nil if the input is empty.
func OperationUnion_UnaryUnionOp_UnionCollection(geoms []*Geom_Geometry) *Geom_Geometry {
	op := OperationUnion_NewUnaryUnionOpFromCollection(geoms)
	return op.Union()
}

// OperationUnion_UnaryUnionOp_UnionCollectionWithFactory computes the geometric
// union of a Collection of Geometrys.
//
// If no input geometries were provided but a GeometryFactory was provided, an
// empty GeometryCollection is returned.
//
// Returns the union of the geometries, or an empty GEOMETRYCOLLECTION.
func OperationUnion_UnaryUnionOp_UnionCollectionWithFactory(geoms []*Geom_Geometry, geomFact *Geom_GeometryFactory) *Geom_Geometry {
	op := OperationUnion_NewUnaryUnionOpFromCollectionWithFactory(geoms, geomFact)
	return op.Union()
}

// OperationUnion_UnaryUnionOp_Union constructs a unary union operation for a
// Geometry (which may be a GeometryCollection).
//
// Returns the union of the elements of the geometry or an empty
// GEOMETRYCOLLECTION.
func OperationUnion_UnaryUnionOp_Union(geom *Geom_Geometry) *Geom_Geometry {
	op := OperationUnion_NewUnaryUnionOpFromGeometry(geom)
	return op.Union()
}

// OperationUnion_NewUnaryUnionOpFromCollectionWithFactory constructs a unary
// union operation for a Collection of Geometrys.
func OperationUnion_NewUnaryUnionOpFromCollectionWithFactory(geoms []*Geom_Geometry, geomFact *Geom_GeometryFactory) *OperationUnion_UnaryUnionOp {
	op := &OperationUnion_UnaryUnionOp{
		geomFact:      geomFact,
		unionFunction: OperationUnion_CascadedPolygonUnion_ClassicUnion,
	}
	op.extractCollection(geoms)
	return op
}

// OperationUnion_NewUnaryUnionOpFromCollection constructs a unary union
// operation for a Collection of Geometrys, using the GeometryFactory of the
// input geometries.
func OperationUnion_NewUnaryUnionOpFromCollection(geoms []*Geom_Geometry) *OperationUnion_UnaryUnionOp {
	op := &OperationUnion_UnaryUnionOp{
		unionFunction: OperationUnion_CascadedPolygonUnion_ClassicUnion,
	}
	op.extractCollection(geoms)
	return op
}

// OperationUnion_NewUnaryUnionOpFromGeometry constructs a unary union operation
// for a Geometry (which may be a GeometryCollection).
func OperationUnion_NewUnaryUnionOpFromGeometry(geom *Geom_Geometry) *OperationUnion_UnaryUnionOp {
	op := &OperationUnion_UnaryUnionOp{
		unionFunction: OperationUnion_CascadedPolygonUnion_ClassicUnion,
	}
	op.extractGeometry(geom)
	return op
}

// SetUnionFunction sets the union strategy to use.
func (op *OperationUnion_UnaryUnionOp) SetUnionFunction(unionFun OperationUnion_UnionStrategy) {
	op.unionFunction = unionFun
}

func (op *OperationUnion_UnaryUnionOp) extractCollection(geoms []*Geom_Geometry) {
	op.extracter = operationUnion_InputExtracter_ExtractFromCollection(geoms)
}

func (op *OperationUnion_UnaryUnionOp) extractGeometry(geom *Geom_Geometry) {
	op.extracter = operationUnion_InputExtracter_Extract(geom)
}

// Union gets the union of the input geometries.
//
// The result of empty input is determined as follows:
//  1. If the input is empty and a dimension can be determined (i.e. an empty
//     geometry is present), an empty atomic geometry of that dimension is
//     returned.
//  2. If no input geometries were provided but a GeometryFactory was provided,
//     an empty GeometryCollection is returned.
//  3. Otherwise, the return value is nil.
//
// Returns a Geometry containing the union, or an empty atomic geometry, or an
// empty GEOMETRYCOLLECTION, or nil if no GeometryFactory was provided.
func (op *OperationUnion_UnaryUnionOp) Union() *Geom_Geometry {
	if op.geomFact == nil {
		op.geomFact = op.extracter.GetFactory()
	}

	// Case 3.
	if op.geomFact == nil {
		return nil
	}

	// Case 1 & 2.
	if op.extracter.IsEmpty() {
		return op.geomFact.CreateEmpty(op.extracter.GetDimension())
	}

	points := op.extracter.GetExtract(0)
	lines := op.extracter.GetExtract(1)
	polygons := op.extracter.GetExtract(2)

	// For points and lines, only a single union operation is required, since
	// the OGC model allows self-intersecting MultiPoint and MultiLineStrings.
	// This is not the case for polygons, so Cascaded Union is required.
	var unionPoints *Geom_Geometry
	if len(points) > 0 {
		ptGeom := op.geomFact.BuildGeometry(points)
		unionPoints = op.unionNoOpt(ptGeom)
	}

	var unionLines *Geom_Geometry
	if len(lines) > 0 {
		lineGeom := op.geomFact.BuildGeometry(lines)
		unionLines = op.unionNoOpt(lineGeom)
	}

	var unionPolygons *Geom_Geometry
	if len(polygons) > 0 {
		unionPolygons = OperationUnion_CascadedPolygonUnion_UnionWithStrategy(polygons, op.unionFunction)
	}

	// Performing two unions is somewhat inefficient, but is mitigated by
	// unioning lines and points first.
	unionLA := op.unionWithNull(unionLines, unionPolygons)
	var union *Geom_Geometry
	if unionPoints == nil {
		union = unionLA
	} else if unionLA == nil {
		union = unionPoints
	} else {
		union = OperationUnion_PointGeometryUnion_Union(java.GetLeaf(unionPoints).(Geom_Puntal), unionLA)
	}

	if union == nil {
		return op.geomFact.CreateGeometryCollection().Geom_Geometry
	}

	return union
}

// unionWithNull computes the union of two geometries, either or both of which
// may be nil.
func (op *OperationUnion_UnaryUnionOp) unionWithNull(g0, g1 *Geom_Geometry) *Geom_Geometry {
	if g0 == nil && g1 == nil {
		return nil
	}
	if g1 == nil {
		return g0
	}
	if g0 == nil {
		return g1
	}
	return g0.Union(g1)
}

// unionNoOpt computes a unary union with no extra optimization, and no
// short-circuiting. Due to the way the overlay operations are implemented, this
// is still efficient in the case of linear and puntal geometries. Uses robust
// version of overlay operation to ensure identical behaviour to the
// Union(Geometry) operation.
func (op *OperationUnion_UnaryUnionOp) unionNoOpt(g0 *Geom_Geometry) *Geom_Geometry {
	empty := op.geomFact.CreatePoint()
	return op.unionFunction.Union(g0, empty.Geom_Geometry)
}
