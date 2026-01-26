package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomUtil_GeometryTransformer is a framework for processes which transform an
// input Geometry into an output Geometry, possibly changing its structure and
// type(s). This class is a framework for implementing subclasses which perform
// transformations on various different Geometry subclasses. It provides an easy
// way of applying specific transformations to given geometry types, while
// allowing unhandled types to be simply copied. Also, the framework ensures
// that if subcomponents change type the parent geometries types change
// appropriately to maintain valid structure. Subclasses will override whichever
// transformX methods they need to handle particular Geometry types.
//
// A typical usage would be a transformation class that transforms Polygons into
// Polygons, LineStrings or Points, depending on the geometry of the input (For
// instance, a simplification operation). This class would likely need to
// override the TransformMultiPolygon method to ensure that if input Polygons
// change type the result is a GeometryCollection, not a MultiPolygon.
//
// The default behaviour of this class is simply to recursively transform each
// Geometry component into an identical object by deep copying down to the level
// of, but not including, coordinates.
//
// All transformX methods may return nil, to avoid creating empty or invalid
// geometry objects. This will be handled correctly by the transformer.
// transformXXX methods should always return valid geometry - if they cannot do
// this they should return nil (for instance, it may not be possible for a
// TransformLineString implementation to return at least two points - in this
// case, it should return nil). The Transform method itself will always return a
// non-nil Geometry object (but this may be empty).
type GeomUtil_GeometryTransformer struct {
	child java.Polymorphic

	inputGeom *Geom_Geometry
	factory   *Geom_GeometryFactory

	// pruneEmptyGeometry is true if empty geometries should not be included in
	// the result.
	pruneEmptyGeometry bool

	// preserveGeometryCollectionType is true if a homogenous collection result
	// from a GeometryCollection should still be a general GeometryCollection.
	preserveGeometryCollectionType bool

	// preserveCollections is true if the output from a collection argument
	// should still be a collection.
	preserveCollections bool

	// preserveType is true if the type of the input should be preserved.
	preserveType bool
}

// GetChild returns the immediate child in the type hierarchy chain.
func (gt *GeomUtil_GeometryTransformer) GetChild() java.Polymorphic {
	return gt.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (gt *GeomUtil_GeometryTransformer) GetParent() java.Polymorphic {
	return nil
}

// GeomUtil_NewGeometryTransformer creates a new GeometryTransformer.
func GeomUtil_NewGeometryTransformer() *GeomUtil_GeometryTransformer {
	return &GeomUtil_GeometryTransformer{
		pruneEmptyGeometry:             true,
		preserveGeometryCollectionType: true,
		preserveCollections:            false,
		preserveType:                   false,
	}
}

// GetInputGeometry is a utility function to make input geometry available.
func (gt *GeomUtil_GeometryTransformer) GetInputGeometry() *Geom_Geometry {
	return gt.inputGeom
}

// Transform transforms the input geometry.
func (gt *GeomUtil_GeometryTransformer) Transform(inputGeom *Geom_Geometry) *Geom_Geometry {
	gt.inputGeom = inputGeom
	gt.factory = inputGeom.GetFactory()

	if java.InstanceOf[*Geom_Point](inputGeom) {
		return gt.TransformPoint(java.Cast[*Geom_Point](inputGeom), nil)
	}
	if java.InstanceOf[*Geom_MultiPoint](inputGeom) {
		return gt.TransformMultiPoint(java.Cast[*Geom_MultiPoint](inputGeom), nil)
	}
	if java.InstanceOf[*Geom_LinearRing](inputGeom) {
		return gt.TransformLinearRing(java.Cast[*Geom_LinearRing](inputGeom), nil)
	}
	if java.InstanceOf[*Geom_LineString](inputGeom) {
		return gt.TransformLineString(java.Cast[*Geom_LineString](inputGeom), nil)
	}
	if java.InstanceOf[*Geom_MultiLineString](inputGeom) {
		return gt.TransformMultiLineString(java.Cast[*Geom_MultiLineString](inputGeom), nil)
	}
	if java.InstanceOf[*Geom_Polygon](inputGeom) {
		return gt.TransformPolygon(java.Cast[*Geom_Polygon](inputGeom), nil)
	}
	if java.InstanceOf[*Geom_MultiPolygon](inputGeom) {
		return gt.TransformMultiPolygon(java.Cast[*Geom_MultiPolygon](inputGeom), nil)
	}
	if java.InstanceOf[*Geom_GeometryCollection](inputGeom) {
		return gt.TransformGeometryCollection(java.Cast[*Geom_GeometryCollection](inputGeom), nil)
	}

	panic("Unknown Geometry subtype: " + inputGeom.GetGeometryType())
}

// CreateCoordinateSequence is a convenience method which provides a standard
// way of creating a CoordinateSequence.
func (gt *GeomUtil_GeometryTransformer) CreateCoordinateSequence(coords []*Geom_Coordinate) Geom_CoordinateSequence {
	return gt.factory.GetCoordinateSequenceFactory().CreateFromCoordinates(coords)
}

// Copy is a convenience method which provides a standard way of copying
// CoordinateSequences.
func (gt *GeomUtil_GeometryTransformer) Copy(seq Geom_CoordinateSequence) Geom_CoordinateSequence {
	return seq.Copy()
}

// TransformCoordinates transforms a CoordinateSequence. This method should
// always return a valid coordinate list for the desired result type. (E.g. a
// coordinate list for a LineString must have 0 or at least 2 points). If this
// is not possible, return an empty sequence - this will be pruned out.
func (gt *GeomUtil_GeometryTransformer) TransformCoordinates(coords Geom_CoordinateSequence, parent *Geom_Geometry) Geom_CoordinateSequence {
	if impl, ok := java.GetLeaf(gt).(interface {
		TransformCoordinates_BODY(Geom_CoordinateSequence, *Geom_Geometry) Geom_CoordinateSequence
	}); ok {
		return impl.TransformCoordinates_BODY(coords, parent)
	}
	return gt.TransformCoordinates_BODY(coords, parent)
}

// TransformCoordinates_BODY is the default implementation of
// TransformCoordinates.
func (gt *GeomUtil_GeometryTransformer) TransformCoordinates_BODY(coords Geom_CoordinateSequence, parent *Geom_Geometry) Geom_CoordinateSequence {
	return gt.Copy(coords)
}

// TransformPoint transforms a Point geometry.
func (gt *GeomUtil_GeometryTransformer) TransformPoint(geom *Geom_Point, parent *Geom_Geometry) *Geom_Geometry {
	if impl, ok := java.GetLeaf(gt).(interface {
		TransformPoint_BODY(*Geom_Point, *Geom_Geometry) *Geom_Geometry
	}); ok {
		return impl.TransformPoint_BODY(geom, parent)
	}
	return gt.TransformPoint_BODY(geom, parent)
}

// TransformPoint_BODY is the default implementation of TransformPoint.
func (gt *GeomUtil_GeometryTransformer) TransformPoint_BODY(geom *Geom_Point, parent *Geom_Geometry) *Geom_Geometry {
	return gt.factory.CreatePointFromCoordinateSequence(
		gt.TransformCoordinates(geom.GetCoordinateSequence(), geom.Geom_Geometry)).Geom_Geometry
}

// TransformMultiPoint transforms a MultiPoint geometry.
func (gt *GeomUtil_GeometryTransformer) TransformMultiPoint(geom *Geom_MultiPoint, parent *Geom_Geometry) *Geom_Geometry {
	if impl, ok := java.GetLeaf(gt).(interface {
		TransformMultiPoint_BODY(*Geom_MultiPoint, *Geom_Geometry) *Geom_Geometry
	}); ok {
		return impl.TransformMultiPoint_BODY(geom, parent)
	}
	return gt.TransformMultiPoint_BODY(geom, parent)
}

// TransformMultiPoint_BODY is the default implementation of TransformMultiPoint.
func (gt *GeomUtil_GeometryTransformer) TransformMultiPoint_BODY(geom *Geom_MultiPoint, parent *Geom_Geometry) *Geom_Geometry {
	var transGeomList []*Geom_Geometry
	for i := 0; i < geom.GetNumGeometries(); i++ {
		transformGeom := gt.TransformPoint(java.Cast[*Geom_Point](geom.GetGeometryN(i)), geom.Geom_GeometryCollection.Geom_Geometry)
		if transformGeom == nil {
			continue
		}
		if transformGeom.IsEmpty() {
			continue
		}
		transGeomList = append(transGeomList, transformGeom)
	}
	if len(transGeomList) == 0 {
		return gt.factory.CreateMultiPoint().Geom_GeometryCollection.Geom_Geometry
	}
	return gt.factory.BuildGeometry(transGeomList)
}

// TransformLinearRing transforms a LinearRing. The transformation of a
// LinearRing may result in a coordinate sequence which does not form a
// structurally valid ring (i.e. a degenerate ring of 3 or fewer points). In
// this case a LineString is returned. Subclasses may wish to override this
// method and check for this situation (e.g. a subclass may choose to eliminate
// degenerate linear rings).
func (gt *GeomUtil_GeometryTransformer) TransformLinearRing(geom *Geom_LinearRing, parent *Geom_Geometry) *Geom_Geometry {
	if impl, ok := java.GetLeaf(gt).(interface {
		TransformLinearRing_BODY(*Geom_LinearRing, *Geom_Geometry) *Geom_Geometry
	}); ok {
		return impl.TransformLinearRing_BODY(geom, parent)
	}
	return gt.TransformLinearRing_BODY(geom, parent)
}

// TransformLinearRing_BODY is the default implementation of TransformLinearRing.
func (gt *GeomUtil_GeometryTransformer) TransformLinearRing_BODY(geom *Geom_LinearRing, parent *Geom_Geometry) *Geom_Geometry {
	seq := gt.TransformCoordinates(geom.GetCoordinateSequence(), geom.Geom_LineString.Geom_Geometry)
	if seq == nil {
		return gt.factory.CreateLinearRingFromCoordinateSequence(nil).Geom_LineString.Geom_Geometry
	}
	seqSize := seq.Size()
	// Ensure a valid LinearRing.
	if seqSize > 0 && seqSize < 4 && !gt.preserveType {
		return gt.factory.CreateLineStringFromCoordinateSequence(seq).Geom_Geometry
	}
	return gt.factory.CreateLinearRingFromCoordinateSequence(seq).Geom_LineString.Geom_Geometry
}

// TransformLineString transforms a LineString geometry.
func (gt *GeomUtil_GeometryTransformer) TransformLineString(geom *Geom_LineString, parent *Geom_Geometry) *Geom_Geometry {
	if impl, ok := java.GetLeaf(gt).(interface {
		TransformLineString_BODY(*Geom_LineString, *Geom_Geometry) *Geom_Geometry
	}); ok {
		return impl.TransformLineString_BODY(geom, parent)
	}
	return gt.TransformLineString_BODY(geom, parent)
}

// TransformLineString_BODY is the default implementation of TransformLineString.
func (gt *GeomUtil_GeometryTransformer) TransformLineString_BODY(geom *Geom_LineString, parent *Geom_Geometry) *Geom_Geometry {
	// Should check for 1-point sequences and downgrade them to points.
	return gt.factory.CreateLineStringFromCoordinateSequence(
		gt.TransformCoordinates(geom.GetCoordinateSequence(), geom.Geom_Geometry)).Geom_Geometry
}

// TransformMultiLineString transforms a MultiLineString geometry.
func (gt *GeomUtil_GeometryTransformer) TransformMultiLineString(geom *Geom_MultiLineString, parent *Geom_Geometry) *Geom_Geometry {
	if impl, ok := java.GetLeaf(gt).(interface {
		TransformMultiLineString_BODY(*Geom_MultiLineString, *Geom_Geometry) *Geom_Geometry
	}); ok {
		return impl.TransformMultiLineString_BODY(geom, parent)
	}
	return gt.TransformMultiLineString_BODY(geom, parent)
}

// TransformMultiLineString_BODY is the default implementation of
// TransformMultiLineString.
func (gt *GeomUtil_GeometryTransformer) TransformMultiLineString_BODY(geom *Geom_MultiLineString, parent *Geom_Geometry) *Geom_Geometry {
	var transGeomList []*Geom_Geometry
	for i := 0; i < geom.GetNumGeometries(); i++ {
		transformGeom := gt.TransformLineString(java.Cast[*Geom_LineString](geom.GetGeometryN(i)), geom.Geom_GeometryCollection.Geom_Geometry)
		if transformGeom == nil {
			continue
		}
		if transformGeom.IsEmpty() {
			continue
		}
		transGeomList = append(transGeomList, transformGeom)
	}
	if len(transGeomList) == 0 {
		return gt.factory.CreateMultiLineString().Geom_GeometryCollection.Geom_Geometry
	}
	return gt.factory.BuildGeometry(transGeomList)
}

// TransformPolygon transforms a Polygon geometry.
func (gt *GeomUtil_GeometryTransformer) TransformPolygon(geom *Geom_Polygon, parent *Geom_Geometry) *Geom_Geometry {
	if impl, ok := java.GetLeaf(gt).(interface {
		TransformPolygon_BODY(*Geom_Polygon, *Geom_Geometry) *Geom_Geometry
	}); ok {
		return impl.TransformPolygon_BODY(geom, parent)
	}
	return gt.TransformPolygon_BODY(geom, parent)
}

// TransformPolygon_BODY is the default implementation of TransformPolygon.
func (gt *GeomUtil_GeometryTransformer) TransformPolygon_BODY(geom *Geom_Polygon, parent *Geom_Geometry) *Geom_Geometry {
	isAllValidLinearRings := true
	shell := gt.TransformLinearRing(geom.GetExteriorRing(), geom.Geom_Geometry)

	// Handle empty inputs, or inputs which are made empty.
	shellIsNullOrEmpty := shell == nil || shell.IsEmpty()
	if geom.IsEmpty() && shellIsNullOrEmpty {
		return gt.factory.CreatePolygon().Geom_Geometry
	}

	if shellIsNullOrEmpty || !java.InstanceOf[*Geom_LinearRing](shell) {
		isAllValidLinearRings = false
	}

	var holes []*Geom_Geometry
	for i := 0; i < geom.GetNumInteriorRing(); i++ {
		hole := gt.TransformLinearRing(geom.GetInteriorRingN(i), geom.Geom_Geometry)
		if hole == nil || hole.IsEmpty() {
			continue
		}
		if !java.InstanceOf[*Geom_LinearRing](hole) {
			isAllValidLinearRings = false
		}
		holes = append(holes, hole)
	}

	if isAllValidLinearRings {
		holeRings := make([]*Geom_LinearRing, len(holes))
		for i, h := range holes {
			holeRings[i] = java.Cast[*Geom_LinearRing](h)
		}
		return gt.factory.CreatePolygonWithLinearRingAndHoles(java.Cast[*Geom_LinearRing](shell), holeRings).Geom_Geometry
	}
	var components []*Geom_Geometry
	if shell != nil {
		components = append(components, shell)
	}
	components = append(components, holes...)
	return gt.factory.BuildGeometry(components)
}

// TransformMultiPolygon transforms a MultiPolygon geometry.
func (gt *GeomUtil_GeometryTransformer) TransformMultiPolygon(geom *Geom_MultiPolygon, parent *Geom_Geometry) *Geom_Geometry {
	if impl, ok := java.GetLeaf(gt).(interface {
		TransformMultiPolygon_BODY(*Geom_MultiPolygon, *Geom_Geometry) *Geom_Geometry
	}); ok {
		return impl.TransformMultiPolygon_BODY(geom, parent)
	}
	return gt.TransformMultiPolygon_BODY(geom, parent)
}

// TransformMultiPolygon_BODY is the default implementation of
// TransformMultiPolygon.
func (gt *GeomUtil_GeometryTransformer) TransformMultiPolygon_BODY(geom *Geom_MultiPolygon, parent *Geom_Geometry) *Geom_Geometry {
	var transGeomList []*Geom_Geometry
	for i := 0; i < geom.GetNumGeometries(); i++ {
		transformGeom := gt.TransformPolygon(java.Cast[*Geom_Polygon](geom.GetGeometryN(i)), geom.Geom_GeometryCollection.Geom_Geometry)
		if transformGeom == nil {
			continue
		}
		if transformGeom.IsEmpty() {
			continue
		}
		transGeomList = append(transGeomList, transformGeom)
	}
	if len(transGeomList) == 0 {
		return gt.factory.CreateMultiPolygon().Geom_GeometryCollection.Geom_Geometry
	}
	return gt.factory.BuildGeometry(transGeomList)
}

// TransformGeometryCollection transforms a GeometryCollection geometry.
func (gt *GeomUtil_GeometryTransformer) TransformGeometryCollection(geom *Geom_GeometryCollection, parent *Geom_Geometry) *Geom_Geometry {
	if impl, ok := java.GetLeaf(gt).(interface {
		TransformGeometryCollection_BODY(*Geom_GeometryCollection, *Geom_Geometry) *Geom_Geometry
	}); ok {
		return impl.TransformGeometryCollection_BODY(geom, parent)
	}
	return gt.TransformGeometryCollection_BODY(geom, parent)
}

// TransformGeometryCollection_BODY is the default implementation of
// TransformGeometryCollection.
func (gt *GeomUtil_GeometryTransformer) TransformGeometryCollection_BODY(geom *Geom_GeometryCollection, parent *Geom_Geometry) *Geom_Geometry {
	var transGeomList []*Geom_Geometry
	for i := 0; i < geom.GetNumGeometries(); i++ {
		transformGeom := gt.Transform(geom.GetGeometryN(i))
		if transformGeom == nil {
			continue
		}
		if gt.pruneEmptyGeometry && transformGeom.IsEmpty() {
			continue
		}
		transGeomList = append(transGeomList, transformGeom)
	}
	if gt.preserveGeometryCollectionType {
		return gt.factory.CreateGeometryCollectionFromGeometries(Geom_GeometryFactory_ToGeometryArray(transGeomList)).Geom_Geometry
	}
	return gt.factory.BuildGeometry(transGeomList)
}
