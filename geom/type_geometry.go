package geom

import (
	"database/sql/driver"
	"fmt"
	"unsafe"
)

// Geometry is a single geometry of any type. Its zero value is valid and is
// an empty GeometryCollection. It is immutable after creation.
type Geometry struct {
	gtype GeometryType
	ptr   unsafe.Pointer
}

// GeometryType represents one of the 7 geometry types.
type GeometryType int

const (
	// TypeGeometryCollection is the type of a GeometryCollection.
	TypeGeometryCollection GeometryType = iota
	// TypePoint is the type of a Point.
	TypePoint
	// TypeLineString is the type of a LineString.
	TypeLineString
	// TypePolygon is the type of a Polygon.
	TypePolygon
	// TypeMultiPoint is the type of a MultiPoint.
	TypeMultiPoint
	// TypeMultiLineString is the type of a MultiLineString.
	TypeMultiLineString
	// TypeMultiPolygon is the type of a MultiPolygon.
	TypeMultiPolygon
)

func (t GeometryType) String() string {
	switch t {
	case TypeGeometryCollection:
		return "GeometryCollection"
	case TypePoint:
		return "Point"
	case TypeLineString:
		return "LineString"
	case TypePolygon:
		return "Polygon"
	case TypeMultiPoint:
		return "MultiPoint"
	case TypeMultiLineString:
		return "MultiLineString"
	case TypeMultiPolygon:
		return "MultiPolygon"
	default:
		return "invalid"
	}
}

// Type returns a string representation of the geometry's type.
func (g Geometry) Type() GeometryType {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().Type()
	case TypePoint:
		return g.AsPoint().Type()
	case TypeLineString:
		return g.AsLineString().Type()
	case TypePolygon:
		return g.AsPolygon().Type()
	case TypeMultiPoint:
		return g.AsMultiPoint().Type()
	case TypeMultiLineString:
		return g.AsMultiLineString().Type()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().Type()
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// IsGeometryCollection return true iff the Geometry is a GeometryCollection geometry.
func (g Geometry) IsGeometryCollection() bool { return g.gtype == TypeGeometryCollection }

// IsPoint return true iff the Geometry is a Point geometry.
func (g Geometry) IsPoint() bool { return g.gtype == TypePoint }

// IsLineString return true iff the Geometry is a LineString geometry.
func (g Geometry) IsLineString() bool { return g.gtype == TypeLineString }

// IsPolygon return true iff the Geometry is a Polygon geometry.
func (g Geometry) IsPolygon() bool { return g.gtype == TypePolygon }

// IsMultiPoint return true iff the Geometry is a MultiPoint geometry.
func (g Geometry) IsMultiPoint() bool { return g.gtype == TypeMultiPoint }

// IsMultiLineString return true iff the Geometry is a MultiLineString geometry.
func (g Geometry) IsMultiLineString() bool { return g.gtype == TypeMultiLineString }

// IsMultiPolygon return true iff the Geometry is a MultiPolygon geometry.
func (g Geometry) IsMultiPolygon() bool { return g.gtype == TypeMultiPolygon }

func (g Geometry) check(gtype GeometryType) {
	if g.gtype != gtype {
		panic(fmt.Sprintf("called As%s on Geometry containing %s", gtype, g.gtype))
	}
}

// AsGeometryCollection returns the geometry as a GeometryCollection. It panics
// if the geometry is not a GeometryCollection.
func (g Geometry) AsGeometryCollection() GeometryCollection {
	g.check(TypeGeometryCollection)
	if g.ptr == nil {
		// Special case so that the zero Geometry value is interpreted as an
		// empty GeometryCollection.
		return GeometryCollection{}
	}
	return *(*GeometryCollection)(g.ptr)
}

// AsPoint returns the geometry as a Point. It panics if the geometry is not a
// Point.
func (g Geometry) AsPoint() Point {
	g.check(TypePoint)
	return *(*Point)(g.ptr)
}

// AsLineString returns the geometry as a LineString. It panics if the geometry
// is not a LineString.
func (g Geometry) AsLineString() LineString {
	g.check(TypeLineString)
	return *(*LineString)(g.ptr)
}

// AsPolygon returns the geometry as a Polygon. It panics if the geometry is
// not a Polygon.
func (g Geometry) AsPolygon() Polygon {
	g.check(TypePolygon)
	return *(*Polygon)(g.ptr)
}

// AsMultiPoint returns the geometry as a MultiPoint. It panics if the geometry
// is not a MultiPoint.
func (g Geometry) AsMultiPoint() MultiPoint {
	g.check(TypeMultiPoint)
	return *(*MultiPoint)(g.ptr)
}

// AsMultiLineString returns the geometry as a MultiLineString. It panics if
// the geometry is not a MultiLineString.
func (g Geometry) AsMultiLineString() MultiLineString {
	g.check(TypeMultiLineString)
	return *(*MultiLineString)(g.ptr)
}

// AsMultiPolygon returns the geometry as a MultiPolygon. It panics if the
// Geometry is not a MultiPolygon.
func (g Geometry) AsMultiPolygon() MultiPolygon {
	g.check(TypeMultiPolygon)
	return *(*MultiPolygon)(g.ptr)
}

// AsText returns the WKT (Well Known Text) representation of this geometry.
func (g Geometry) AsText() string {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().AsText()
	case TypePoint:
		return g.AsPoint().AsText()
	case TypeLineString:
		return g.AsLineString().AsText()
	case TypePolygon:
		return g.AsPolygon().AsText()
	case TypeMultiPoint:
		return g.AsMultiPoint().AsText()
	case TypeMultiLineString:
		return g.AsMultiLineString().AsText()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().AsText()
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// MarshalJSON implements the encoding/json.Marshaller interface by encoding
// this geometry as a GeoJSON geometry object.
func (g Geometry) MarshalJSON() ([]byte, error) {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().MarshalJSON()
	case TypePoint:
		return g.AsPoint().MarshalJSON()
	case TypeLineString:
		return g.AsLineString().MarshalJSON()
	case TypePolygon:
		return g.AsPolygon().MarshalJSON()
	case TypeMultiPoint:
		return g.AsMultiPoint().MarshalJSON()
	case TypeMultiLineString:
		return g.AsMultiLineString().MarshalJSON()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().MarshalJSON()
	default:
		panic("unknown geometry: " + g.gtype.String())
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

// AppendWKT appends the WKT (Well Known Text) representation of this geometry
// to the input byte slice.
func (g Geometry) AppendWKT(dst []byte) []byte {
	switch g.gtype {
	case TypeGeometryCollection:
		return (*GeometryCollection)(g.ptr).AppendWKT(dst)
	case TypePoint:
		return (*Point)(g.ptr).AppendWKT(dst)
	case TypeLineString:
		return (*LineString)(g.ptr).AppendWKT(dst)
	case TypePolygon:
		return (*Polygon)(g.ptr).AppendWKT(dst)
	case TypeMultiPoint:
		return (*MultiPoint)(g.ptr).AppendWKT(dst)
	case TypeMultiLineString:
		return (*MultiLineString)(g.ptr).AppendWKT(dst)
	case TypeMultiPolygon:
		return (*MultiPolygon)(g.ptr).AppendWKT(dst)
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (g Geometry) AsBinary() []byte {
	return g.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (g Geometry) AppendWKB(dst []byte) []byte {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().AppendWKB(dst)
	case TypePoint:
		return g.AsPoint().AppendWKB(dst)
	case TypeLineString:
		return g.AsLineString().AppendWKB(dst)
	case TypePolygon:
		return g.AsPolygon().AppendWKB(dst)
	case TypeMultiPoint:
		return g.AsMultiPoint().AppendWKB(dst)
	case TypeMultiLineString:
		return g.AsMultiLineString().AppendWKB(dst)
	case TypeMultiPolygon:
		return g.AsMultiPolygon().AppendWKB(dst)
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (g Geometry) Value() (driver.Value, error) {
	return g.AsBinary(), nil
}

// Scan implements the database/sql.Scanner interface by parsing the src value
// as WKB (Well Known Binary).
//
// It constructs the resultant geometry with no ConstructionOptions. If
// ConstructionOptions are needed, then the value should be scanned into a byte
// slice and then UnmarshalWKB called manually (passing in the
// ConstructionOptions as desired).
func (g *Geometry) Scan(src interface{}) error {
	var wkb []byte
	switch src := src.(type) {
	case []byte:
		wkb = src
	case string:
		wkb = []byte(src)
	default:
		// nil is specifically not supported. It _could_ map to an empty
		// geometry, however then the caller wouldn't be able to differentiate
		// between a real empty geometry and a NULL. Users needing this
		// functionality should use the NullGeometry type.
		return fmt.Errorf("unsupported src type in Scan: %T", src)
	}

	unmarshalled, err := UnmarshalWKB(wkb)
	if err != nil {
		return wrap(err, "scanning as WKB")
	}
	*g = unmarshalled
	return nil
}

// scanAsType helps to implement the sql.Scanner interface for concrete
// geometry types. The src should be the input to Scan, typ should be the
// concrete geometry type, and dst should be a pointer to the concrete geometry
// to update (e.g. *LineString).
func scanAsType(src interface{}, dst interface{}, typ GeometryType) error {
	var g Geometry
	if err := g.Scan(src); err != nil {
		return err
	}
	if g.Type() != typ {
		return fmt.Errorf("scanned geometry is a %s rather than a %s", g.Type(), typ)
	}
	switch typ {
	case TypeGeometryCollection:
		*dst.(*GeometryCollection) = g.AsGeometryCollection()
	case TypePoint:
		*dst.(*Point) = g.AsPoint()
	case TypeLineString:
		*dst.(*LineString) = g.AsLineString()
	case TypePolygon:
		*dst.(*Polygon) = g.AsPolygon()
	case TypeMultiPoint:
		*dst.(*MultiPoint) = g.AsMultiPoint()
	case TypeMultiLineString:
		*dst.(*MultiLineString) = g.AsMultiLineString()
	case TypeMultiPolygon:
		*dst.(*MultiPolygon) = g.AsMultiPolygon()
	default:
		panic("unknown geometry type: " + typ.String())
	}
	return nil
}

// Dimension returns the dimension of the Geometry. This is  0 for Points and
// MultiPoints, 1 for LineStrings and MultiLineStrings, and 2 for Polygons and
// MultiPolygons (regardless of whether or not they are empty). For
// GeometryCollections it is the maximum dimension over the collection (or 0 if
// the collection is the empty collection).
func (g Geometry) Dimension() int {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().Dimension()
	case TypePoint, TypeMultiPoint:
		return 0
	case TypeLineString, TypeMultiLineString:
		return 1
	case TypePolygon, TypeMultiPolygon:
		return 2
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// IsEmpty returns true if this geometry is empty. Collection types are empty
// if they have zero elements or only contain empty elements.
func (g Geometry) IsEmpty() bool {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().IsEmpty()
	case TypePoint:
		return g.AsPoint().IsEmpty()
	case TypeLineString:
		return g.AsLineString().IsEmpty()
	case TypePolygon:
		return g.AsPolygon().IsEmpty()
	case TypeMultiPoint:
		return g.AsMultiPoint().IsEmpty()
	case TypeMultiLineString:
		return g.AsMultiLineString().IsEmpty()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().IsEmpty()
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// Envelope returns the axis aligned bounding box that most tightly surrounds
// the geometry.
func (g Geometry) Envelope() Envelope {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().Envelope()
	case TypePoint:
		return g.AsPoint().Envelope()
	case TypeLineString:
		return g.AsLineString().Envelope()
	case TypePolygon:
		return g.AsPolygon().Envelope()
	case TypeMultiPoint:
		return g.AsMultiPoint().Envelope()
	case TypeMultiLineString:
		return g.AsMultiLineString().Envelope()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().Envelope()
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// Boundary returns the Geometry representing the spatial limit of this
// geometry. The precise definition is dependant on the concrete geometry type
// (see the documentation of each concrete Geometry's Boundary method for
// details).
func (g Geometry) Boundary() Geometry {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().Boundary().AsGeometry()
	case TypePoint:
		return g.AsPoint().Boundary().AsGeometry()
	case TypeLineString:
		return g.AsLineString().Boundary().AsGeometry()
	case TypePolygon:
		mls := g.AsPolygon().Boundary()
		// Ensure holeless polygons return a LineString boundary.
		if mls.NumLineStrings() == 1 {
			return mls.LineStringN(0).AsGeometry()
		}
		return mls.AsGeometry()
	case TypeMultiPoint:
		return g.AsMultiPoint().Boundary().AsGeometry()
	case TypeMultiLineString:
		return g.AsMultiLineString().Boundary().AsGeometry()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().Boundary().AsGeometry()
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// ConvexHull returns the geometry representing the smallest convex geometry
// that contains this geometry.
func (g Geometry) ConvexHull() Geometry {
	return convexHull(g)
}

// TransformXY transforms this Geometry into another geometry according the
// mapping provided by the XY function. Some classes of mappings (such as
// affine transformations) will preserve the validity this Geometry in the
// transformed Geometry, in which case no error will be returned. Other
// types of transformations may result in a validation error if their
// mapping results in an invalid Geometry.
func (g Geometry) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	switch g.gtype {
	case TypeGeometryCollection:
		gt, err := g.AsGeometryCollection().TransformXY(fn, opts...)
		return gt.AsGeometry(), err
	case TypePoint:
		gt, err := g.AsPoint().TransformXY(fn, opts...)
		return gt.AsGeometry(), err
	case TypeLineString:
		gt, err := g.AsLineString().TransformXY(fn, opts...)
		return gt.AsGeometry(), err
	case TypePolygon:
		gt, err := g.AsPolygon().TransformXY(fn, opts...)
		return gt.AsGeometry(), err
	case TypeMultiPoint:
		gt, err := g.AsMultiPoint().TransformXY(fn, opts...)
		return gt.AsGeometry(), err
	case TypeMultiLineString:
		gt, err := g.AsMultiLineString().TransformXY(fn, opts...)
		return gt.AsGeometry(), err
	case TypeMultiPolygon:
		gt, err := g.AsMultiPolygon().TransformXY(fn, opts...)
		return gt.AsGeometry(), err
	default:
		panic("unknown geometry: " + g.gtype.String())
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
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().Centroid()
	case TypePoint:
		return g.AsPoint().Centroid()
	case TypeLineString:
		return g.AsLineString().Centroid()
	case TypePolygon:
		return g.AsPolygon().Centroid()
	case TypeMultiPoint:
		return g.AsMultiPoint().Centroid()
	case TypeMultiLineString:
		return g.AsMultiLineString().Centroid()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().Centroid()
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// Area gives the area of the Polygon or MultiPolygon or GeometryCollection.
// If the Geometry is none of those types, then 0 is returned.
func (g Geometry) Area(opts ...AreaOption) float64 {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().Area(opts...)
	case TypePoint:
		return 0
	case TypeLineString:
		return 0
	case TypePolygon:
		return g.AsPolygon().Area(opts...)
	case TypeMultiPoint:
		return 0
	case TypeMultiLineString:
		return 0
	case TypeMultiPolygon:
		return g.AsMultiPolygon().Area(opts...)
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// IsSimple calculates whether or not the geometry contains any anomalous
// geometric points such as self intersection or self tangency. For details
// about the precise definition for each type of geometry, see the IsSimple
// method documentation on that type. It is not defined for
// GeometryCollections, in which case false is returned.
func (g Geometry) IsSimple() (isSimple bool, wellDefined bool) {
	switch g.gtype {
	case TypeGeometryCollection:
		return false, false
	case TypePoint:
		return g.AsPoint().IsSimple(), true
	case TypeLineString:
		return g.AsLineString().IsSimple(), true
	case TypePolygon:
		return g.AsPolygon().IsSimple(), true
	case TypeMultiPoint:
		return g.AsMultiPoint().IsSimple(), true
	case TypeMultiLineString:
		return g.AsMultiLineString().IsSimple(), true
	case TypeMultiPolygon:
		return g.AsMultiPolygon().IsSimple(), true
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// Reverse returns a new geometry containing coordinates listed in reverse order.
// Multi component geometries do not reverse the order of their components,
// but merely reverse each component's coordinates in place.
func (g Geometry) Reverse() Geometry {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().Reverse().AsGeometry()
	case TypePoint:
		return g.AsPoint().Reverse().AsGeometry()
	case TypeLineString:
		return g.AsLineString().Reverse().AsGeometry()
	case TypePolygon:
		return g.AsPolygon().Reverse().AsGeometry()
	case TypeMultiPoint:
		return g.AsMultiPoint().Reverse().AsGeometry()
	case TypeMultiLineString:
		return g.AsMultiLineString().Reverse().AsGeometry()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().Reverse().AsGeometry()
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// CoordinatesType returns the CoordinatesType used to represent points making
// up the geometry.
func (g Geometry) CoordinatesType() CoordinatesType {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().CoordinatesType()
	case TypePoint:
		return g.AsPoint().CoordinatesType()
	case TypeLineString:
		return g.AsLineString().CoordinatesType()
	case TypePolygon:
		return g.AsPolygon().CoordinatesType()
	case TypeMultiPoint:
		return g.AsMultiPoint().CoordinatesType()
	case TypeMultiLineString:
		return g.AsMultiLineString().CoordinatesType()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().CoordinatesType()
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// ForceCoordinatesType returns a new Geometry with a different CoordinatesType. If a
// dimension is added, then new values are populated with 0.
func (g Geometry) ForceCoordinatesType(newCType CoordinatesType) Geometry {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().ForceCoordinatesType(newCType).AsGeometry()
	case TypePoint:
		return g.AsPoint().ForceCoordinatesType(newCType).AsGeometry()
	case TypeLineString:
		return g.AsLineString().ForceCoordinatesType(newCType).AsGeometry()
	case TypePolygon:
		return g.AsPolygon().ForceCoordinatesType(newCType).AsGeometry()
	case TypeMultiPoint:
		return g.AsMultiPoint().ForceCoordinatesType(newCType).AsGeometry()
	case TypeMultiLineString:
		return g.AsMultiLineString().ForceCoordinatesType(newCType).AsGeometry()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().ForceCoordinatesType(newCType).AsGeometry()
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// Force2D returns a copy of the geometry with Z and M values removed.
func (g Geometry) Force2D() Geometry {
	return g.ForceCoordinatesType(DimXY)
}

// PointOnSurface returns a Point that lies inside the geometry.
func (g Geometry) PointOnSurface() Point {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().PointOnSurface()
	case TypePoint:
		return g.AsPoint().PointOnSurface()
	case TypeLineString:
		return g.AsLineString().PointOnSurface()
	case TypePolygon:
		return g.AsPolygon().PointOnSurface()
	case TypeMultiPoint:
		return g.AsMultiPoint().PointOnSurface()
	case TypeMultiLineString:
		return g.AsMultiLineString().PointOnSurface()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().PointOnSurface()
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// ForceCW returns the equivalent Geometry that has its exterior rings in a
// clockwise orientation and any inner rings in a counter-clockwise
// orientation. Non-areal geometrys are returned as is.
func (g Geometry) ForceCW() Geometry {
	return g.forceOrientation(true)
}

// ForceCCW returns the equivalent Geometry that has its exterior rings in a
// counter-clockwise orientation and any inner rings in a clockwise
// orientation. Non-areal geometrys are returned as is.
func (g Geometry) ForceCCW() Geometry {
	return g.forceOrientation(false)
}

func (g Geometry) forceOrientation(forceCW bool) Geometry {
	switch g.gtype {
	case TypePolygon:
		return g.AsPolygon().forceOrientation(forceCW).AsGeometry()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().forceOrientation(forceCW).AsGeometry()
	case TypeGeometryCollection:
		return g.AsGeometryCollection().forceOrientation(forceCW).AsGeometry()
	default:
		return g
	}
}

func (g Geometry) controlPoints() int {
	switch g.gtype {
	case TypeGeometryCollection:
		var sum int
		for _, g := range g.AsGeometryCollection().geoms {
			sum += g.controlPoints()
		}
		return sum
	case TypePoint:
		return 1
	case TypeLineString:
		return g.AsLineString().Coordinates().Length()
	case TypePolygon:
		return g.AsPolygon().controlPoints()
	case TypeMultiPoint:
		return g.AsMultiPoint().NumPoints()
	case TypeMultiLineString:
		return g.AsMultiLineString().controlPoints()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().controlPoints()
	default:
		panic("unknown geometry: " + g.gtype.String())
	}
}

// Dump breaks multi types (MultiPoints, MultiLineStrings, and MultiPolygons)
// and GeometryCollections into their constituent non-multi types (Points,
// LineStrings, and Polygons).
//
// The returned slice will only ever contain Points, LineStrings, and Polygons.
//
// When called on a Point, LineString, or Polygon is input, the original value
// is returned in a slice of length 1.
func (g Geometry) Dump() []Geometry {
	return g.appendDump(nil)
}

func (g Geometry) appendDump(gs []Geometry) []Geometry {
	switch g.Type() {
	case TypePoint, TypeLineString, TypePolygon:
		gs = append(gs, g)
	case TypeMultiPoint:
		mp := g.AsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			gs = append(gs, mp.PointN(i).AsGeometry())
		}
	case TypeMultiLineString:
		mls := g.AsMultiLineString()
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			gs = append(gs, mls.LineStringN(i).AsGeometry())
		}
	case TypeMultiPolygon:
		mp := g.AsMultiPolygon()
		n := mp.NumPolygons()
		for i := 0; i < n; i++ {
			gs = append(gs, mp.PolygonN(i).AsGeometry())
		}
	case TypeGeometryCollection:
		gc := g.AsGeometryCollection()
		n := gc.NumGeometries()
		for i := 0; i < n; i++ {
			gs = gc.GeometryN(i).appendDump(gs)
		}
	default:
		panic("unknown type: " + g.Type().String())
	}
	return gs
}

// DumpCoordinates returns the control points making up the geometry as a
// Sequence.
func (g Geometry) DumpCoordinates() Sequence {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().DumpCoordinates()
	case TypePoint:
		return g.AsPoint().DumpCoordinates()
	case TypeLineString:
		return g.AsLineString().Coordinates()
	case TypePolygon:
		return g.AsPolygon().DumpCoordinates()
	case TypeMultiPoint:
		return g.AsMultiPoint().Coordinates()
	case TypeMultiLineString:
		return g.AsMultiLineString().DumpCoordinates()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().DumpCoordinates()
	default:
		panic("unknown type: " + g.Type().String())
	}
}

// Summary returns a text summary of the Geometry following a similar format to https://postgis.net/docs/ST_Summary.html.
func (g Geometry) Summary() string {
	switch g.gtype {
	case TypeGeometryCollection:
		return g.AsGeometryCollection().Summary()
	case TypePoint:
		return g.AsPoint().Summary()
	case TypeLineString:
		return g.AsLineString().Summary()
	case TypePolygon:
		return g.AsPolygon().Summary()
	case TypeMultiPoint:
		return g.AsMultiPoint().Summary()
	case TypeMultiLineString:
		return g.AsMultiLineString().Summary()
	case TypeMultiPolygon:
		return g.AsMultiPolygon().Summary()
	default:
		panic("unknown type: " + g.Type().String())
	}
}

// String returns the string representation of the Geometry.
func (g Geometry) String() string {
	return g.Summary()
}
