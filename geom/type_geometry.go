package geom

import (
	"database/sql/driver"
	"fmt"
)

// Geometry is a single geometry of any type. Its zero value is valid and is an
// empty [GeometryCollection]. It is immutable after creation.
//
// Note: The internal representation is designed such that [reflect.DeepEqual]
// produces the same result as [ExactEquals] when called with no options.
type Geometry struct {
	impl interface {
		Type() GeometryType
		AsText() string
		MarshalJSON() ([]byte, error)
		AppendWKT(dst []byte) []byte
		AppendWKB(dst []byte) []byte
		IsEmpty() bool
		Envelope() Envelope
		Centroid() Point
		CoordinatesType() CoordinatesType
		Validate() error
		Summary() string
	}
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

// Type returns the type of the Geometry.
func (g Geometry) Type() GeometryType {
	if g.impl == nil {
		return TypeGeometryCollection
	}
	return g.impl.Type()
}

// IsGeometryCollection return true iff the Geometry is a GeometryCollection geometry.
func (g Geometry) IsGeometryCollection() bool { return g.Type() == TypeGeometryCollection }

// IsPoint return true iff the Geometry is a Point geometry.
func (g Geometry) IsPoint() bool { return g.Type() == TypePoint }

// IsLineString return true iff the Geometry is a LineString geometry.
func (g Geometry) IsLineString() bool { return g.Type() == TypeLineString }

// IsPolygon return true iff the Geometry is a Polygon geometry.
func (g Geometry) IsPolygon() bool { return g.Type() == TypePolygon }

// IsMultiPoint return true iff the Geometry is a MultiPoint geometry.
func (g Geometry) IsMultiPoint() bool { return g.Type() == TypeMultiPoint }

// IsMultiLineString return true iff the Geometry is a MultiLineString geometry.
func (g Geometry) IsMultiLineString() bool { return g.Type() == TypeMultiLineString }

// IsMultiPolygon return true iff the Geometry is a MultiPolygon geometry.
func (g Geometry) IsMultiPolygon() bool { return g.Type() == TypeMultiPolygon }

func (g Geometry) check(gtype GeometryType) {
	if g.Type() != gtype {
		panic(fmt.Sprintf("called As%s on Geometry containing %s", gtype, g.Type()))
	}
}

// MustAsGeometryCollection returns the geometry as a GeometryCollection. It
// panics if the geometry is not a GeometryCollection.
func (g Geometry) MustAsGeometryCollection() GeometryCollection {
	if g.impl == nil {
		// Special case so that the zero Geometry value is interpreted as an
		// empty GeometryCollection.
		return GeometryCollection{}
	}
	g.check(TypeGeometryCollection)
	return g.impl.(GeometryCollection)
}

// MustAsPoint returns the geometry as a Point. It panics if the geometry is
// not a Point.
func (g Geometry) MustAsPoint() Point {
	g.check(TypePoint)
	return g.impl.(Point)
}

// MustAsLineString returns the geometry as a LineString. It panics if the
// geometry is not a LineString.
func (g Geometry) MustAsLineString() LineString {
	g.check(TypeLineString)
	return g.impl.(LineString)
}

// MustAsPolygon returns the geometry as a Polygon. It panics if the geometry
// is not a Polygon.
func (g Geometry) MustAsPolygon() Polygon {
	g.check(TypePolygon)
	return g.impl.(Polygon)
}

// MustAsMultiPoint returns the geometry as a MultiPoint. It panics if the
// geometry is not a MultiPoint.
func (g Geometry) MustAsMultiPoint() MultiPoint {
	g.check(TypeMultiPoint)
	return g.impl.(MultiPoint)
}

// MustAsMultiLineString returns the geometry as a MultiLineString. It panics
// if the geometry is not a MultiLineString.
func (g Geometry) MustAsMultiLineString() MultiLineString {
	g.check(TypeMultiLineString)
	return g.impl.(MultiLineString)
}

// MustAsMultiPolygon returns the geometry as a MultiPolygon. It panics if the
// Geometry is not a MultiPolygon.
func (g Geometry) MustAsMultiPolygon() MultiPolygon {
	g.check(TypeMultiPolygon)
	return g.impl.(MultiPolygon)
}

// AsGeometryCollection checks if the geometry is a GeometryCollection, and
// returns it as a GeometryCollection if it is. The returned flag indicates if
// the conversion was successful.
func (g Geometry) AsGeometryCollection() (GeometryCollection, bool) {
	if !g.IsGeometryCollection() {
		return GeometryCollection{}, false
	}
	return g.MustAsGeometryCollection(), true
}

// AsPoint checks if the geometry is a Point, and returns it as a Point if it
// is. The returned flag indicates if the conversion was successful.
func (g Geometry) AsPoint() (Point, bool) {
	if !g.IsPoint() {
		return Point{}, false
	}
	return g.MustAsPoint(), true
}

// AsLineString checks if the geometry is a LineString, and returns it as a
// LineString if it is. The returned flag indicates if the conversion was
// successful.
func (g Geometry) AsLineString() (LineString, bool) {
	if !g.IsLineString() {
		return LineString{}, false
	}
	return g.MustAsLineString(), true
}

// AsPolygon checks if the geometry is a Polygon, and returns it as a Polygon
// if it is. The returned flag indicates if the conversion was successful.
func (g Geometry) AsPolygon() (Polygon, bool) {
	if !g.IsPolygon() {
		return Polygon{}, false
	}
	return g.MustAsPolygon(), true
}

// AsMultiPoint checks if the geometry is a MultiPoint, and returns it as a
// MultiPoint if it is. The returned flag indicates if the conversion was
// successful.
func (g Geometry) AsMultiPoint() (MultiPoint, bool) {
	if !g.IsMultiPoint() {
		return MultiPoint{}, false
	}
	return g.MustAsMultiPoint(), true
}

// AsMultiLineString checks if the geometry is a MultiLineString, and returns
// it as a MultiLineString if it is. The returned flag indicates if the
// conversion was successful.
func (g Geometry) AsMultiLineString() (MultiLineString, bool) {
	if !g.IsMultiLineString() {
		return MultiLineString{}, false
	}
	return g.MustAsMultiLineString(), true
}

// AsMultiPolygon checks if the geometry is a MultiPolygon, and returns it as a
// MultiPolygon if it is. The returned flag indicates if the conversion was
// successful.
func (g Geometry) AsMultiPolygon() (MultiPolygon, bool) {
	if !g.IsMultiPolygon() {
		return MultiPolygon{}, false
	}
	return g.MustAsMultiPolygon(), true
}

// AsText returns the WKT (Well Known Text) representation of this geometry.
func (g Geometry) AsText() string {
	if g.impl == nil {
		return GeometryCollection{}.AsText()
	}
	return g.impl.AsText()
}

// MarshalJSON implements the encoding/json.Marshaler interface by encoding
// this geometry as a GeoJSON geometry object.
func (g Geometry) MarshalJSON() ([]byte, error) {
	if g.impl == nil {
		return GeometryCollection{}.MarshalJSON()
	}
	return g.impl.MarshalJSON()
}

// UnmarshalJSON implements the encoding/json.Unmarshaller interface by
// parsing the JSON stream as GeoJSON geometry object.
//
// Geometry constraint validation is performed on the resultant geometry (an
// error will be returned if the geometry is invalid). If this validation isn't
// needed or is undesirable, then the GeoJSON value should be scanned into a
// json.RawMessage value and then UnmarshalJSON called manually (passing in
// NoValidate{}).
func (g *Geometry) UnmarshalJSON(p []byte) error {
	geom, err := UnmarshalGeoJSON(p)
	if err != nil {
		return err
	}
	*g = geom
	return nil
}

// unmarshalGeoJSONAsType unmarshals GeoJSON directly into the concrete
// geometry specified by dst (which should be a pointer to the concrete
// geometry type).
func unmarshalGeoJSONAsType(p []byte, dst any) error {
	g, err := UnmarshalGeoJSON(p)
	if err != nil {
		return err
	}
	dstType := dst.(interface{ Type() GeometryType }).Type()
	if g.Type() != dstType {
		return unmarshalGeoJSONSourceDestinationMismatchError{
			SourceType:      g.Type(),
			DestinationType: dstType,
		}
	}
	assignToConcrete(dst, g)
	return nil
}

// AppendWKT appends the WKT (Well Known Text) representation of this geometry
// to the input byte slice.
func (g Geometry) AppendWKT(dst []byte) []byte {
	if g.impl == nil {
		return GeometryCollection{}.AppendWKT(dst)
	}
	return g.impl.AppendWKT(dst)
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (g Geometry) AsBinary() []byte {
	return g.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (g Geometry) AppendWKB(dst []byte) []byte {
	if g.impl == nil {
		return GeometryCollection{}.AppendWKB(dst)
	}
	return g.impl.AppendWKB(dst)
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (g Geometry) Value() (driver.Value, error) {
	return g.AsBinary(), nil
}

// Scan implements the database/sql.Scanner interface by parsing the src value
// as WKB (Well Known Binary).
//
// Geometry constraint validation is performed on the resultant geometry (an
// error will be returned if the geometry is invalid). If this validation isn't
// needed or is undesirable, then the WKB should be scanned into a byte slice
// and then UnmarshalWKB called manually (passing in NoValidate{}).
func (g *Geometry) Scan(src any) error {
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
func scanAsType(src any, dst any) error {
	var g Geometry
	if err := g.Scan(src); err != nil {
		return err
	}
	dstType := dst.(interface{ Type() GeometryType }).Type()
	if g.Type() != dstType {
		return fmt.Errorf("scanned geometry is a %s rather than a %s", g.Type(), dstType)
	}
	assignToConcrete(dst, g)
	return nil
}

// assignToConcrete assigns the geometry stored in g to the concrete geometry
// pointed to by dst (i.e. dst must be a pointer to a concrete geometry). It
// panics if the type of dst doesn't match the geometry stored in g.
func assignToConcrete(dst any, g Geometry) {
	switch g.Type() {
	case TypeGeometryCollection:
		*dst.(*GeometryCollection) = g.MustAsGeometryCollection()
	case TypePoint:
		*dst.(*Point) = g.MustAsPoint()
	case TypeLineString:
		*dst.(*LineString) = g.MustAsLineString()
	case TypePolygon:
		*dst.(*Polygon) = g.MustAsPolygon()
	case TypeMultiPoint:
		*dst.(*MultiPoint) = g.MustAsMultiPoint()
	case TypeMultiLineString:
		*dst.(*MultiLineString) = g.MustAsMultiLineString()
	case TypeMultiPolygon:
		*dst.(*MultiPolygon) = g.MustAsMultiPolygon()
	default:
		panic("unknown geometry type: " + g.Type().String())
	}
}

// Dimension returns the dimension of the Geometry. This is  0 for Points and
// MultiPoints, 1 for LineStrings and MultiLineStrings, and 2 for Polygons and
// MultiPolygons (regardless of whether or not they are empty). For
// GeometryCollections it is the maximum dimension over the collection (or 0 if
// the collection is the empty collection).
func (g Geometry) Dimension() int {
	switch g.Type() {
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().Dimension()
	case TypePoint, TypeMultiPoint:
		return 0
	case TypeLineString, TypeMultiLineString:
		return 1
	case TypePolygon, TypeMultiPolygon:
		return 2
	default:
		panic("unknown geometry: " + g.Type().String())
	}
}

// IsEmpty returns true if this geometry is empty. Collection types are empty
// if they have zero elements or only contain empty elements.
func (g Geometry) IsEmpty() bool {
	if g.impl == nil {
		return GeometryCollection{}.IsEmpty()
	}
	return g.impl.IsEmpty()
}

// Envelope returns the axis aligned bounding box that most tightly surrounds
// the geometry.
func (g Geometry) Envelope() Envelope {
	if g.impl == nil {
		return GeometryCollection{}.Envelope()
	}
	return g.impl.Envelope()
}

// Boundary returns the Geometry representing the spatial limit of this
// geometry. The precise definition is dependant on the concrete geometry type
// (see the documentation of each concrete Geometry's Boundary method for
// details).
func (g Geometry) Boundary() Geometry {
	switch g.Type() {
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().Boundary().AsGeometry()
	case TypePoint:
		return g.MustAsPoint().Boundary().AsGeometry()
	case TypeLineString:
		return g.MustAsLineString().Boundary().AsGeometry()
	case TypePolygon:
		mls := g.MustAsPolygon().Boundary()
		// Ensure holeless polygons return a LineString boundary.
		if mls.NumLineStrings() == 1 {
			return mls.LineStringN(0).AsGeometry()
		}
		return mls.AsGeometry()
	case TypeMultiPoint:
		return g.MustAsMultiPoint().Boundary().AsGeometry()
	case TypeMultiLineString:
		return g.MustAsMultiLineString().Boundary().AsGeometry()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().Boundary().AsGeometry()
	default:
		panic("unknown geometry: " + g.Type().String())
	}
}

// ConvexHull returns the geometry representing the smallest convex geometry
// that contains this geometry.
func (g Geometry) ConvexHull() Geometry {
	return convexHull(g)
}

// TransformXY transforms this Geometry into another geometry according the
// mapping provided by the XY function. Because the mapping is arbitrary, it
// has the potential to create an invalid geometry. This can be checked by
// calling the Validate method on the result. Most mappings useful for GIS
// applications will preserve validity.
func (g Geometry) TransformXY(fn func(XY) XY) Geometry {
	switch g.Type() {
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().TransformXY(fn).AsGeometry()
	case TypePoint:
		return g.MustAsPoint().TransformXY(fn).AsGeometry()
	case TypeLineString:
		return g.MustAsLineString().TransformXY(fn).AsGeometry()
	case TypePolygon:
		return g.MustAsPolygon().TransformXY(fn).AsGeometry()
	case TypeMultiPoint:
		return g.MustAsMultiPoint().TransformXY(fn).AsGeometry()
	case TypeMultiLineString:
		return g.MustAsMultiLineString().TransformXY(fn).AsGeometry()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().TransformXY(fn).AsGeometry()
	default:
		panic("unknown geometry: " + g.Type().String())
	}
}

// Transform transforms this Geometry into another Geometry according to the
// provided function, which should transform coordinates in place in the
// float64 slice argument. Coordinates are laid out in order X1, Y1, X2, Y2...
// (for XY geometries), X1, Y1, Z1, X2, Y2, Z2... (for XYZ geometries), X1, Y1,
// M1, X2, Y2, M2... (for XYM geometries), or X1, Y1, Z1, M1, X2, Y2, Z2, M2...
// (for XYZM geometries). The function should return an error if the
// transformation fails.
func (g Geometry) Transform(fn func(CoordinatesType, []float64) error) (Geometry, error) {
	switch g.Type() {
	case TypeGeometryCollection:
		tf, err := g.MustAsGeometryCollection().Transform(fn)
		return tf.AsGeometry(), err
	case TypePoint:
		tf, err := g.MustAsPoint().Transform(fn)
		return tf.AsGeometry(), err
	case TypeLineString:
		tf, err := g.MustAsLineString().Transform(fn)
		return tf.AsGeometry(), err
	case TypePolygon:
		tf, err := g.MustAsPolygon().Transform(fn)
		return tf.AsGeometry(), err
	case TypeMultiPoint:
		tf, err := g.MustAsMultiPoint().Transform(fn)
		return tf.AsGeometry(), err
	case TypeMultiLineString:
		tf, err := g.MustAsMultiLineString().Transform(fn)
		return tf.AsGeometry(), err
	case TypeMultiPolygon:
		tf, err := g.MustAsMultiPolygon().Transform(fn)
		return tf.AsGeometry(), err
	default:
		panic("unknown geometry: " + g.Type().String())
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
		return g.MustAsGeometryCollection().Length()
	case g.IsLineString():
		return g.MustAsLineString().Length()
	case g.IsMultiLineString():
		return g.MustAsMultiLineString().Length()
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
	if g.impl == nil {
		return GeometryCollection{}.Centroid()
	}
	return g.impl.Centroid()
}

// Area gives the area of the Polygon or MultiPolygon or GeometryCollection.
// If the Geometry is none of those types, then 0 is returned.
func (g Geometry) Area(opts ...AreaOption) float64 {
	switch g.Type() {
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().Area(opts...)
	case TypePoint:
		return 0
	case TypeLineString:
		return 0
	case TypePolygon:
		return g.MustAsPolygon().Area(opts...)
	case TypeMultiPoint:
		return 0
	case TypeMultiLineString:
		return 0
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().Area(opts...)
	default:
		panic("unknown geometry: " + g.Type().String())
	}
}

// IsSimple calculates whether or not the geometry contains any anomalous
// geometric points such as self intersection or self tangency. For details
// about the precise definition for each type of geometry, see the IsSimple
// method documentation on that type. It is not defined for
// GeometryCollections, in which case false is returned.
func (g Geometry) IsSimple() (isSimple bool, wellDefined bool) {
	switch g.Type() {
	case TypeGeometryCollection:
		return false, false
	case TypePoint:
		return g.MustAsPoint().IsSimple(), true
	case TypeLineString:
		return g.MustAsLineString().IsSimple(), true
	case TypePolygon:
		return g.MustAsPolygon().IsSimple(), true
	case TypeMultiPoint:
		return g.MustAsMultiPoint().IsSimple(), true
	case TypeMultiLineString:
		return g.MustAsMultiLineString().IsSimple(), true
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().IsSimple(), true
	default:
		panic("unknown geometry: " + g.Type().String())
	}
}

// Reverse returns a new geometry containing coordinates listed in reverse order.
// Multi component geometries do not reverse the order of their components,
// but merely reverse each component's coordinates in place.
func (g Geometry) Reverse() Geometry {
	switch g.Type() {
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().Reverse().AsGeometry()
	case TypePoint:
		return g.MustAsPoint().Reverse().AsGeometry()
	case TypeLineString:
		return g.MustAsLineString().Reverse().AsGeometry()
	case TypePolygon:
		return g.MustAsPolygon().Reverse().AsGeometry()
	case TypeMultiPoint:
		return g.MustAsMultiPoint().Reverse().AsGeometry()
	case TypeMultiLineString:
		return g.MustAsMultiLineString().Reverse().AsGeometry()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().Reverse().AsGeometry()
	default:
		panic("unknown geometry: " + g.Type().String())
	}
}

// CoordinatesType returns the CoordinatesType used to represent points making
// up the geometry.
func (g Geometry) CoordinatesType() CoordinatesType {
	if g.impl == nil {
		return GeometryCollection{}.CoordinatesType()
	}
	return g.impl.CoordinatesType()
}

// ForceCoordinatesType returns a new Geometry with a different CoordinatesType. If a
// dimension is added, then new values are populated with 0.
func (g Geometry) ForceCoordinatesType(newCType CoordinatesType) Geometry {
	switch g.Type() {
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().ForceCoordinatesType(newCType).AsGeometry()
	case TypePoint:
		return g.MustAsPoint().ForceCoordinatesType(newCType).AsGeometry()
	case TypeLineString:
		return g.MustAsLineString().ForceCoordinatesType(newCType).AsGeometry()
	case TypePolygon:
		return g.MustAsPolygon().ForceCoordinatesType(newCType).AsGeometry()
	case TypeMultiPoint:
		return g.MustAsMultiPoint().ForceCoordinatesType(newCType).AsGeometry()
	case TypeMultiLineString:
		return g.MustAsMultiLineString().ForceCoordinatesType(newCType).AsGeometry()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().ForceCoordinatesType(newCType).AsGeometry()
	default:
		panic("unknown geometry: " + g.Type().String())
	}
}

// Force2D returns a copy of the geometry with Z and M values removed.
func (g Geometry) Force2D() Geometry {
	return g.ForceCoordinatesType(DimXY)
}

// PointOnSurface returns a Point that lies inside the geometry.
func (g Geometry) PointOnSurface() Point {
	switch g.Type() {
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().PointOnSurface()
	case TypePoint:
		return g.MustAsPoint().PointOnSurface()
	case TypeLineString:
		return g.MustAsLineString().PointOnSurface()
	case TypePolygon:
		return g.MustAsPolygon().PointOnSurface()
	case TypeMultiPoint:
		return g.MustAsMultiPoint().PointOnSurface()
	case TypeMultiLineString:
		return g.MustAsMultiLineString().PointOnSurface()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().PointOnSurface()
	default:
		panic("unknown geometry: " + g.Type().String())
	}
}

// ForceCW returns the equivalent Geometry that has its exterior rings in a
// clockwise orientation and any inner rings in a counter-clockwise
// orientation. Non-areal geometrys are returned as is.
func (g Geometry) ForceCW() Geometry {
	if g.IsCW() {
		return g
	}
	return g.forceOrientation(true)
}

// ForceCCW returns the equivalent Geometry that has its exterior rings in a
// counter-clockwise orientation and any inner rings in a clockwise
// orientation. Non-areal geometrys are returned as is.
func (g Geometry) ForceCCW() Geometry {
	if g.IsCCW() {
		return g
	}
	return g.forceOrientation(false)
}

func (g Geometry) forceOrientation(forceCW bool) Geometry {
	switch g.Type() {
	case TypePolygon:
		return g.MustAsPolygon().forceOrientation(forceCW).AsGeometry()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().forceOrientation(forceCW).AsGeometry()
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().forceOrientation(forceCW).AsGeometry()
	default:
		return g
	}
}

// IsCW returns true iff the underlying geometry is CW.
// For geometries (such as points) where this is undefined, return true.
func (g Geometry) IsCW() bool {
	switch g.Type() {
	case TypePolygon:
		return g.MustAsPolygon().IsCW()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().IsCW()
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().IsCW()
	default:
		return true
	}
}

// IsCCW returns true iff the underlying geometry is CCW.
// For geometries (such as points) where this is undefined, return true.
func (g Geometry) IsCCW() bool {
	switch g.Type() {
	case TypePolygon:
		return g.MustAsPolygon().IsCCW()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().IsCCW()
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().IsCCW()
	default:
		return true
	}
}

func (g Geometry) controlPoints() int {
	switch g.Type() {
	case TypeGeometryCollection:
		var sum int
		for _, g := range g.MustAsGeometryCollection().geoms {
			sum += g.controlPoints()
		}
		return sum
	case TypePoint:
		return 1
	case TypeLineString:
		return g.MustAsLineString().Coordinates().Length()
	case TypePolygon:
		return g.MustAsPolygon().controlPoints()
	case TypeMultiPoint:
		return g.MustAsMultiPoint().NumPoints()
	case TypeMultiLineString:
		return g.MustAsMultiLineString().controlPoints()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().controlPoints()
	default:
		panic("unknown geometry: " + g.Type().String())
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
		mp := g.MustAsMultiPoint()
		n := mp.NumPoints()
		for i := 0; i < n; i++ {
			gs = append(gs, mp.PointN(i).AsGeometry())
		}
	case TypeMultiLineString:
		mls := g.MustAsMultiLineString()
		n := mls.NumLineStrings()
		for i := 0; i < n; i++ {
			gs = append(gs, mls.LineStringN(i).AsGeometry())
		}
	case TypeMultiPolygon:
		mp := g.MustAsMultiPolygon()
		n := mp.NumPolygons()
		for i := 0; i < n; i++ {
			gs = append(gs, mp.PolygonN(i).AsGeometry())
		}
	case TypeGeometryCollection:
		gc := g.MustAsGeometryCollection()
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
	switch g.Type() {
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().DumpCoordinates()
	case TypePoint:
		return g.MustAsPoint().DumpCoordinates()
	case TypeLineString:
		return g.MustAsLineString().Coordinates()
	case TypePolygon:
		return g.MustAsPolygon().DumpCoordinates()
	case TypeMultiPoint:
		return g.MustAsMultiPoint().Coordinates()
	case TypeMultiLineString:
		return g.MustAsMultiLineString().DumpCoordinates()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().DumpCoordinates()
	default:
		panic("unknown type: " + g.Type().String())
	}
}

// Summary returns a text summary of the Geometry following a similar format to https://postgis.net/docs/ST_Summary.html.
func (g Geometry) Summary() string {
	if g.impl == nil {
		return GeometryCollection{}.Summary()
	}
	return g.impl.Summary()
}

// String returns the string representation of the Geometry.
func (g Geometry) String() string {
	return g.Summary()
}

// Simplify returns a simplified version of the geometry using the
// Ramer-Douglas-Peucker algorithm. Simplification can cause Polygons and
// MultiPolygons to become invalid, in which case an error is returned rather
// than attempting to fix them them. NoValidate{} can be passed in, causing
// this validation to be skipped (potentially resulting in invalid geometries
// being returned).
func (g Geometry) Simplify(threshold float64, nv ...NoValidate) (Geometry, error) {
	switch g.Type() {
	case TypeGeometryCollection:
		c, err := g.MustAsGeometryCollection().Simplify(threshold, nv...)
		return c.AsGeometry(), err
	case TypePoint:
		return g, nil
	case TypeLineString:
		c := g.MustAsLineString().Simplify(threshold)
		return c.AsGeometry(), nil
	case TypePolygon:
		c, err := g.MustAsPolygon().Simplify(threshold, nv...)
		return c.AsGeometry(), err
	case TypeMultiPoint:
		return g, nil
	case TypeMultiLineString:
		return g.MustAsMultiLineString().Simplify(threshold).AsGeometry(), nil
	case TypeMultiPolygon:
		c, err := g.MustAsMultiPolygon().Simplify(threshold, nv...)
		return c.AsGeometry(), err
	default:
		panic("unknown type: " + g.Type().String())
	}
}

// Validate checks if the Geometry is valid. See the documentation for each
// concrete geometry's Validate method for details about the validation rules.
func (g Geometry) Validate() error {
	if g.impl == nil {
		return GeometryCollection{}.Validate()
	}
	return g.impl.Validate()
}

// Densify returns a new Geometry with additional linearly interpolated control
// points such that the distance between any two consecutive control points is
// at most the given maxDistance.
//
// Panics if maxDistance is zero or negative.
func (g Geometry) Densify(maxDistance float64) Geometry {
	switch g.Type() {
	case TypePoint, TypeMultiPoint:
		// Points cannot be densified, but still check that the max distance is
		// valid for consistency between types.
		if maxDistance <= 0 {
			panic("maxDistance must be positive")
		}
		return g
	case TypeLineString:
		return g.MustAsLineString().Densify(maxDistance).AsGeometry()
	case TypeMultiLineString:
		return g.MustAsMultiLineString().Densify(maxDistance).AsGeometry()
	case TypePolygon:
		return g.MustAsPolygon().Densify(maxDistance).AsGeometry()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().Densify(maxDistance).AsGeometry()
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().Densify(maxDistance).AsGeometry()
	default:
		panic("unknown type: " + g.Type().String())
	}
}

// SnapToGrid returns a copy of the geometry with all coordinates snapped to a
// base 10 grid.
//
// The grid spacing is specified by the number of decimal places to round to
// (with negative decimal places being allowed). E.g., a decimalPlaces value of
// 2 would cause all coordinates to be rounded to the nearest 0.01, and a
// decimalPlaces of -1 would cause all coordinates to be rounded to the nearest
// 10.
//
// Returned geometries may be invalid due to snapping, even if the input
// geometry was valid.
func (g Geometry) SnapToGrid(decimalPlaces int) Geometry {
	switch g.Type() {
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().SnapToGrid(decimalPlaces).AsGeometry()
	case TypePoint:
		return g.MustAsPoint().SnapToGrid(decimalPlaces).AsGeometry()
	case TypeLineString:
		return g.MustAsLineString().SnapToGrid(decimalPlaces).AsGeometry()
	case TypePolygon:
		return g.MustAsPolygon().SnapToGrid(decimalPlaces).AsGeometry()
	case TypeMultiPoint:
		return g.MustAsMultiPoint().SnapToGrid(decimalPlaces).AsGeometry()
	case TypeMultiLineString:
		return g.MustAsMultiLineString().SnapToGrid(decimalPlaces).AsGeometry()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().SnapToGrid(decimalPlaces).AsGeometry()
	default:
		panic("unknown type: " + g.Type().String())
	}
}

// FlipCoordinates returns a new geometry with X and Y coordinates swapped.
func (g Geometry) FlipCoordinates() Geometry {
	switch g.Type() {
	case TypeGeometryCollection:
		return g.MustAsGeometryCollection().FlipCoordinates().AsGeometry()
	case TypePoint:
		return g.MustAsPoint().FlipCoordinates().AsGeometry()
	case TypeLineString:
		return g.MustAsLineString().FlipCoordinates().AsGeometry()
	case TypePolygon:
		return g.MustAsPolygon().FlipCoordinates().AsGeometry()
	case TypeMultiPoint:
		return g.MustAsMultiPoint().FlipCoordinates().AsGeometry()
	case TypeMultiLineString:
		return g.MustAsMultiLineString().FlipCoordinates().AsGeometry()
	case TypeMultiPolygon:
		return g.MustAsMultiPolygon().FlipCoordinates().AsGeometry()
	default:
		panic("unknown geometry: " + g.Type().String())
	}
}
