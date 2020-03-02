package geom

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"unsafe"
)

// Geometry is a single geometry of any type. It's zero value is valid and is
// an empty GeometryCollection.
type Geometry struct {
	tag geometryTag
	ptr unsafe.Pointer
}

type geometryTag int

const (
	geometryCollectionTag geometryTag = iota
	pointTag
	lineTag
	lineStringTag
	polygonTag
	multiPointTag
	multiLineStringTag
	multiPolygonTag
)

func (t geometryTag) String() string {
	s, ok := map[geometryTag]string{
		geometryCollectionTag: "GeometryCollection",
		pointTag:              "Point",
		lineTag:               "Line",
		lineStringTag:         "LineString",
		polygonTag:            "Polygon",
		multiPointTag:         "MultiPoint",
		multiLineStringTag:    "MultiLineString",
		multiPolygonTag:       "MultiPolygon",
	}[t]
	if !ok {
		return "invalid"
	}
	return s
}

const (
	geometryCollectionType = "GeometryCollection"
	pointType              = "Point"
	lineStringType         = "LineString"
	polygonType            = "Polygon"
	multiPointType         = "MultiPoint"
	multiLineStringType    = "MultiLineString"
	multiPolygonType       = "MultiPolygon"
)

// Type returns a string representation of the geometry's type.
func (g Geometry) Type() string {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().Type()
	case pointTag:
		return g.AsPoint().Type()
	case lineTag:
		return g.AsLine().Type()
	case lineStringTag:
		return g.AsLineString().Type()
	case polygonTag:
		return g.AsPolygon().Type()
	case multiPointTag:
		return g.AsMultiPoint().Type()
	case multiLineStringTag:
		return g.AsMultiLineString().Type()
	case multiPolygonTag:
		return g.AsMultiPolygon().Type()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// IsGeometryCollection return true iff the Geometry is a GeometryCollection geometry.
func (g Geometry) IsGeometryCollection() bool { return g.tag == geometryCollectionTag }

// IsPoint return true iff the Geometry is a Point geometry.
func (g Geometry) IsPoint() bool { return g.tag == pointTag }

// IsLine return true iff the Geometry is a Line geometry.
func (g Geometry) IsLine() bool { return g.tag == lineTag }

// IsLineString return true iff the Geometry is a LineString geometry.
func (g Geometry) IsLineString() bool { return g.tag == lineStringTag }

// IsPolygon return true iff the Geometry is a Polygon geometry.
func (g Geometry) IsPolygon() bool { return g.tag == polygonTag }

// IsMultiPoint return true iff the Geometry is a MultiPoint geometry.
func (g Geometry) IsMultiPoint() bool { return g.tag == multiPointTag }

// IsMultiLineString return true iff the Geometry is a MultiLineString geometry.
func (g Geometry) IsMultiLineString() bool { return g.tag == multiLineStringTag }

// IsMultiPolygon return true iff the Geometry is a MultiPolygon geometry.
func (g Geometry) IsMultiPolygon() bool { return g.tag == multiPolygonTag }

func (g Geometry) check(tag geometryTag) {
	if g.tag != tag {
		panic(fmt.Sprintf("called As%s on Geometry containing %s", tag, g.tag))
	}
}

// AsGeometryCollection returns the geometry as a GeometryCollection. It panics
// if the geometry is not a GeometryCollection.
func (g Geometry) AsGeometryCollection() GeometryCollection {
	g.check(geometryCollectionTag)
	if g.ptr == nil {
		// Special case so that the zero Geometry value is interpreted as an
		// empty GeometryCollection.
		return NewEmptyGeometryCollection(XYOnly)
	}
	return *(*GeometryCollection)(g.ptr)
}

// AsPoint returns the geometry as a Point. It panics if the geometry is not a
// Point.
func (g Geometry) AsPoint() Point {
	g.check(pointTag)
	return *(*Point)(g.ptr)
}

// AsLine returns the geometry as a Line. It panics if the geometry is not a
// Line.
func (g Geometry) AsLine() Line {
	g.check(lineTag)
	return *(*Line)(g.ptr)
}

// AsLineString returns the geometry as a LineString. It panics if the geometry
// is not a LineString.
func (g Geometry) AsLineString() LineString {
	g.check(lineStringTag)
	return *(*LineString)(g.ptr)
}

// AsPolygon returns the geometry as a Polygon. It panics if the geometry is
// not a Polygon.
func (g Geometry) AsPolygon() Polygon {
	g.check(polygonTag)
	return *(*Polygon)(g.ptr)
}

// AsMultiPoint returns the geometry as a MultiPoint. It panics if the geometry
// is not a MultiPoint.
func (g Geometry) AsMultiPoint() MultiPoint {
	g.check(multiPointTag)
	return *(*MultiPoint)(g.ptr)
}

// AsMultiLineString returns the geometry as a MultiLineString. It panics if
// the geometry is not a MultiLineString.
func (g Geometry) AsMultiLineString() MultiLineString {
	g.check(multiLineStringTag)
	return *(*MultiLineString)(g.ptr)
}

// AsMultiPolygon returns the geometry as a MultiPolygon. It panics if the
// Geometry is not a MultiPolygon.
func (g Geometry) AsMultiPolygon() MultiPolygon {
	g.check(multiPolygonTag)
	return *(*MultiPolygon)(g.ptr)
}

// AsText returns the WKT representation of the geometry.
func (g Geometry) AsText() string {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().AsText()
	case pointTag:
		return g.AsPoint().AsText()
	case lineTag:
		return g.AsLine().AsText()
	case lineStringTag:
		return g.AsLineString().AsText()
	case polygonTag:
		return g.AsPolygon().AsText()
	case multiPointTag:
		return g.AsMultiPoint().AsText()
	case multiLineStringTag:
		return g.AsMultiLineString().AsText()
	case multiPolygonTag:
		return g.AsMultiPolygon().AsText()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// MarshalJSON implements the encoding/json.Marshaller interface by returning a
// GeoJSON geometry object.
func (g Geometry) MarshalJSON() ([]byte, error) {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().MarshalJSON()
	case pointTag:
		return g.AsPoint().MarshalJSON()
	case lineTag:
		return g.AsLine().MarshalJSON()
	case lineStringTag:
		return g.AsLineString().MarshalJSON()
	case polygonTag:
		return g.AsPolygon().MarshalJSON()
	case multiPointTag:
		return g.AsMultiPoint().MarshalJSON()
	case multiLineStringTag:
		return g.AsMultiLineString().MarshalJSON()
	case multiPolygonTag:
		return g.AsMultiPolygon().MarshalJSON()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// UnmarshalJSON implements the encoding/json.Unmarshaller interface by
// parsing the JSON stream as GeoJSON geometry object.
//
// It constructs the resultant geometry with no ConstructionOptions. If
// ConstructionOptions are needed, then the value should be unmarshalled into a
// json.RawMessage value and then UnmarshalJSON called manually (passing in the
// ConstructionOptions as desired).
func (g *Geometry) UnmarshalJSON(p []byte) error {
	geom, err := UnmarshalGeoJSON(p)
	if err != nil {
		return err
	}
	*g = geom
	return nil
}

func (g Geometry) appendWKT(dst []byte) []byte {
	switch g.tag {
	case geometryCollectionTag:
		return (*GeometryCollection)(g.ptr).AppendWKT(dst)
	case pointTag:
		return (*Point)(g.ptr).AppendWKT(dst)
	case lineTag:
		return (*Line)(g.ptr).AppendWKT(dst)
	case lineStringTag:
		return (*LineString)(g.ptr).AppendWKT(dst)
	case polygonTag:
		return (*Polygon)(g.ptr).AppendWKT(dst)
	case multiPointTag:
		return (*MultiPoint)(g.ptr).AppendWKT(dst)
	case multiLineStringTag:
		return (*MultiLineString)(g.ptr).AppendWKT(dst)
	case multiPolygonTag:
		return (*MultiPolygon)(g.ptr).AppendWKT(dst)
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// AsBinary writes the WKB (Well Known Binary) representation of the geometry
// to the writer.
func (g Geometry) AsBinary(w io.Writer) error {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().AsBinary(w)
	case pointTag:
		return g.AsPoint().AsBinary(w)
	case lineTag:
		return g.AsLine().AsBinary(w)
	case lineStringTag:
		return g.AsLineString().AsBinary(w)
	case polygonTag:
		return g.AsPolygon().AsBinary(w)
	case multiPointTag:
		return g.AsMultiPoint().AsBinary(w)
	case multiLineStringTag:
		return g.AsMultiLineString().AsBinary(w)
	case multiPolygonTag:
		return g.AsMultiPolygon().AsBinary(w)
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (g Geometry) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := g.AsBinary(&buf)
	return buf.Bytes(), err
}

// Scan implements the database/sql.Scanner interface by parsing the src value
// as WKB (Well Known Binary).
//
// It constructs the resultant geometry with no ConstructionOptions. If
// ConstructionOptions are needed, then the value should be scanned into a byte
// slice and then UnmarshalWKB called manually (passing in the
// ConstructionOptions as desired).
func (g *Geometry) Scan(src interface{}) error {
	var r io.Reader
	switch src := src.(type) {
	case []byte:
		r = bytes.NewReader(src)
	case string:
		r = strings.NewReader(src)
	default:
		// nil is specifically not supported. It _could_ map to an empty
		// geometry, however then the caller wouldn't be able to differentiate
		// between a real empty geometry and a NULL. Instead, we should
		// additionally provide a NullableGeometry type with an IsValid flag.
		return fmt.Errorf("unsupported src type in Scan: %T", src)
	}

	unmarshalled, err := UnmarshalWKB(r)
	if err != nil {
		return err
	}
	*g = unmarshalled
	return nil
}

// Dimension returns the dimension of the Geometry. This is  0 for points, 1
// for curves, and 2 for surfaces (regardless of whether or not they are
// empty). For GeometryCollections it is the maximum dimension over the
// collection (or 0 if the collection is the empty collection).
func (g Geometry) Dimension() int {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().Dimension()
	case pointTag, multiPointTag:
		return 0
	case lineTag, lineStringTag, multiLineStringTag:
		return 1
	case polygonTag, multiPolygonTag:
		return 2
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// IsEmpty returns true if this object an empty geometry.
func (g Geometry) IsEmpty() bool {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().IsEmpty()
	case pointTag:
		return g.AsPoint().IsEmpty()
	case lineTag:
		return false
	case lineStringTag:
		return g.AsLineString().IsEmpty()
	case polygonTag:
		return g.AsPolygon().IsEmpty()
	case multiPointTag:
		return g.AsMultiPoint().IsEmpty()
	case multiLineStringTag:
		return g.AsMultiLineString().IsEmpty()
	case multiPolygonTag:
		return g.AsMultiPolygon().IsEmpty()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// Envelope returns the axis aligned bounding box that most tightly surrounds
// the geometry. Envelopes are not defined for empty geometries, in which case
// the returned flag will be false.
func (g Geometry) Envelope() (Envelope, bool) {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().Envelope()
	case pointTag:
		return g.AsPoint().Envelope()
	case lineTag:
		return g.AsLine().Envelope(), true
	case lineStringTag:
		return g.AsLineString().Envelope()
	case polygonTag:
		return g.AsPolygon().Envelope()
	case multiPointTag:
		return g.AsMultiPoint().Envelope()
	case multiLineStringTag:
		return g.AsMultiLineString().Envelope()
	case multiPolygonTag:
		return g.AsMultiPolygon().Envelope()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// Boundary returns the Geometry representing the limit of this geometry.
func (g Geometry) Boundary() Geometry {
	// TODO: Investigate to see if the behaviour from libgeos would make more
	// sense to use here.
	if g.IsEmpty() {
		// Match PostGIS behaviour.
		return g
	}

	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().Boundary().AsGeometry()
	case pointTag:
		return g.AsPoint().Boundary().AsGeometry()
	case lineTag:
		return g.AsLine().Boundary().AsGeometry()
	case lineStringTag:
		return g.AsLineString().Boundary().AsGeometry()
	case polygonTag:
		mls := g.AsPolygon().Boundary()
		// Ensure holeless polygons return a LineString boundary.
		if mls.NumLineStrings() == 1 {
			return mls.LineStringN(0).AsGeometry()
		}
		return mls.AsGeometry()
	case multiPointTag:
		return g.AsMultiPoint().Boundary().AsGeometry()
	case multiLineStringTag:
		return g.AsMultiLineString().Boundary().AsGeometry()
	case multiPolygonTag:
		return g.AsMultiPolygon().Boundary().AsGeometry()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// EqualsExact checks if this geometry is equal to another geometry from a
// structural pointwise equality perspective. Geometries that are
// structurally equal are defined by exactly same control points in the
// same order. Note that even if two geometries are spatially equal (i.e.
// represent the same point set), they may not be defined by exactly the
// same way. Ordering differences and numeric tolerances can be accounted
// for using options.
func (g Geometry) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().EqualsExact(other, opts...)
	case pointTag:
		return g.AsPoint().EqualsExact(other, opts...)
	case lineTag:
		return g.AsLine().EqualsExact(other, opts...)
	case lineStringTag:
		return g.AsLineString().EqualsExact(other, opts...)
	case polygonTag:
		return g.AsPolygon().EqualsExact(other, opts...)
	case multiPointTag:
		return g.AsMultiPoint().EqualsExact(other, opts...)
	case multiLineStringTag:
		return g.AsMultiLineString().EqualsExact(other, opts...)
	case multiPolygonTag:
		return g.AsMultiPolygon().EqualsExact(other, opts...)
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// Convex hull returns a Geometry that represents the smallest convex set
// that contains this geometry.
func (g Geometry) ConvexHull() Geometry {
	return convexHull(g)
}

// IsValid returns if the current geometry is valid. It is useful to use when
// validation is disabled at constructing, for example, json.Unmarshal
func (g Geometry) IsValid() bool {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().IsValid()
	case pointTag:
		return g.AsPoint().IsValid()
	case lineTag:
		return g.AsLine().IsValid()
	case lineStringTag:
		return g.AsLineString().IsValid()
	case polygonTag:
		return g.AsPolygon().IsValid()
	case multiPointTag:
		return g.AsMultiPoint().IsValid()
	case multiLineStringTag:
		return g.AsMultiLineString().IsValid()
	case multiPolygonTag:
		return g.AsMultiPolygon().IsValid()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// Intersects returns true if the intersection of this gemoetry with the
// specified other geometry is not empty, or false if it is empty.
func (g Geometry) Intersects(other Geometry) bool {
	return hasIntersection(g, other)
}

// Intersection returns a geometric object that represents the point set
// intersection of this geometry with another geometry.
//
// It is not implemented for all possible pairs of geometries, and returns an
// error in those cases.
func (g Geometry) Intersection(other Geometry) (Geometry, error) {
	result, err := intersection(g, other)
	if err != nil {
		return Geometry{}, err
	}
	return result, nil
}

// TransformXY transforms this Geometry into another geometry according the
// mapping provided by the XY function. Some classes of mappings (such as
// affine transformations) will preserve the validity this Geometry in the
// transformed Geometry, in which case no error will be returned. Other
// types of transformations may result in a validation error if their
// mapping results in an invalid Geometry.
func (g Geometry) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().TransformXY(fn, opts...)
	case pointTag:
		return g.AsPoint().TransformXY(fn, opts...)
	case lineTag:
		return g.AsLine().TransformXY(fn, opts...)
	case lineStringTag:
		return g.AsLineString().TransformXY(fn, opts...)
	case polygonTag:
		return g.AsPolygon().TransformXY(fn, opts...)
	case multiPointTag:
		return g.AsMultiPoint().TransformXY(fn, opts...)
	case multiLineStringTag:
		return g.AsMultiLineString().TransformXY(fn, opts...)
	case multiPolygonTag:
		return g.AsMultiPolygon().TransformXY(fn, opts...)
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// Length gives the length of a Line, LineString, or MultiLineString
// or the sum of the lengths of the components of a GeometryCollection.
// Other Geometries are defined to return a length of zero.
func (g Geometry) Length() float64 {
	switch {
	case g.IsEmpty():
		return 0
	case g.IsGeometryCollection():
		return g.AsGeometryCollection().Length()
	case g.IsLine():
		return g.AsLine().Length()
	case g.IsLineString():
		return g.AsLineString().Length()
	case g.IsMultiLineString():
		return g.AsMultiLineString().Length()
	case g.IsPoint():
		return 0
	case g.IsMultiPoint():
		return 0
	case g.IsPolygon():
		return 0
	case g.IsMultiPolygon():
		return 0
	default:
		return 0
	}
}

// Centroid returns the geometry's centroid Point. If the Geometry is empty,
// then an empty Point is returned.
func (g Geometry) Centroid() Point {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().Centroid()
	case pointTag:
		return g.AsPoint().Centroid()
	case lineTag:
		return g.AsLine().Centroid()
	case lineStringTag:
		return g.AsLineString().Centroid()
	case polygonTag:
		return g.AsPolygon().Centroid()
	case multiPointTag:
		return g.AsMultiPoint().Centroid()
	case multiLineStringTag:
		return g.AsMultiLineString().Centroid()
	case multiPolygonTag:
		return g.AsMultiPolygon().Centroid()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// Area gives the area of the Polygon or MultiPolygon or GeometryCollection.
// If the Geometry is none of those types, then 0 is returned.
func (g Geometry) Area() float64 {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().Area()
	case pointTag:
		return 0
	case lineTag:
		return 0
	case lineStringTag:
		return 0
	case polygonTag:
		return g.AsPolygon().Area()
	case multiPointTag:
		return 0
	case multiLineStringTag:
		return 0
	case multiPolygonTag:
		return g.AsMultiPolygon().Area()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// IsSimple calculates whether or not the geometry contains any anomalous
// geometric points such as self intersection or self tangency. For details
// about the precise definition for each type of geometry, see the IsSimple
// method documentation on that type. It is not defined for
// GeometryCollections, in which case false is returned.
func (g Geometry) IsSimple() (isSimple bool, wellDefined bool) {
	switch g.tag {
	case geometryCollectionTag:
		return false, false
	case pointTag:
		return g.AsPoint().IsSimple(), true
	case lineTag:
		return g.AsLine().IsSimple(), true
	case lineStringTag:
		return g.AsLineString().IsSimple(), true
	case polygonTag:
		return g.AsPolygon().IsSimple(), true
	case multiPointTag:
		return g.AsMultiPoint().IsSimple(), true
	case multiLineStringTag:
		return g.AsMultiLineString().IsSimple(), true
	case multiPolygonTag:
		return g.AsMultiPolygon().IsSimple(), true
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// Reverse returns a new geometry containing coordinates listed in reverse order.
// Multi component geometries do not reverse the order of their components,
// but merely reverse each component's coordinates in place.
func (g Geometry) Reverse() Geometry {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().Reverse().AsGeometry()
	case pointTag:
		return g.AsPoint().Reverse().AsGeometry()
	case lineTag:
		return g.AsLine().Reverse().AsGeometry()
	case lineStringTag:
		return g.AsLineString().Reverse().AsGeometry()
	case polygonTag:
		return g.AsPolygon().Reverse().AsGeometry()
	case multiPointTag:
		return g.AsMultiPoint().Reverse().AsGeometry()
	case multiLineStringTag:
		return g.AsMultiLineString().Reverse().AsGeometry()
	case multiPolygonTag:
		return g.AsMultiPolygon().Reverse().AsGeometry()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

func (g Geometry) CoordinatesType() CoordinatesType {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection().CoordinatesType()
	case pointTag:
		return g.AsPoint().CoordinatesType()
	case lineTag:
		return g.AsLine().CoordinatesType()
	case lineStringTag:
		return g.AsLineString().CoordinatesType()
	case polygonTag:
		return g.AsPolygon().CoordinatesType()
	case multiPointTag:
		return g.AsMultiPoint().CoordinatesType()
	case multiLineStringTag:
		return g.AsMultiLineString().CoordinatesType()
	case multiPolygonTag:
		return g.AsMultiPolygon().CoordinatesType()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}
