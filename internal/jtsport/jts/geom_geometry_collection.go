package jts

import (
	"sort"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geom_GeometryCollection models a collection of Geometry objects of arbitrary
// type and dimension.
type Geom_GeometryCollection struct {
	*Geom_Geometry
	geometries []*Geom_Geometry
	child      java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (gc *Geom_GeometryCollection) GetChild() java.Polymorphic {
	return gc.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (gc *Geom_GeometryCollection) GetParent() java.Polymorphic {
	return gc.Geom_Geometry
}

// Geom_NewGeometryCollectionWithPrecisionModelAndSRID constructs a
// GeometryCollection with the given geometries.
//
// Deprecated: Use GeometryFactory instead.
func Geom_NewGeometryCollectionWithPrecisionModelAndSRID(geometries []*Geom_Geometry, precisionModel *Geom_PrecisionModel, SRID int) *Geom_GeometryCollection {
	return Geom_NewGeometryCollection(geometries, Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, SRID))
}

// Geom_NewGeometryCollection constructs a GeometryCollection.
//
// Parameters:
//   - geometries: the Geometrys for this GeometryCollection, or nil or an empty
//     array to create the empty geometry. Elements may be empty Geometrys, but
//     not nils.
func Geom_NewGeometryCollection(geometries []*Geom_Geometry, factory *Geom_GeometryFactory) *Geom_GeometryCollection {
	geom := &Geom_Geometry{factory: factory}
	if geometries == nil {
		geometries = []*Geom_Geometry{}
	}
	if geom_GeometryCollection_hasNullElements(geometries) {
		panic("geometries must not contain nil elements")
	}
	gc := &Geom_GeometryCollection{
		Geom_Geometry: geom,
		geometries:   geometries,
	}
	geom.child = gc
	return gc
}

func geom_GeometryCollection_hasNullElements(geometries []*Geom_Geometry) bool {
	for _, g := range geometries {
		if g == nil {
			return true
		}
	}
	return false
}

func (gc *Geom_GeometryCollection) GetCoordinate_BODY() *Geom_Coordinate {
	for i := 0; i < len(gc.geometries); i++ {
		if !gc.geometries[i].IsEmpty() {
			return gc.geometries[i].GetCoordinate()
		}
	}
	return nil
}

// GetCoordinates collects all coordinates of all subgeometries into an array.
//
// Note that while changes to the coordinate objects themselves may modify the
// Geometries in place, the returned array as such is only a temporary container
// which is not synchronized back.
func (gc *Geom_GeometryCollection) GetCoordinates_BODY() []*Geom_Coordinate {
	coordinates := make([]*Geom_Coordinate, gc.GetNumPoints())
	k := -1
	for i := 0; i < len(gc.geometries); i++ {
		childCoordinates := gc.geometries[i].GetCoordinates()
		for j := 0; j < len(childCoordinates); j++ {
			k++
			coordinates[k] = childCoordinates[j]
		}
	}
	return coordinates
}

func (gc *Geom_GeometryCollection) IsEmpty_BODY() bool {
	for i := 0; i < len(gc.geometries); i++ {
		if !gc.geometries[i].IsEmpty() {
			return false
		}
	}
	return true
}

func (gc *Geom_GeometryCollection) GetDimension_BODY() int {
	dimension := Geom_Dimension_False
	for i := 0; i < len(gc.geometries); i++ {
		dim := gc.geometries[i].GetDimension()
		if dim > dimension {
			dimension = dim
		}
	}
	return dimension
}

func (gc *Geom_GeometryCollection) HasDimension_BODY(dim int) bool {
	for i := 0; i < len(gc.geometries); i++ {
		if gc.geometries[i].HasDimension(dim) {
			return true
		}
	}
	return false
}

func (gc *Geom_GeometryCollection) GetBoundaryDimension_BODY() int {
	dimension := Geom_Dimension_False
	for i := 0; i < len(gc.geometries); i++ {
		dim := gc.geometries[i].GetBoundaryDimension()
		if dim > dimension {
			dimension = dim
		}
	}
	return dimension
}

func (gc *Geom_GeometryCollection) GetNumGeometries_BODY() int {
	return len(gc.geometries)
}

func (gc *Geom_GeometryCollection) GetGeometryN_BODY(n int) *Geom_Geometry {
	return gc.geometries[n]
}

func (gc *Geom_GeometryCollection) GetNumPoints_BODY() int {
	numPoints := 0
	for i := 0; i < len(gc.geometries); i++ {
		numPoints += gc.geometries[i].GetNumPoints()
	}
	return numPoints
}

func (gc *Geom_GeometryCollection) GetGeometryType_BODY() string {
	return Geom_Geometry_TypeNameGeometryCollection
}

func (gc *Geom_GeometryCollection) GetBoundary_BODY() *Geom_Geometry {
	Geom_Geometry_CheckNotGeometryCollection(gc.Geom_Geometry)
	Util_Assert_ShouldNeverReachHere()
	return nil
}

// GetArea returns the area of this GeometryCollection.
func (gc *Geom_GeometryCollection) GetArea_BODY() float64 {
	area := 0.0
	for i := 0; i < len(gc.geometries); i++ {
		area += gc.geometries[i].GetArea()
	}
	return area
}

func (gc *Geom_GeometryCollection) GetLength_BODY() float64 {
	sum := 0.0
	for i := 0; i < len(gc.geometries); i++ {
		sum += gc.geometries[i].GetLength()
	}
	return sum
}

func (gc *Geom_GeometryCollection) EqualsExactWithTolerance_BODY(other *Geom_Geometry, tolerance float64) bool {
	if !gc.IsEquivalentClass(other) {
		return false
	}
	otherCollection := java.Cast[*Geom_GeometryCollection](other)
	if len(gc.geometries) != len(otherCollection.geometries) {
		return false
	}
	for i := 0; i < len(gc.geometries); i++ {
		if !gc.geometries[i].EqualsExactWithTolerance(otherCollection.geometries[i], tolerance) {
			return false
		}
	}
	return true
}

func (gc *Geom_GeometryCollection) ApplyCoordinateFilter_BODY(filter Geom_CoordinateFilter) {
	for i := 0; i < len(gc.geometries); i++ {
		gc.geometries[i].ApplyCoordinateFilter(filter)
	}
}

func (gc *Geom_GeometryCollection) ApplyCoordinateSequenceFilter_BODY(filter Geom_CoordinateSequenceFilter) {
	if len(gc.geometries) == 0 {
		return
	}
	for i := 0; i < len(gc.geometries); i++ {
		gc.geometries[i].ApplyCoordinateSequenceFilter(filter)
		if filter.IsDone() {
			break
		}
	}
	if filter.IsGeometryChanged() {
		gc.GeometryChanged()
	}
}

func (gc *Geom_GeometryCollection) ApplyGeometryFilter_BODY(filter Geom_GeometryFilter) {
	filter.Filter(gc.Geom_Geometry)
	for i := 0; i < len(gc.geometries); i++ {
		gc.geometries[i].ApplyGeometryFilter(filter)
	}
}

func (gc *Geom_GeometryCollection) Apply_BODY(filter Geom_GeometryComponentFilter) {
	filter.Filter(gc.Geom_Geometry)
	for i := range gc.geometries {
		gc.geometries[i].Apply(filter)
	}
}

func (gc *Geom_GeometryCollection) CopyInternal_BODY() *Geom_Geometry {
	geometries := make([]*Geom_Geometry, len(gc.geometries))
	for i := 0; i < len(geometries); i++ {
		geometries[i] = gc.geometries[i].Copy()
	}
	return Geom_NewGeometryCollection(geometries, gc.factory).Geom_Geometry
}

func (gc *Geom_GeometryCollection) Normalize_BODY() {
	for i := 0; i < len(gc.geometries); i++ {
		gc.geometries[i].Normalize()
	}
	sort.Slice(gc.geometries, func(i, j int) bool {
		return gc.geometries[i].CompareTo(gc.geometries[j]) < 0
	})
}

func (gc *Geom_GeometryCollection) ComputeEnvelopeInternal_BODY() *Geom_Envelope {
	envelope := Geom_NewEnvelope()
	for i := 0; i < len(gc.geometries); i++ {
		envelope.ExpandToIncludeEnvelope(gc.geometries[i].GetEnvelopeInternal())
	}
	return envelope
}

func (gc *Geom_GeometryCollection) CompareToSameClass_BODY(o any) int {
	theseElements := make([]*Geom_Geometry, len(gc.geometries))
	copy(theseElements, gc.geometries)
	sort.Slice(theseElements, func(i, j int) bool {
		return theseElements[i].CompareTo(theseElements[j]) < 0
	})

	// o might be *Geom_GeometryCollection or a subtype like *Geom_MultiPoint.
	otherGC := java.Cast[*Geom_GeometryCollection](o.(*Geom_Geometry))
	otherElements := make([]*Geom_Geometry, len(otherGC.geometries))
	copy(otherElements, otherGC.geometries)
	sort.Slice(otherElements, func(i, j int) bool {
		return otherElements[i].CompareTo(otherElements[j]) < 0
	})

	return geom_GeometryCollection_compareGeometrySlices(theseElements, otherElements)
}

func geom_GeometryCollection_compareGeometrySlices(a, b []*Geom_Geometry) int {
	i := 0
	for i < len(a) && i < len(b) {
		comparison := a[i].CompareTo(b[i])
		if comparison != 0 {
			return comparison
		}
		i++
	}
	if i < len(a) {
		return 1
	}
	if i < len(b) {
		return -1
	}
	return 0
}

func (gc *Geom_GeometryCollection) CompareToSameClassWithComparator_BODY(o any, comp *Geom_CoordinateSequenceComparator) int {
	// o might be *Geom_GeometryCollection or a subtype like *Geom_MultiPoint.
	otherGC := java.Cast[*Geom_GeometryCollection](o.(*Geom_Geometry))

	n1 := gc.GetNumGeometries()
	n2 := otherGC.GetNumGeometries()
	i := 0
	for i < n1 && i < n2 {
		thisGeom := gc.GetGeometryN(i)
		otherGeom := otherGC.GetGeometryN(i)
		holeComp := thisGeom.CompareToSameClassWithComparator(otherGeom, comp)
		if holeComp != 0 {
			return holeComp
		}
		i++
	}
	if i < n1 {
		return 1
	}
	if i < n2 {
		return -1
	}
	return 0
}

func (gc *Geom_GeometryCollection) GetTypeCode_BODY() int {
	return Geom_Geometry_TypeCodeGeometryCollection
}

// Reverse creates a GeometryCollection with every component reversed. The order
// of the components in the collection are not reversed.
func (gc *Geom_GeometryCollection) Reverse_BODY() *Geom_Geometry {
	return gc.ReverseInternal().Geom_Geometry
}

func (gc *Geom_GeometryCollection) ReverseInternal() *Geom_GeometryCollection {
	geometries := make([]*Geom_Geometry, len(gc.geometries))
	for i := 0; i < len(geometries); i++ {
		geometries[i] = gc.geometries[i].Reverse()
	}
	return Geom_NewGeometryCollection(geometries, gc.factory)
}
