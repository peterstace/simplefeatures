package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geom_GeometryFactory supplies a set of utility methods for building Geometry objects from lists
// of Coordinates.
//
// Note that the factory constructor methods do not change the input coordinates in any way.
// In particular, they are not rounded to the supplied PrecisionModel.
// It is assumed that input Coordinates meet the given precision.
//
// Instances of this type are thread-safe.
type Geom_GeometryFactory struct {
	precisionModel            *Geom_PrecisionModel
	coordinateSequenceFactory Geom_CoordinateSequenceFactory
	srid                      int
}

// Geom_GeometryFactory_CreatePointFromInternalCoord creates a Point from an internal coordinate.
// The coordinate is made precise using the exemplar's precision model.
func Geom_GeometryFactory_CreatePointFromInternalCoord(coord *Geom_Coordinate, exemplar *Geom_Geometry) *Geom_Point {
	exemplar.GetPrecisionModel().MakePreciseCoordinate(coord)
	return exemplar.GetFactory().CreatePointFromCoordinate(coord)
}

// Geom_NewGeometryFactory constructs a GeometryFactory that generates Geometries having the given
// PrecisionModel, spatial-reference ID, and CoordinateSequence implementation.
func Geom_NewGeometryFactory(precisionModel *Geom_PrecisionModel, srid int, coordinateSequenceFactory Geom_CoordinateSequenceFactory) *Geom_GeometryFactory {
	return &Geom_GeometryFactory{
		precisionModel:            precisionModel,
		coordinateSequenceFactory: coordinateSequenceFactory,
		srid:                      srid,
	}
}

// Geom_NewGeometryFactoryWithCoordinateSequenceFactory constructs a GeometryFactory that generates Geometries having the given
// CoordinateSequence implementation, a double-precision floating PrecisionModel and a
// spatial-reference ID of 0.
func Geom_NewGeometryFactoryWithCoordinateSequenceFactory(coordinateSequenceFactory Geom_CoordinateSequenceFactory) *Geom_GeometryFactory {
	return Geom_NewGeometryFactory(Geom_NewPrecisionModel(), 0, coordinateSequenceFactory)
}

// Geom_NewGeometryFactoryWithPrecisionModel constructs a GeometryFactory that generates Geometries having the given
// PrecisionModel and the default CoordinateSequence implementation.
func Geom_NewGeometryFactoryWithPrecisionModel(precisionModel *Geom_PrecisionModel) *Geom_GeometryFactory {
	return Geom_NewGeometryFactory(precisionModel, 0, geom_GeometryFactory_getDefaultCoordinateSequenceFactory())
}

// Geom_NewGeometryFactoryWithPrecisionModelAndSRID constructs a GeometryFactory that generates Geometries having the given
// PrecisionModel and spatial-reference ID, and the default CoordinateSequence implementation.
func Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel *Geom_PrecisionModel, srid int) *Geom_GeometryFactory {
	return Geom_NewGeometryFactory(precisionModel, srid, geom_GeometryFactory_getDefaultCoordinateSequenceFactory())
}

// Geom_NewGeometryFactoryDefault constructs a GeometryFactory that generates Geometries having a floating
// PrecisionModel and a spatial-reference ID of 0.
func Geom_NewGeometryFactoryDefault() *Geom_GeometryFactory {
	return Geom_NewGeometryFactoryWithPrecisionModelAndSRID(Geom_NewPrecisionModel(), 0)
}

// geom_GeometryFactory_defaultCoordinateSequenceFactory is set by the impl package during initialization.
var geom_GeometryFactory_defaultCoordinateSequenceFactory Geom_CoordinateSequenceFactory

// Geom_SetDefaultCoordinateSequenceFactory sets the default CoordinateSequenceFactory.
// This is called by the impl package during initialization.
func Geom_SetDefaultCoordinateSequenceFactory(factory Geom_CoordinateSequenceFactory) {
	geom_GeometryFactory_defaultCoordinateSequenceFactory = factory
}

func geom_GeometryFactory_getDefaultCoordinateSequenceFactory() Geom_CoordinateSequenceFactory {
	return geom_GeometryFactory_defaultCoordinateSequenceFactory
}

// Geom_GeometryFactory_ToPointArray converts the slice to an array.
func Geom_GeometryFactory_ToPointArray(points []*Geom_Point) []*Geom_Point {
	return points
}

// Geom_GeometryFactory_ToGeometryArray converts the slice to an array.
func Geom_GeometryFactory_ToGeometryArray(geometries []*Geom_Geometry) []*Geom_Geometry {
	if geometries == nil {
		return nil
	}
	return geometries
}

// Geom_GeometryFactory_ToLinearRingArray converts the slice to an array.
func Geom_GeometryFactory_ToLinearRingArray(linearRings []*Geom_LinearRing) []*Geom_LinearRing {
	return linearRings
}

// Geom_GeometryFactory_ToLineStringArray converts the slice to an array.
func Geom_GeometryFactory_ToLineStringArray(lineStrings []*Geom_LineString) []*Geom_LineString {
	return lineStrings
}

// Geom_GeometryFactory_ToPolygonArray converts the slice to an array.
func Geom_GeometryFactory_ToPolygonArray(polygons []*Geom_Polygon) []*Geom_Polygon {
	return polygons
}

// Geom_GeometryFactory_ToMultiPolygonArray converts the slice to an array.
func Geom_GeometryFactory_ToMultiPolygonArray(multiPolygons []*Geom_MultiPolygon) []*Geom_MultiPolygon {
	return multiPolygons
}

// Geom_GeometryFactory_ToMultiLineStringArray converts the slice to an array.
func Geom_GeometryFactory_ToMultiLineStringArray(multiLineStrings []*Geom_MultiLineString) []*Geom_MultiLineString {
	return multiLineStrings
}

// Geom_GeometryFactory_ToMultiPointArray converts the slice to an array.
func Geom_GeometryFactory_ToMultiPointArray(multiPoints []*Geom_MultiPoint) []*Geom_MultiPoint {
	return multiPoints
}

// ToGeometry creates a Geometry with the same extent as the given envelope.
// The Geometry returned is guaranteed to be valid.
// To provide this behaviour, the following cases occur:
//
// If the Envelope is:
//   - null: returns an empty Point
//   - a point: returns a non-empty Point
//   - a line: returns a two-point LineString
//   - a rectangle: returns a Polygon whose points are (minx, miny),
//     (minx, maxy), (maxx, maxy), (maxx, miny), (minx, miny).
func (gf *Geom_GeometryFactory) ToGeometry(envelope *Geom_Envelope) *Geom_Geometry {
	if envelope.IsNull() {
		point := gf.CreatePoint()
		return point.Geom_Geometry
	}

	if envelope.GetMinX() == envelope.GetMaxX() && envelope.GetMinY() == envelope.GetMaxY() {
		point := gf.CreatePointFromCoordinate(Geom_NewCoordinateWithXY(envelope.GetMinX(), envelope.GetMinY()))
		return point.Geom_Geometry
	}

	if envelope.GetMinX() == envelope.GetMaxX() || envelope.GetMinY() == envelope.GetMaxY() {
		lineString := gf.CreateLineStringFromCoordinates([]*Geom_Coordinate{
			Geom_NewCoordinateWithXY(envelope.GetMinX(), envelope.GetMinY()),
			Geom_NewCoordinateWithXY(envelope.GetMaxX(), envelope.GetMaxY()),
		})
		return lineString.Geom_Geometry
	}

	polygon := gf.CreatePolygonWithLinearRingAndHoles(
		gf.CreateLinearRingFromCoordinates([]*Geom_Coordinate{
			Geom_NewCoordinateWithXY(envelope.GetMinX(), envelope.GetMinY()),
			Geom_NewCoordinateWithXY(envelope.GetMinX(), envelope.GetMaxY()),
			Geom_NewCoordinateWithXY(envelope.GetMaxX(), envelope.GetMaxY()),
			Geom_NewCoordinateWithXY(envelope.GetMaxX(), envelope.GetMinY()),
			Geom_NewCoordinateWithXY(envelope.GetMinX(), envelope.GetMinY()),
		}),
		nil,
	)
	return polygon.Geom_Geometry
}

// GetPrecisionModel returns the PrecisionModel that Geometries created by this factory
// will be associated with.
func (gf *Geom_GeometryFactory) GetPrecisionModel() *Geom_PrecisionModel {
	return gf.precisionModel
}

// CreatePoint constructs an empty Point geometry.
func (gf *Geom_GeometryFactory) CreatePoint() *Geom_Point {
	return gf.CreatePointFromCoordinateSequence(gf.GetCoordinateSequenceFactory().CreateFromCoordinates([]*Geom_Coordinate{}))
}

// CreatePointFromCoordinate creates a Point using the given Coordinate.
// A nil Coordinate creates an empty Geometry.
func (gf *Geom_GeometryFactory) CreatePointFromCoordinate(coordinate *Geom_Coordinate) *Geom_Point {
	var coords Geom_CoordinateSequence
	if coordinate != nil {
		coords = gf.GetCoordinateSequenceFactory().CreateFromCoordinates([]*Geom_Coordinate{coordinate})
	}
	return gf.CreatePointFromCoordinateSequence(coords)
}

// CreatePointFromCoordinateSequence creates a Point using the given CoordinateSequence; a nil or empty
// CoordinateSequence will create an empty Point.
func (gf *Geom_GeometryFactory) CreatePointFromCoordinateSequence(coordinates Geom_CoordinateSequence) *Geom_Point {
	return Geom_NewPoint(coordinates, gf)
}

// CreateMultiLineString constructs an empty MultiLineString geometry.
func (gf *Geom_GeometryFactory) CreateMultiLineString() *Geom_MultiLineString {
	return Geom_NewMultiLineString(nil, gf)
}

// CreateMultiLineStringFromLineStrings creates a MultiLineString using the given LineStrings; a nil or empty
// array will create an empty MultiLineString.
func (gf *Geom_GeometryFactory) CreateMultiLineStringFromLineStrings(lineStrings []*Geom_LineString) *Geom_MultiLineString {
	return Geom_NewMultiLineString(lineStrings, gf)
}

// CreateGeometryCollection constructs an empty GeometryCollection geometry.
func (gf *Geom_GeometryFactory) CreateGeometryCollection() *Geom_GeometryCollection {
	return Geom_NewGeometryCollection(nil, gf)
}

// CreateGeometryCollectionFromGeometries creates a GeometryCollection using the given Geometries; a nil or empty
// array will create an empty GeometryCollection.
func (gf *Geom_GeometryFactory) CreateGeometryCollectionFromGeometries(geometries []*Geom_Geometry) *Geom_GeometryCollection {
	return Geom_NewGeometryCollection(geometries, gf)
}

// CreateMultiPolygon constructs an empty MultiPolygon geometry.
func (gf *Geom_GeometryFactory) CreateMultiPolygon() *Geom_MultiPolygon {
	return Geom_NewMultiPolygon(nil, gf)
}

// CreateMultiPolygonFromPolygons creates a MultiPolygon using the given Polygons; a nil or empty array
// will create an empty Polygon. The polygons must conform to the
// assertions specified in the OpenGIS Simple Features Specification for SQL.
func (gf *Geom_GeometryFactory) CreateMultiPolygonFromPolygons(polygons []*Geom_Polygon) *Geom_MultiPolygon {
	return Geom_NewMultiPolygon(polygons, gf)
}

// CreateLinearRing constructs an empty LinearRing geometry.
func (gf *Geom_GeometryFactory) CreateLinearRing() *Geom_LinearRing {
	return gf.CreateLinearRingFromCoordinateSequence(gf.GetCoordinateSequenceFactory().CreateFromCoordinates([]*Geom_Coordinate{}))
}

// CreateLinearRingFromCoordinates creates a LinearRing using the given Coordinates.
// A nil or empty array creates an empty LinearRing.
// The points must form a closed and simple linestring.
func (gf *Geom_GeometryFactory) CreateLinearRingFromCoordinates(coordinates []*Geom_Coordinate) *Geom_LinearRing {
	var coords Geom_CoordinateSequence
	if coordinates != nil {
		coords = gf.GetCoordinateSequenceFactory().CreateFromCoordinates(coordinates)
	}
	return gf.CreateLinearRingFromCoordinateSequence(coords)
}

// CreateLinearRingFromCoordinateSequence creates a LinearRing using the given CoordinateSequence.
// A nil or empty array creates an empty LinearRing.
// The points must form a closed and simple linestring.
func (gf *Geom_GeometryFactory) CreateLinearRingFromCoordinateSequence(coordinates Geom_CoordinateSequence) *Geom_LinearRing {
	return Geom_NewLinearRing(coordinates, gf)
}

// CreateMultiPoint constructs an empty MultiPoint geometry.
func (gf *Geom_GeometryFactory) CreateMultiPoint() *Geom_MultiPoint {
	return Geom_NewMultiPoint(nil, gf)
}

// CreateMultiPointFromPoints creates a MultiPoint using the given Points.
// A nil or empty array will create an empty MultiPoint.
func (gf *Geom_GeometryFactory) CreateMultiPointFromPoints(point []*Geom_Point) *Geom_MultiPoint {
	return Geom_NewMultiPoint(point, gf)
}

// CreateMultiPointFromCoordinates creates a MultiPoint using the given Coordinates.
// A nil or empty array will create an empty MultiPoint.
//
// Deprecated: Use CreateMultiPointFromCoords instead.
func (gf *Geom_GeometryFactory) CreateMultiPointFromCoordinates(coordinates []*Geom_Coordinate) *Geom_MultiPoint {
	var coords Geom_CoordinateSequence
	if coordinates != nil {
		coords = gf.GetCoordinateSequenceFactory().CreateFromCoordinates(coordinates)
	}
	return gf.CreateMultiPointFromCoordinateSequence(coords)
}

// CreateMultiPointFromCoords creates a MultiPoint using the given Coordinates.
// A nil or empty array will create an empty MultiPoint.
func (gf *Geom_GeometryFactory) CreateMultiPointFromCoords(coordinates []*Geom_Coordinate) *Geom_MultiPoint {
	var coords Geom_CoordinateSequence
	if coordinates != nil {
		coords = gf.GetCoordinateSequenceFactory().CreateFromCoordinates(coordinates)
	}
	return gf.CreateMultiPointFromCoordinateSequence(coords)
}

// CreateMultiPointFromCoordinateSequence creates a MultiPoint using the
// points in the given CoordinateSequence.
// A nil or empty CoordinateSequence creates an empty MultiPoint.
func (gf *Geom_GeometryFactory) CreateMultiPointFromCoordinateSequence(coordinates Geom_CoordinateSequence) *Geom_MultiPoint {
	if coordinates == nil || coordinates.Size() == 0 {
		return gf.CreateMultiPointFromPoints([]*Geom_Point{})
	}
	points := make([]*Geom_Point, coordinates.Size())
	for i := 0; i < coordinates.Size(); i++ {
		ptSeq := gf.GetCoordinateSequenceFactory().CreateWithSizeAndDimensionAndMeasures(1, coordinates.GetDimension(), coordinates.GetMeasures())
		Geom_CoordinateSequences_Copy(coordinates, i, ptSeq, 0, 1)
		points[i] = gf.CreatePointFromCoordinateSequence(ptSeq)
	}
	return gf.CreateMultiPointFromPoints(points)
}

// CreatePolygonWithLinearRingAndHoles constructs a Polygon with the given exterior boundary and
// interior boundaries.
func (gf *Geom_GeometryFactory) CreatePolygonWithLinearRingAndHoles(shell *Geom_LinearRing, holes []*Geom_LinearRing) *Geom_Polygon {
	return Geom_NewPolygon(shell, holes, gf)
}

// CreatePolygonFromCoordinateSequence constructs a Polygon with the given exterior boundary.
func (gf *Geom_GeometryFactory) CreatePolygonFromCoordinateSequence(shell Geom_CoordinateSequence) *Geom_Polygon {
	return gf.CreatePolygonFromLinearRing(gf.CreateLinearRingFromCoordinateSequence(shell))
}

// CreatePolygonFromCoordinates constructs a Polygon with the given exterior boundary.
func (gf *Geom_GeometryFactory) CreatePolygonFromCoordinates(shell []*Geom_Coordinate) *Geom_Polygon {
	return gf.CreatePolygonFromLinearRing(gf.CreateLinearRingFromCoordinates(shell))
}

// CreatePolygonFromLinearRing constructs a Polygon with the given exterior boundary.
func (gf *Geom_GeometryFactory) CreatePolygonFromLinearRing(shell *Geom_LinearRing) *Geom_Polygon {
	return gf.CreatePolygonWithLinearRingAndHoles(shell, nil)
}

// CreatePolygon constructs an empty Polygon geometry.
func (gf *Geom_GeometryFactory) CreatePolygon() *Geom_Polygon {
	return gf.CreatePolygonWithLinearRingAndHoles(nil, nil)
}

// BuildGeometry builds an appropriate Geometry, MultiGeometry, or
// GeometryCollection to contain the Geometrys in it.
// For example:
//
//   - If geomList contains a single Polygon, the Polygon is returned.
//   - If geomList contains several Polygons, a MultiPolygon is returned.
//   - If geomList contains some Polygons and some LineStrings, a GeometryCollection is returned.
//   - If geomList is empty, an empty GeometryCollection is returned.
//
// Note that this method does not "flatten" Geometries in the input, and hence if
// any MultiGeometries are contained in the input a GeometryCollection containing
// them will be returned.
func (gf *Geom_GeometryFactory) BuildGeometry(geomList []*Geom_Geometry) *Geom_Geometry {
	var geomClass string
	isHeterogeneous := false
	hasGeometryCollection := false

	for _, geom := range geomList {
		partClass := geom.GetGeometryType()
		if geomClass == "" {
			geomClass = partClass
		}
		if partClass != geomClass {
			isHeterogeneous = true
		}
		if java.InstanceOf[*Geom_GeometryCollection](geom) {
			hasGeometryCollection = true
		}
	}

	if geomClass == "" {
		geomColl := gf.CreateGeometryCollection()
		return geomColl.Geom_Geometry
	}

	if isHeterogeneous || hasGeometryCollection {
		geomColl := gf.CreateGeometryCollectionFromGeometries(Geom_GeometryFactory_ToGeometryArray(geomList))
		return geomColl.Geom_Geometry
	}

	geom0 := geomList[0]
	isCollection := len(geomList) > 1

	if isCollection {
		if java.InstanceOf[*Geom_Polygon](geom0) {
			polygons := make([]*Geom_Polygon, len(geomList))
			for i, g := range geomList {
				polygons[i] = java.Cast[*Geom_Polygon](g)
			}
			multiPoly := gf.CreateMultiPolygonFromPolygons(Geom_GeometryFactory_ToPolygonArray(polygons))
			return multiPoly.Geom_Geometry
		} else if java.InstanceOf[*Geom_LineString](geom0) {
			lineStrings := make([]*Geom_LineString, len(geomList))
			for i, g := range geomList {
				lineStrings[i] = java.Cast[*Geom_LineString](g)
			}
			multiLine := gf.CreateMultiLineStringFromLineStrings(Geom_GeometryFactory_ToLineStringArray(lineStrings))
			return multiLine.Geom_Geometry
		} else if java.InstanceOf[*Geom_Point](geom0) {
			points := make([]*Geom_Point, len(geomList))
			for i, g := range geomList {
				points[i] = java.Cast[*Geom_Point](g)
			}
			multiPoint := gf.CreateMultiPointFromPoints(Geom_GeometryFactory_ToPointArray(points))
			return multiPoint.Geom_Geometry
		}
		Util_Assert_ShouldNeverReachHereWithMessage("Unhandled class: " + geom0.GetGeometryType())
	}

	return geom0
}

// CreateLineString constructs an empty LineString geometry.
func (gf *Geom_GeometryFactory) CreateLineString() *Geom_LineString {
	return gf.CreateLineStringFromCoordinateSequence(gf.GetCoordinateSequenceFactory().CreateFromCoordinates([]*Geom_Coordinate{}))
}

// CreateLineStringFromCoordinates creates a LineString using the given Coordinates.
// A nil or empty array creates an empty LineString.
func (gf *Geom_GeometryFactory) CreateLineStringFromCoordinates(coordinates []*Geom_Coordinate) *Geom_LineString {
	var coords Geom_CoordinateSequence
	if coordinates != nil {
		coords = gf.GetCoordinateSequenceFactory().CreateFromCoordinates(coordinates)
	}
	return gf.CreateLineStringFromCoordinateSequence(coords)
}

// CreateLineStringFromCoordinateSequence creates a LineString using the given CoordinateSequence.
// A nil or empty CoordinateSequence creates an empty LineString.
func (gf *Geom_GeometryFactory) CreateLineStringFromCoordinateSequence(coordinates Geom_CoordinateSequence) *Geom_LineString {
	return Geom_NewLineString(coordinates, gf)
}

// CreateEmpty creates an empty atomic geometry of the given dimension.
// If passed a dimension of -1 will create an empty GeometryCollection.
func (gf *Geom_GeometryFactory) CreateEmpty(dimension int) *Geom_Geometry {
	switch dimension {
	case -1:
		geomColl := gf.CreateGeometryCollection()
		return geomColl.Geom_Geometry
	case 0:
		point := gf.CreatePoint()
		return point.Geom_Geometry
	case 1:
		lineString := gf.CreateLineString()
		return lineString.Geom_Geometry
	case 2:
		polygon := gf.CreatePolygon()
		return polygon.Geom_Geometry
	default:
		panic("Invalid dimension")
	}
}

// CreateGeometry creates a deep copy of the input Geometry.
// The CoordinateSequenceFactory defined for this factory
// is used to copy the CoordinateSequences of the input geometry.
//
// This is a convenient way to change the CoordinateSequence
// used to represent a geometry, or to change the factory used for a geometry.
//
// Geometry.Copy can also be used to make a deep copy,
// but it does not allow changing the CoordinateSequence type.
func (gf *Geom_GeometryFactory) CreateGeometry(g *Geom_Geometry) *Geom_Geometry {
	editor := GeomUtil_NewGeometryEditorWithFactory(gf)
	op := geom_newCoordSeqCloneOp(gf.coordinateSequenceFactory)
	return editor.Edit(g, op)
}

type geom_GeometryFactory_coordSeqCloneOp struct {
	GeomUtil_GeometryEditor_CoordinateSequenceOperationBase
	coordinateSequenceFactory Geom_CoordinateSequenceFactory
}

func geom_newCoordSeqCloneOp(csf Geom_CoordinateSequenceFactory) *geom_GeometryFactory_coordSeqCloneOp {
	op := &geom_GeometryFactory_coordSeqCloneOp{
		coordinateSequenceFactory: csf,
	}
	op.GeomUtil_GeometryEditor_CoordinateSequenceOperationBase.child = op
	return op
}

func (op *geom_GeometryFactory_coordSeqCloneOp) GetChild() java.Polymorphic {
	return nil
}

func (op *geom_GeometryFactory_coordSeqCloneOp) EditCoordinateSequence(coordSeq Geom_CoordinateSequence, geometry *Geom_Geometry) Geom_CoordinateSequence {
	return op.coordinateSequenceFactory.CreateFromCoordinateSequence(coordSeq)
}

// GetSRID gets the SRID value defined for this factory.
func (gf *Geom_GeometryFactory) GetSRID() int {
	return gf.srid
}

// GetCoordinateSequenceFactory returns the CoordinateSequenceFactory for this factory.
func (gf *Geom_GeometryFactory) GetCoordinateSequenceFactory() Geom_CoordinateSequenceFactory {
	return gf.coordinateSequenceFactory
}
