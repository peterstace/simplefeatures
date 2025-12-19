package jts

import (
	"reflect"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const Geom_Geometry_TypeCodePoint = 0
const Geom_Geometry_TypeCodeMultiPoint = 1
const Geom_Geometry_TypeCodeLineString = 2
const Geom_Geometry_TypeCodeLinearRing = 3
const Geom_Geometry_TypeCodeMultiLineString = 4
const Geom_Geometry_TypeCodePolygon = 5
const Geom_Geometry_TypeCodeMultiPolygon = 6
const Geom_Geometry_TypeCodeGeometryCollection = 7

const Geom_Geometry_TypeNamePoint = "Point"
const Geom_Geometry_TypeNameMultiPoint = "MultiPoint"
const Geom_Geometry_TypeNameLineString = "LineString"
const Geom_Geometry_TypeNameLinearRing = "LinearRing"
const Geom_Geometry_TypeNameMultiLineString = "MultiLineString"
const Geom_Geometry_TypeNamePolygon = "Polygon"
const Geom_Geometry_TypeNameMultiPolygon = "MultiPolygon"
const Geom_Geometry_TypeNameGeometryCollection = "GeometryCollection"

var geom_Geometry_geometryChangedFilter Geom_GeometryComponentFilter = &geom_geometryChangedFilterImpl{}

type geom_geometryChangedFilterImpl struct{}

func (g *geom_geometryChangedFilterImpl) IsGeom_GeometryComponentFilter() {}

func (g *geom_geometryChangedFilterImpl) Filter(geom *Geom_Geometry) {
	geom.GeometryChangedAction()
}

type Geom_Geometry struct {
	child    java.Polymorphic
	envelope *Geom_Envelope
	factory  *Geom_GeometryFactory
	srid     int
	userData any
}

func Geom_NewGeometry(factory *Geom_GeometryFactory) *Geom_Geometry {
	return &Geom_Geometry{
		factory: factory,
		srid:    factory.GetSRID(),
	}
}

// GetChild returns the immediate child in the type hierarchy chain.
func (g *Geom_Geometry) GetChild() java.Polymorphic {
	return g.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (g *Geom_Geometry) GetParent() java.Polymorphic {
	return nil
}

func (g *Geom_Geometry) GetGeometryType() string {
	if impl, ok := java.GetLeaf(g).(interface{ GetGeometryType_BODY() string }); ok {
		return impl.GetGeometryType_BODY()
	}
	panic("abstract method called")
}

func Geom_Geometry_HasNonEmptyElements(geometries []*Geom_Geometry) bool {
	for i := range geometries {
		if !geometries[i].IsEmpty() {
			return true
		}
	}
	return false
}

func Geom_Geometry_HasNullElements(array []any) bool {
	for i := range array {
		if array[i] == nil {
			return true
		}
	}
	return false
}

func (g *Geom_Geometry) GetSRID() int {
	return g.srid
}

func (g *Geom_Geometry) SetSRID(srid int) {
	g.srid = srid
}

func (g *Geom_Geometry) GetFactory() *Geom_GeometryFactory {
	return g.factory
}

func (g *Geom_Geometry) GetUserData() any {
	return g.userData
}

func (g *Geom_Geometry) GetNumGeometries() int {
	if impl, ok := java.GetLeaf(g).(interface{ GetNumGeometries_BODY() int }); ok {
		return impl.GetNumGeometries_BODY()
	}
	return g.GetNumGeometries_BODY()
}

func (g *Geom_Geometry) GetNumGeometries_BODY() int {
	return 1
}

func (g *Geom_Geometry) GetGeometryN(n int) *Geom_Geometry {
	if impl, ok := java.GetLeaf(g).(interface{ GetGeometryN_BODY(int) *Geom_Geometry }); ok {
		return impl.GetGeometryN_BODY(n)
	}
	return g.GetGeometryN_BODY(n)
}

func (g *Geom_Geometry) GetGeometryN_BODY(n int) *Geom_Geometry {
	return g
}

func (g *Geom_Geometry) SetUserData(userData any) {
	g.userData = userData
}

func (g *Geom_Geometry) GetPrecisionModel() *Geom_PrecisionModel {
	return g.factory.GetPrecisionModel()
}

func (g *Geom_Geometry) GetCoordinate() *Geom_Coordinate {
	if impl, ok := java.GetLeaf(g).(interface{ GetCoordinate_BODY() *Geom_Coordinate }); ok {
		return impl.GetCoordinate_BODY()
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) GetCoordinates() []*Geom_Coordinate {
	if impl, ok := java.GetLeaf(g).(interface{ GetCoordinates_BODY() []*Geom_Coordinate }); ok {
		return impl.GetCoordinates_BODY()
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) GetNumPoints() int {
	if impl, ok := java.GetLeaf(g).(interface{ GetNumPoints_BODY() int }); ok {
		return impl.GetNumPoints_BODY()
	}
	panic("abstract method called")
}

// IsSimple tests whether this geometry is simple. The SFS definition of
// simplicity follows the general rule that a Geometry is simple if it has no
// points of self-tangency, self-intersection or other anomalous points.
func (g *Geom_Geometry) IsSimple() bool {
	return OperationValid_IsSimpleOp_IsSimple(g)
}

func (g *Geom_Geometry) IsValid() bool {
	return OperationValid_IsValidOp_IsValid(g)
}

func (g *Geom_Geometry) IsEmpty() bool {
	if impl, ok := java.GetLeaf(g).(interface{ IsEmpty_BODY() bool }); ok {
		return impl.IsEmpty_BODY()
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) Distance(other *Geom_Geometry) float64 {
	return OperationDistance_DistanceOp_Distance(g, other)
}

func (g *Geom_Geometry) IsWithinDistance(geom *Geom_Geometry, distance float64) bool {
	return OperationDistance_DistanceOp_IsWithinDistance(g, geom, distance)
}

func (g *Geom_Geometry) IsRectangle() bool {
	if impl, ok := java.GetLeaf(g).(interface{ IsRectangle_BODY() bool }); ok {
		return impl.IsRectangle_BODY()
	}
	return g.IsRectangle_BODY()
}

func (g *Geom_Geometry) IsRectangle_BODY() bool {
	return false
}

func (g *Geom_Geometry) GetArea() float64 {
	if impl, ok := java.GetLeaf(g).(interface{ GetArea_BODY() float64 }); ok {
		return impl.GetArea_BODY()
	}
	return g.GetArea_BODY()
}

func (g *Geom_Geometry) GetArea_BODY() float64 {
	return 0.0
}

func (g *Geom_Geometry) GetLength() float64 {
	if impl, ok := java.GetLeaf(g).(interface{ GetLength_BODY() float64 }); ok {
		return impl.GetLength_BODY()
	}
	return g.GetLength_BODY()
}

func (g *Geom_Geometry) GetLength_BODY() float64 {
	return 0.0
}

func (g *Geom_Geometry) GetCentroid() *Geom_Point {
	if g.IsEmpty() {
		return g.factory.CreatePoint()
	}
	centPt := Algorithm_Centroid_GetCentroid(g)
	return g.createPointFromInternalCoord(centPt, g)
}

func (g *Geom_Geometry) GetInteriorPoint() *Geom_Point {
	if g.IsEmpty() {
		return g.factory.CreatePoint()
	}
	pt := Algorithm_InteriorPoint_GetInteriorPoint(g)
	return g.createPointFromInternalCoord(pt, g)
}

func (g *Geom_Geometry) GetDimension() int {
	if impl, ok := java.GetLeaf(g).(interface{ GetDimension_BODY() int }); ok {
		return impl.GetDimension_BODY()
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) HasDimension(dim int) bool {
	if impl, ok := java.GetLeaf(g).(interface{ HasDimension_BODY(int) bool }); ok {
		return impl.HasDimension_BODY(dim)
	}
	return g.HasDimension_BODY(dim)
}

func (g *Geom_Geometry) HasDimension_BODY(dim int) bool {
	return dim == g.GetDimension()
}

func (g *Geom_Geometry) GetBoundary() *Geom_Geometry {
	if impl, ok := java.GetLeaf(g).(interface{ GetBoundary_BODY() *Geom_Geometry }); ok {
		return impl.GetBoundary_BODY()
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) GetBoundaryDimension() int {
	if impl, ok := java.GetLeaf(g).(interface{ GetBoundaryDimension_BODY() int }); ok {
		return impl.GetBoundaryDimension_BODY()
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) GetEnvelope() *Geom_Geometry {
	return g.GetFactory().ToGeometry(g.GetEnvelopeInternal())
}

func (g *Geom_Geometry) GetEnvelopeInternal() *Geom_Envelope {
	if g.envelope == nil {
		g.envelope = g.ComputeEnvelopeInternal()
	}
	return Geom_NewEnvelopeFromEnvelope(g.envelope)
}

func (g *Geom_Geometry) GeometryChanged() {
	g.Apply(geom_Geometry_geometryChangedFilter)
}

func (g *Geom_Geometry) GeometryChangedAction() {
	g.envelope = nil
}

func (g *Geom_Geometry) Disjoint(other *Geom_Geometry) bool {
	return !g.Intersects(other)
}

func (g *Geom_Geometry) Touches(other *Geom_Geometry) bool {
	return geom_GeometryRelate_Touches(g, other)
}

func (g *Geom_Geometry) Intersects(other *Geom_Geometry) bool {
	if !g.GetEnvelopeInternal().IntersectsEnvelope(other.GetEnvelopeInternal()) {
		return false
	}
	if g.IsRectangle() {
		return OperationPredicate_RectangleIntersects_Intersects(java.Cast[*Geom_Polygon](g), other)
	}
	if other.IsRectangle() {
		return OperationPredicate_RectangleIntersects_Intersects(java.Cast[*Geom_Polygon](other), g)
	}
	return geom_GeometryRelate_Intersects(g, other)
}

func (g *Geom_Geometry) Crosses(other *Geom_Geometry) bool {
	// Short-circuit test.
	if !g.GetEnvelopeInternal().IntersectsEnvelope(other.GetEnvelopeInternal()) {
		return false
	}
	return g.RelateMatrix(other).IsCrosses(g.GetDimension(), other.GetDimension())
}

func (g *Geom_Geometry) Within(other *Geom_Geometry) bool {
	return geom_GeometryRelate_Within(g, other)
}

func (g *Geom_Geometry) Contains(other *Geom_Geometry) bool {
	if g.IsRectangle() {
		return OperationPredicate_RectangleContains_Contains(java.Cast[*Geom_Polygon](g), other)
	}
	return geom_GeometryRelate_Contains(g, other)
}

func (g *Geom_Geometry) Overlaps(other *Geom_Geometry) bool {
	return geom_GeometryRelate_Overlaps(g, other)
}

func (g *Geom_Geometry) Covers(other *Geom_Geometry) bool {
	return geom_GeometryRelate_Covers(g, other)
}

func (g *Geom_Geometry) CoveredBy(other *Geom_Geometry) bool {
	return geom_GeometryRelate_CoveredBy(g, other)
}

func (g *Geom_Geometry) Relate(other *Geom_Geometry, intersectionPattern string) bool {
	return geom_GeometryRelate_RelatePattern(g, other, intersectionPattern)
}

func (g *Geom_Geometry) RelateMatrix(other *Geom_Geometry) *Geom_IntersectionMatrix {
	return geom_GeometryRelate_Relate(g, other)
}

func (g *Geom_Geometry) EqualsGeometry(other *Geom_Geometry) bool {
	if other == nil {
		return false
	}
	return g.EqualsTopo(other)
}

func (g *Geom_Geometry) EqualsTopo(other *Geom_Geometry) bool {
	return geom_GeometryRelate_EqualsTopo(g, other)
}

func (g *Geom_Geometry) EqualsObject(o any) bool {
	p, ok := o.(java.Polymorphic)
	if !ok {
		return false
	}
	if !java.InstanceOf[*Geom_Geometry](p) {
		return false
	}
	geom := java.Cast[*Geom_Geometry](p)
	return g.EqualsExact(geom)
}

func (g *Geom_Geometry) HashCode() int {
	return g.GetEnvelopeInternal().HashCode()
}

func (g *Geom_Geometry) String() string {
	return g.ToText()
}

func (g *Geom_Geometry) ToText() string {
	writer := Io_NewWKTWriter()
	return writer.Write(g)
}

func (g *Geom_Geometry) Buffer(distance float64) *Geom_Geometry {
	return OperationBuffer_BufferOp_BufferOp(g, distance)
}

func (g *Geom_Geometry) BufferWithQuadrantSegments(distance float64, quadrantSegments int) *Geom_Geometry {
	return OperationBuffer_BufferOp_BufferOpWithQuadrantSegments(g, distance, quadrantSegments)
}

func (g *Geom_Geometry) BufferWithQuadrantSegmentsAndEndCapStyle(distance float64, quadrantSegments, endCapStyle int) *Geom_Geometry {
	return OperationBuffer_BufferOp_BufferOpWithQuadrantSegmentsAndEndCapStyle(g, distance, quadrantSegments, endCapStyle)
}

func (g *Geom_Geometry) ConvexHull() *Geom_Geometry {
	return Algorithm_NewConvexHull(g).GetConvexHull()
}

func (g *Geom_Geometry) Reverse() *Geom_Geometry {
	if impl, ok := java.GetLeaf(g).(interface{ Reverse_BODY() *Geom_Geometry }); ok {
		return impl.Reverse_BODY()
	}
	return g.Reverse_BODY()
}

func (g *Geom_Geometry) Reverse_BODY() *Geom_Geometry {
	res := g.ReverseInternal()
	if g.envelope != nil {
		res.envelope = g.envelope.Copy()
	}
	res.SetSRID(g.GetSRID())
	return res
}

func (g *Geom_Geometry) ReverseInternal() *Geom_Geometry {
	if impl, ok := java.GetLeaf(g).(interface{ ReverseInternal_BODY() *Geom_Geometry }); ok {
		return impl.ReverseInternal_BODY()
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) Intersection(other *Geom_Geometry) *Geom_Geometry {
	return Geom_GeometryOverlay_Intersection(g, other)
}

func (g *Geom_Geometry) Union(other *Geom_Geometry) *Geom_Geometry {
	return Geom_GeometryOverlay_Union(g, other)
}

func (g *Geom_Geometry) Difference(other *Geom_Geometry) *Geom_Geometry {
	return Geom_GeometryOverlay_Difference(g, other)
}

func (g *Geom_Geometry) SymDifference(other *Geom_Geometry) *Geom_Geometry {
	return Geom_GeometryOverlay_SymDifference(g, other)
}

func (g *Geom_Geometry) UnionSelf() *Geom_Geometry {
	return Geom_GeometryOverlay_UnionSelf(g)
}

func (g *Geom_Geometry) EqualsExactWithTolerance(other *Geom_Geometry, tolerance float64) bool {
	if impl, ok := java.GetLeaf(g).(interface {
		EqualsExactWithTolerance_BODY(*Geom_Geometry, float64) bool
	}); ok {
		return impl.EqualsExactWithTolerance_BODY(other, tolerance)
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) EqualsExact(other *Geom_Geometry) bool {
	if impl, ok := java.GetLeaf(g).(interface{ EqualsExact_BODY(*Geom_Geometry) bool }); ok {
		return impl.EqualsExact_BODY(other)
	}
	return g.EqualsExact_BODY(other)
}

func (g *Geom_Geometry) EqualsExact_BODY(other *Geom_Geometry) bool {
	if g == other {
		return true
	}
	return g.EqualsExactWithTolerance(other, 0)
}

func (g *Geom_Geometry) EqualsNorm(other *Geom_Geometry) bool {
	if other == nil {
		return false
	}
	return g.Norm().EqualsExact(other.Norm())
}

func (g *Geom_Geometry) ApplyCoordinateFilter(filter Geom_CoordinateFilter) {
	if impl, ok := java.GetLeaf(g).(interface{ ApplyCoordinateFilter_BODY(Geom_CoordinateFilter) }); ok {
		impl.ApplyCoordinateFilter_BODY(filter)
		return
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) ApplyCoordinateSequenceFilter(filter Geom_CoordinateSequenceFilter) {
	if impl, ok := java.GetLeaf(g).(interface {
		ApplyCoordinateSequenceFilter_BODY(Geom_CoordinateSequenceFilter)
	}); ok {
		impl.ApplyCoordinateSequenceFilter_BODY(filter)
		return
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) ApplyGeometryFilter(filter Geom_GeometryFilter) {
	if impl, ok := java.GetLeaf(g).(interface{ ApplyGeometryFilter_BODY(Geom_GeometryFilter) }); ok {
		impl.ApplyGeometryFilter_BODY(filter)
		return
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) Apply(filter Geom_GeometryComponentFilter) {
	if impl, ok := java.GetLeaf(g).(interface {
		Apply_BODY(Geom_GeometryComponentFilter)
	}); ok {
		impl.Apply_BODY(filter)
		return
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) Clone() any {
	defer func() {
		if r := recover(); r != nil {
			Util_Assert_ShouldNeverReachHere()
		}
	}()
	panic("Clone is deprecated, use Copy instead")
}

func (g *Geom_Geometry) Copy() *Geom_Geometry {
	if impl, ok := java.GetLeaf(g).(interface{ Copy_BODY() *Geom_Geometry }); ok {
		return impl.Copy_BODY()
	}
	return g.Copy_BODY()
}

func (g *Geom_Geometry) Copy_BODY() *Geom_Geometry {
	copyGeom := g.CopyInternal()
	if g.envelope == nil {
		copyGeom.envelope = nil
	} else {
		copyGeom.envelope = g.envelope.Copy()
	}
	copyGeom.srid = g.srid
	copyGeom.userData = g.userData
	return copyGeom
}

func (g *Geom_Geometry) CopyInternal() *Geom_Geometry {
	if impl, ok := java.GetLeaf(g).(interface{ CopyInternal_BODY() *Geom_Geometry }); ok {
		return impl.CopyInternal_BODY()
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) Normalize() {
	if impl, ok := java.GetLeaf(g).(interface{ Normalize_BODY() }); ok {
		impl.Normalize_BODY()
		return
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) Norm() *Geom_Geometry {
	copyGeom := g.Copy()
	copyGeom.Normalize()
	return copyGeom
}

func (g *Geom_Geometry) CompareTo(o any) int {
	other := o.(*Geom_Geometry)
	if g.GetTypeCode() != other.GetTypeCode() {
		return g.GetTypeCode() - other.GetTypeCode()
	}
	if g.IsEmpty() && other.IsEmpty() {
		return 0
	}
	if g.IsEmpty() {
		return -1
	}
	if other.IsEmpty() {
		return 1
	}
	return g.CompareToSameClass(o)
}

func (g *Geom_Geometry) CompareToWithComparator(o any, comp *Geom_CoordinateSequenceComparator) int {
	other := o.(*Geom_Geometry)
	if g.GetTypeCode() != other.GetTypeCode() {
		return g.GetTypeCode() - other.GetTypeCode()
	}
	if g.IsEmpty() && other.IsEmpty() {
		return 0
	}
	if g.IsEmpty() {
		return -1
	}
	if other.IsEmpty() {
		return 1
	}
	return g.CompareToSameClassWithComparator(o, comp)
}

func (g *Geom_Geometry) IsEquivalentClass(other *Geom_Geometry) bool {
	if impl, ok := java.GetLeaf(g).(interface{ IsEquivalentClass_BODY(*Geom_Geometry) bool }); ok {
		return impl.IsEquivalentClass_BODY(other)
	}
	return g.IsEquivalentClass_BODY(other)
}

func (g *Geom_Geometry) IsEquivalentClass_BODY(other *Geom_Geometry) bool {
	// Compare runtime types using GetLeaf, matching Java's behavior where
	// isEquivalentClass compares this.getClass().getName() with other.getClass().getName().
	selfType := reflect.TypeOf(java.GetLeaf(g))
	otherType := reflect.TypeOf(java.GetLeaf(other))
	return selfType == otherType
}

func Geom_Geometry_CheckNotGeometryCollection(geom *Geom_Geometry) {
	if geom.IsGeometryCollection() {
		panic("Operation does not support GeometryCollection arguments")
	}
}

func (g *Geom_Geometry) IsGeometryCollection() bool {
	if impl, ok := java.GetLeaf(g).(interface{ IsGeometryCollection_BODY() bool }); ok {
		return impl.IsGeometryCollection_BODY()
	}
	return g.IsGeometryCollection_BODY()
}

func (g *Geom_Geometry) IsGeometryCollection_BODY() bool {
	return g.GetTypeCode() == Geom_Geometry_TypeCodeGeometryCollection
}

func (g *Geom_Geometry) ComputeEnvelopeInternal() *Geom_Envelope {
	if impl, ok := java.GetLeaf(g).(interface{ ComputeEnvelopeInternal_BODY() *Geom_Envelope }); ok {
		return impl.ComputeEnvelopeInternal_BODY()
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) CompareToSameClass(o any) int {
	if impl, ok := java.GetLeaf(g).(interface{ CompareToSameClass_BODY(any) int }); ok {
		return impl.CompareToSameClass_BODY(o)
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) CompareToSameClassWithComparator(o any, comp *Geom_CoordinateSequenceComparator) int {
	if impl, ok := java.GetLeaf(g).(interface {
		CompareToSameClassWithComparator_BODY(any, *Geom_CoordinateSequenceComparator) int
	}); ok {
		return impl.CompareToSameClassWithComparator_BODY(o, comp)
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) CompareCollections(a, b []any) int {
	i := 0
	j := 0
	for i < len(a) && j < len(b) {
		aElement := a[i].(interface{ CompareTo(any) int })
		bElement := b[j].(interface{ CompareTo(any) int })
		comparison := aElement.CompareTo(bElement)
		if comparison != 0 {
			return comparison
		}
		i++
		j++
	}
	if i < len(a) {
		return 1
	}
	if j < len(b) {
		return -1
	}
	return 0
}

func (g *Geom_Geometry) EqualCoordinate(a, b *Geom_Coordinate, tolerance float64) bool {
	if tolerance == 0 {
		return a.Equals(b)
	}
	return a.Distance(b) <= tolerance
}

func (g *Geom_Geometry) GetTypeCode() int {
	if impl, ok := java.GetLeaf(g).(interface{ GetTypeCode_BODY() int }); ok {
		return impl.GetTypeCode_BODY()
	}
	panic("abstract method called")
}

func (g *Geom_Geometry) createPointFromInternalCoord(coord *Geom_Coordinate, exemplar *Geom_Geometry) *Geom_Point {
	if coord == nil {
		return exemplar.GetFactory().CreatePoint()
	}
	exemplar.GetPrecisionModel().MakePreciseCoordinate(coord)
	return exemplar.GetFactory().CreatePointFromCoordinate(coord)
}
