package geom

import (
	"fmt"
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
		return (*GeometryCollection)(g.ptr).AsText()
	case emptySetTag:
		return (*EmptySet)(g.ptr).AsText()
	case pointTag:
		return (*Point)(g.ptr).AsText()
	case lineTag:
		return (*Line)(g.ptr).AsText()
	case lineStringTag:
		return (*LineString)(g.ptr).AsText()
	case polygonTag:
		return (*Polygon)(g.ptr).AsText()
	case multiPointTag:
		return (*MultiPoint)(g.ptr).AsText()
	case multiLineStringTag:
		return (*MultiLineString)(g.ptr).AsText()
	case multiPolygonTag:
		return (*MultiPolygon)(g.ptr).AsText()
	default:
		panic("unknown geometry: " + g.tag.String())
	}
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
