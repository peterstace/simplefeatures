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
	emptySetTag
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
		emptySetTag:           "EmptySet",
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

// IsGeometryCollection return true iff the Geometry is a GeometryCollection geometry.
func (g Geometry) IsGeometryCollection() bool { return g.tag == geometryCollectionTag }

// IsEmptySet return true iff the Geometry is a EmptySet geometry.
func (g Geometry) IsEmptySet() bool { return g.tag == emptySetTag }

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
		return NewGeometryCollection(nil)
	}
	return *(*GeometryCollection)(g.ptr)
}

// AsEmptySet returns the geometry as a EmptySet. It panics if the geometry is
// not a EmptySet.
func (g Geometry) AsEmptySet() EmptySet {
	g.check(emptySetTag)
	return *(*EmptySet)(g.ptr)
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
	case emptySetTag:
		return g.AsEmptySet().AsText()
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
	case emptySetTag:
		return g.AsEmptySet().MarshalJSON()
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

// AsGeometryX is a temporary helper function to convert to the
// soon-to-be-deleted GeometryX type.
func (g Geometry) AsGeometryX() GeometryX {
	switch g.tag {
	case geometryCollectionTag:
		return g.AsGeometryCollection()
	case emptySetTag:
		return g.AsEmptySet()
	case pointTag:
		return g.AsPoint()
	case lineTag:
		return g.AsLine()
	case lineStringTag:
		return g.AsLineString()
	case polygonTag:
		return g.AsPolygon()
	case multiPointTag:
		return g.AsMultiPoint()
	case multiLineStringTag:
		return g.AsMultiLineString()
	case multiPolygonTag:
		return g.AsMultiPolygon()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
}

// ToGeometry is a temporary helper function to convert the soon-to-be deleted
// GeometryX type to Geometry.
func ToGeometry(g GeometryX) Geometry {
	switch g := g.(type) {
	case GeometryCollection:
		return Geometry{tag: geometryCollectionTag, ptr: unsafe.Pointer(&g)}
	case EmptySet:
		return Geometry{tag: emptySetTag, ptr: unsafe.Pointer(&g)}
	case Point:
		return Geometry{tag: pointTag, ptr: unsafe.Pointer(&g)}
	case Line:
		return Geometry{tag: lineTag, ptr: unsafe.Pointer(&g)}
	case LineString:
		return Geometry{tag: lineStringTag, ptr: unsafe.Pointer(&g)}
	case Polygon:
		return Geometry{tag: polygonTag, ptr: unsafe.Pointer(&g)}
	case MultiPoint:
		return Geometry{tag: multiPointTag, ptr: unsafe.Pointer(&g)}
	case MultiLineString:
		return Geometry{tag: multiLineStringTag, ptr: unsafe.Pointer(&g)}
	case MultiPolygon:
		return Geometry{tag: multiPolygonTag, ptr: unsafe.Pointer(&g)}
	default:
		panic(fmt.Sprintf("unknown type: %T", g))
	}
}

func (g Geometry) appendWKT(dst []byte) []byte {
	switch g.tag {
	case geometryCollectionTag:
		return (*GeometryCollection)(g.ptr).AppendWKT(dst)
	case emptySetTag:
		return (*EmptySet)(g.ptr).AppendWKT(dst)
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
	case emptySetTag:
		return g.AsEmptySet().AsBinary(w)
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
	case emptySetTag:
		return g.AsEmptySet().Dimension()
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
	case emptySetTag:
		return true
	case pointTag, lineTag, lineStringTag, polygonTag:
		return false
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

// Envelope returns the axis aligned bounding box that most tightly
// surrounds the geometry. Envelopes are not defined for empty geometries,
// in which case the returned flag will be false.
func (g Geometry) Envelope() (Envelope, bool) {
	// TODO
	return g.AsGeometryX().Envelope()
}

// Boundary returns the GeometryX representing the limit of this geometry.
func (g Geometry) Boundary() Geometry {
	// TODO
	return ToGeometry(g.AsGeometryX().Boundary())
}

// EqualsExact checks if this geometry is equal to another geometry from a
// structural pointwise equality perspective. Geometries that are
// structurally equal are defined by exactly same control points in the
// same order. Note that even if two geometries are spatially equal (i.e.
// represent the same point set), they may not be defined by exactly the
// same way. Ordering differences and numeric tolerances can be accounted
// for using options.
func (g Geometry) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	// TODO
	return g.AsGeometryX().EqualsExact(other.AsGeometryX(), opts...)
}

// Equals checks if this geometry is equal to another geometry. Two
// geometries are equal if they contain exactly the same points.
//
// It is not implemented for all possible pairs of geometries, and returns
// an error in those cases.
func (g Geometry) Equals(other Geometry) (bool, error) {
	// TODO
	return g.AsGeometryX().Equals(other.AsGeometryX())
}

// Convex hull returns a GeometryX that represents the smallest convex set
// that contains this geometry.
func (g Geometry) ConvexHull() Geometry {
	return convexHull(g)
}

// IsValid returns if the current geometry is valid. It is useful to use when
// validation is disabled at constructing, for example, json.Unmarshal
func (g Geometry) IsValid() bool {
	// TODO
	return g.AsGeometryX().IsValid()
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
	result, err := intersection(g.AsGeometryX(), other.AsGeometryX())
	if err != nil {
		return Geometry{}, err
	}
	return ToGeometry(result), nil
}
