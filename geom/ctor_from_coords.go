package geom

import "fmt"

// NewPointXY builds a new XY Point from its x and y coordinates.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewPointXY(x, y float64) Point {
	c := Coordinates{XY: XY{x, y}, Type: DimXY}
	return NewPoint(c)
}

// NewPointXYZ builds new XYZ Point from its x, y and z coordinates.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewPointXYZ(x, y, z float64) Point {
	c := Coordinates{XY: XY{x, y}, Z: z, Type: DimXYZ}
	return NewPoint(c)
}

// NewPointXYM builds a new XYM Point from its x, y and m coordinates.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewPointXYM(x, y, m float64) Point {
	c := Coordinates{XY: XY{x, y}, M: m, Type: DimXYM}
	return NewPoint(c)
}

// NewPointXYZM builds a new XYZM Point from its x, y, z and m coordinates.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewPointXYZM(x, y, z, m float64) Point {
	c := Coordinates{XY: XY{x, y}, Z: z, M: m, Type: DimXYZM}
	return NewPoint(c)
}

// NewMultiPointXY builds a new XY MultiPoint from its x and y coordinates, x1,
// y1, x2, y2, ..., xn, yn. If the number of coordinates is not a multiple of 2
// the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiPointXY(xys ...float64) MultiPoint {
	xys = clone1DFloat64s(xys)
	return multiPointFromCoords(xys, DimXY)
}

// NewMultiPointXYZ builds a new XYZ MultiPoint from its x, y and z
// coordinates, x1, y1, z1, x2, y2, z2, ..., xn, yn, zn. If the number of
// coordinates is not a multiple of 3 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiPointXYZ(xyzs ...float64) MultiPoint {
	xyzs = clone1DFloat64s(xyzs)
	return multiPointFromCoords(xyzs, DimXYZ)
}

// NewMultiPointXYM builds a new XYM MultiPoint from its x, y and m
// coordinates, x1, y1, m1, x2, y2, m2, ..., xn, yn, mn. If the number of
// coordinates is not a multiple of 3 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiPointXYM(xyms ...float64) MultiPoint {
	xyms = clone1DFloat64s(xyms)
	return multiPointFromCoords(xyms, DimXYM)
}

// NewMultiPointXYZM builds a new XYZM MultiPoint from its x, y, z and m
// coordinates, x1, y1, z1, m1, x2, y2, z2, m2, ..., xn, yn, zn, mn. If the
// number of coordinates is not a multiple of 4 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiPointXYZM(xyzms ...float64) MultiPoint {
	xyzms = clone1DFloat64s(xyzms)
	return multiPointFromCoords(xyzms, DimXYZM)
}

// NewLineStringXY builds a new XY LineString from its x and y coordinates, x1,
// y1, x2, y2, ..., xn, yn. If the number of coordinates is not a multiple of 2
// the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewLineStringXY(xys ...float64) LineString {
	xys = clone1DFloat64s(xys)
	return lineStringFromCoords(xys, DimXY)
}

// NewLineStringXYZ builds a new XYZ LineString from its x, y and z
// coordinates, x1, y1, z1, x2, y2, z2, ..., xn, yn, zn. If the number of
// coordinates is not a multiple of 3 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewLineStringXYZ(xyzs ...float64) LineString {
	xyzs = clone1DFloat64s(xyzs)
	return lineStringFromCoords(xyzs, DimXYZ)
}

// NewLineStringXYM builds a new XYM LineString from its x, y and m
// coordinates, x1, y1, m1, x2, y2, m2, ..., xn, yn, mn. If the number of
// coordinates is not a multiple of 3 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewLineStringXYM(xyms ...float64) LineString {
	xyms = clone1DFloat64s(xyms)
	return lineStringFromCoords(xyms, DimXYM)
}

// NewLineStringXYZM builds a new XYZM LineString from its x, y, z and m
// coordinates, x1, y1, z1, m1, x2, y2, z2, m2, ..., xn, yn, zn, mn. If the
// number of coordinates is not a multiple of 4 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewLineStringXYZM(xyzms ...float64) LineString {
	xyzms = clone1DFloat64s(xyzms)
	return lineStringFromCoords(xyzms, DimXYZM)
}

// NewMultiLineStringXY builds a new XY MultiLineString from the x and y
// coordinates of its LineStrings, each in its own slice, in the form x1, y1,
// x2, y2, ..., xn, yn. If the number of coordinates for each LineString is not
// a multiple of 2 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiLineStringXY(xys ...[]float64) MultiLineString {
	xys = clone2DFloat64s(xys)
	return multiLineStringFromCoords(xys, DimXY)
}

// NewMultiLineStringXYZ builds a new XYZ MultiLineString from the x, y and z
// coordinates of its LineStrings, each in its own slice, in the form x1, y1,
// z1, x2, y2, z2, ..., xn, yn, zn. If the number of coordinates for each
// LineString is not a multiple of 3 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiLineStringXYZ(xyzs ...[]float64) MultiLineString {
	xyzs = clone2DFloat64s(xyzs)
	return multiLineStringFromCoords(xyzs, DimXYZ)
}

// NewMultiLineStringXYM builds a new XYM MultiLineString from the x, y and m
// coordinates of its LineStrings, each in its own slice, in the form x1, y1,
// m1, x2, y2, m2, ..., xn, yn, mn. If the number of coordinates for each
// LineString is not a multiple of 3 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiLineStringXYM(xyms ...[]float64) MultiLineString {
	xyms = clone2DFloat64s(xyms)
	return multiLineStringFromCoords(xyms, DimXYM)
}

// NewMultiLineStringXYZM builds a new XYZM MultiLineString from the x, y, z
// and m coordinates of its LineStrings, each in its own slice, in the form x1,
// y1, z1, m1, x2, y2, z2, m2, ..., xn, yn, zn, mn. If the number of
// coordinates for each LineString is not a multiple of 4 the function will
// panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiLineStringXYZM(xyzms ...[]float64) MultiLineString {
	xyzms = clone2DFloat64s(xyzms)
	return multiLineStringFromCoords(xyzms, DimXYZM)
}

// NewPolygonXY builds a new XY Polygon from the x and y coordinates of its
// rings, each in its own slice, in the form x1, y1, x2, y2, ..., xn, yn, x1,
// y1 (the first and last coordinates of each ring should be the same). If the
// number of coordinates for each ring is not a multiple of 2 the function will
// panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewPolygonXY(xys ...[]float64) Polygon {
	xys = clone2DFloat64s(xys)
	return polygonFromCoords(xys, DimXY)
}

// NewPolygonXYZ builds a new XYZ Polygon from the x, y and z coordinates of
// its rings, each in its own slice, in the form x1, y1, z1, x2, y2, z2, ...,
// xn, yn, zn, x1, y1, z1 (the first and last coordinates of each ring should
// be the same). If the number of coordinates for each ring is not a multiple
// of 3 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewPolygonXYZ(xyzs ...[]float64) Polygon {
	xyzs = clone2DFloat64s(xyzs)
	return polygonFromCoords(xyzs, DimXYZ)
}

// NewPolygonXYM builds a new XYM Polygon from the x, y and m coordinates of
// its rings, each in its own slice, in the form x1, y1, m1, x2, y2, m2, ...,
// xn, yn, mn, x1, y1, m1 (the first and last coordinates of each ring should
// be the same). If the number of coordinates for each ring is not a multiple
// of 3 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewPolygonXYM(xyms ...[]float64) Polygon {
	xyms = clone2DFloat64s(xyms)
	return polygonFromCoords(xyms, DimXYM)
}

// NewPolygonXYZM builds a new XYZM Polygon from the x, y, z and m coordinates
// of its rings, each in its own slice, in the form x1, y1, z1, m1, x2, y2, z2,
// m2, ..., xn, yn, zn, mn, x1, y1, z1, m1 (the first and last coordinates of
// each ring should be the same). If the number of coordinates for each ring is
// not a multiple of 4 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewPolygonXYZM(xyzms ...[]float64) Polygon {
	xyzms = clone2DFloat64s(xyzms)
	return polygonFromCoords(xyzms, DimXYZM)
}

// NewSingleRingPolygonXY builds a new XY Polygon from the x and y coordinates
// of its exterior ring, in the form x1, y1, x2, y2, ..., xn, yn, x1, y1 (the
// first and last coordinates of the ring should be the same). If the number of
// coordinates is not a multiple of 2 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewSingleRingPolygonXY(xys ...float64) Polygon {
	return NewPolygonXY(xys)
}

// NewSingleRingPolygonXYZ builds a new XYZ Polygon from the x, y and z
// coordinates of its exterior ring, in the form x1, y1, z1, x2, y2, z2, ...,
// xn, yn, zn, x1, y1, z1 (the first and last coordinates of the ring should be
// the same). If the number of coordinates is not a multiple of 3 the function
// will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewSingleRingPolygonXYZ(xyzs ...float64) Polygon {
	return NewPolygonXYZ(xyzs)
}

// NewSingleRingPolygonXYM builds a new XYM Polygon from the x, y and m
// coordinates of its exterior ring, in the form x1, y1, m1, x2, y2, m2, ...,
// xn, yn, mn, x1, y1, m1 (the first and last coordinates of the ring should be
// the same). If the number of coordinates is not a multiple of 3 the function
// will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewSingleRingPolygonXYM(xyms ...float64) Polygon {
	return NewPolygonXYM(xyms)
}

// NewSingleRingPolygonXYZM builds a new XYZM Polygon from the x, y, z and m
// coordinates of its exterior ring, in the form x1, y1, z1, m1, x2, y2, z2,
// m2, ..., xn, yn, zn, mn, x1, y1, z1, m1 (the first and last coordinates of
// the ring should be the same). If the number of coordinates is not a multiple
// of 4 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewSingleRingPolygonXYZM(xyzms ...float64) Polygon {
	return NewPolygonXYZM(xyzms)
}

// NewMultiPolygonXY builds a new XY MultiPolygon from the x and y coordinates
// of its Polygons, each in its own slice of rings. Ring coordinates are in the
// form x1, y1, x2, y2, ..., xn, yn, x1, y1 (the first and last coordinates of
// each ring should be the same). If the number of coordinates for each ring is
// not a multiple of 2 the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiPolygonXY(xys ...[][]float64) MultiPolygon {
	xys = clone3DFloat64s(xys)
	return multiPolygonFromCoords(xys, DimXY)
}

// NewMultiPolygonXYZ builds a new XYZ MultiPolygon from the x, y and z
// coordinates of its Polygons, each in its own slice of rings. Ring
// coordinates are in the form x1, y1, z1, x2, y2, z2, ..., xn, yn, zn, x1, y1,
// z1 (the first and last coordinates of each ring should be the same). If the
// number of coordinates for each ring is not a multiple of 3 the function will
// panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiPolygonXYZ(xyzs ...[][]float64) MultiPolygon {
	xyzs = clone3DFloat64s(xyzs)
	return multiPolygonFromCoords(xyzs, DimXYZ)
}

// NewMultiPolygonXYM builds a new XYM MultiPolygon from the x, y and m
// coordinates of its Polygons, each in its own slice of rings. Ring
// coordinates are in the form x1, y1, m1, x2, y2, m2, ..., xn, yn, mn, x1, y1,
// m1 (the first and last coordinates of each ring should be the same). If the
// number of coordinates for each ring is not a multiple of 3 the function will
// panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiPolygonXYM(xyms ...[][]float64) MultiPolygon {
	xyms = clone3DFloat64s(xyms)
	return multiPolygonFromCoords(xyms, DimXYM)
}

// NewMultiPolygonXYZM builds a new XYZM MultiPolygon from the x, y, z and m
// coordinates of its Polygons, each in its own slice of rings. Ring
// coordinates are in the form x1, y1, z1, m1, x2, y2, z2, m2, ..., xn, yn, zn,
// mn, x1, y1, z1, m1 (the first and last coordinates of each ring should be
// the same). If the number of coordinates for each ring is not a multiple of 4
// the function will panic.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiPolygonXYZM(xyzms ...[][]float64) MultiPolygon {
	xyzms = clone3DFloat64s(xyzms)
	return multiPolygonFromCoords(xyzms, DimXYZM)
}

func clone1DFloat64s(src []float64) []float64 {
	dst := make([]float64, len(src))
	copy(dst, src)
	return dst
}

func clone2DFloat64s(src [][]float64) [][]float64 {
	dst := make([][]float64, len(src))
	for i := range src {
		dst[i] = clone1DFloat64s(src[i])
	}
	return dst
}

func clone3DFloat64s(src [][][]float64) [][][]float64 {
	dst := make([][][]float64, len(src))
	for i := range src {
		dst[i] = clone2DFloat64s(src[i])
	}
	return dst
}

func multiPointFromCoords(coords []float64, ct CoordinatesType) MultiPoint {
	if len(coords) == 0 {
		return MultiPoint{}.ForceCoordinatesType(ct)
	}

	dim := ct.Dimension()
	if len(coords)%dim != 0 {
		msg := fmt.Sprintf(
			"geom: coordinate arguments to %s constructor "+
				"must have a length that is a multiple of %d",
			ct.String(), dim,
		)
		panic(msg)
	}

	n := len(coords) / dim
	pts := make([]Point, n)
	for i := 0; i < n; i++ {
		c := Coordinates{
			XY: XY{
				coords[i*dim+0],
				coords[i*dim+1],
			},
			Type: ct,
		}
		if ct.Is3D() {
			c.Z = coords[i*dim+2]
		}
		if ct.IsMeasured() {
			c.M = coords[i*dim+dim-1]
		}
		pts[i] = NewPoint(c)
	}
	return NewMultiPoint(pts)
}

func lineStringFromCoords(coords []float64, ct CoordinatesType) LineString {
	if len(coords) == 0 {
		return LineString{}.ForceCoordinatesType(ct)
	}
	seq := NewSequence(coords, ct)
	return NewLineString(seq)
}

func lineStringSliceFromCoords(coords [][]float64, ct CoordinatesType) []LineString {
	lss := make([]LineString, len(coords))
	for i := range coords {
		seq := NewSequence(coords[i], ct)
		lss[i] = NewLineString(seq)
	}
	return lss
}

func multiLineStringFromCoords(coords [][]float64, ct CoordinatesType) MultiLineString {
	if len(coords) == 0 {
		return MultiLineString{}.ForceCoordinatesType(ct)
	}
	lss := lineStringSliceFromCoords(coords, ct)
	return NewMultiLineString(lss)
}

func polygonFromCoords(coords [][]float64, ct CoordinatesType) Polygon {
	if len(coords) == 0 {
		return Polygon{}.ForceCoordinatesType(ct)
	}
	rings := lineStringSliceFromCoords(coords, ct)
	return NewPolygon(rings)
}

func multiPolygonFromCoords(coords [][][]float64, ct CoordinatesType) MultiPolygon {
	if len(coords) == 0 {
		return MultiPolygon{}.ForceCoordinatesType(ct)
	}
	polys := make([]Polygon, len(coords))
	for i := range coords {
		polys[i] = polygonFromCoords(coords[i], ct)
	}
	return NewMultiPolygon(polys)
}
