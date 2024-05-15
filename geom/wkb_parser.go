package geom

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

// UnmarshalWKB reads the Well Known Binary (WKB), and returns the
// corresponding Geometry.
//
// NoValidate{} can be passed in to disable geometry constraint validation.
func UnmarshalWKB(wkb []byte, nv ...NoValidate) (Geometry, error) {
	// Note that we purposefully DON'T check for the presence of trailing
	// bytes. There is nothing in the OGC spec indicating that trailing bytes
	// are illegal. Some Esri software will add (useless) trailing bytes to
	// their WKBs.
	p := wkbParser{body: wkb}
	g, err := p.run()
	if err != nil {
		return Geometry{}, err
	}
	if len(nv) == 0 {
		if err := g.Validate(); err != nil {
			return Geometry{}, err
		}
	}
	return g, nil
}

type wkbParser struct {
	body []byte
	bo   byte
	no   bool
}

func (p *wkbParser) run() (Geometry, error) {
	if err := p.parseByteOrder(); err != nil {
		return Geometry{}, err
	}
	gtype, ctype, err := p.parseGeomAndCoordType()
	if err != nil {
		return Geometry{}, err
	}
	return p.parseGeomRoot(gtype, ctype)
}

func (p *wkbParser) inner() (Geometry, error) {
	inner := wkbParser{body: p.body}
	g, err := inner.run()
	if err != nil {
		return Geometry{}, err
	}
	p.body = inner.body
	return g, nil
}

func (p *wkbParser) parseByteOrder() error {
	b, err := p.readByte()
	if err != nil {
		return err
	}
	p.bo = b
	switch b {
	case 0:
		p.no = (nativeOrder == binary.BigEndian)
		return nil
	case 1:
		p.no = (nativeOrder == binary.LittleEndian)
		return nil
	default:
		return wkbSyntaxError{fmt.Sprintf("invalid byte order: %#x", b)}
	}
}

func (p *wkbParser) readByte() (byte, error) {
	if len(p.body) == 0 {
		return 0, wkbSyntaxError{"unexpected EOF"}
	}
	b := p.body[0]
	p.body = p.body[1:]
	return b, nil
}

func (p *wkbParser) parseUint32() (uint32, error) {
	if len(p.body) < 4 {
		return 0, wkbSyntaxError{"unexpected EOF"}
	}

	var x uint32
	if p.bo == 0 {
		x = binary.BigEndian.Uint32(p.body)
	} else {
		x = binary.LittleEndian.Uint32(p.body)
	}

	p.body = p.body[4:]
	return x, nil
}

func (p *wkbParser) parseGeomAndCoordType() (GeometryType, CoordinatesType, error) {
	geomCode, err := p.parseUint32()
	if err != nil {
		return 0, 0, err
	}
	var gtype GeometryType
	switch geomCode % 1000 {
	case 1:
		gtype = TypePoint
	case 2:
		gtype = TypeLineString
	case 3:
		gtype = TypePolygon
	case 4:
		gtype = TypeMultiPoint
	case 5:
		gtype = TypeMultiLineString
	case 6:
		gtype = TypeMultiPolygon
	case 7:
		gtype = TypeGeometryCollection
	default:
		return 0, 0, wkbSyntaxError{fmt.Sprintf("invalid or unknown geometry type in geom code: %v", geomCode)}
	}

	var ctype CoordinatesType
	switch geomCode / 1000 {
	case 0:
		ctype = DimXY
	case 1:
		ctype = DimXYZ
	case 2:
		ctype = DimXYM
	case 3:
		ctype = DimXYZM
	default:
		return 0, 0, wkbSyntaxError{fmt.Sprintf("invalid coordinates type in geom code: %v", geomCode)}
	}

	return gtype, ctype, nil
}

func (p *wkbParser) parseGeomRoot(gtype GeometryType, ctype CoordinatesType) (Geometry, error) {
	switch gtype {
	case TypePoint:
		pt, err := p.parsePoint(ctype)
		return pt.AsGeometry(), err
	case TypeLineString:
		ls, err := p.parseLineString(ctype)
		return ls.AsGeometry(), err
	case TypePolygon:
		poly, err := p.parsePolygon(ctype)
		return poly.AsGeometry(), err
	case TypeMultiPoint:
		mp, err := p.parseMultiPoint(ctype)
		return mp.AsGeometry(), err
	case TypeMultiLineString:
		mls, err := p.parseMultiLineString(ctype)
		return mls.AsGeometry(), err
	case TypeMultiPolygon:
		mp, err := p.parseMultiPolygon(ctype)
		return mp.AsGeometry(), err
	case TypeGeometryCollection:
		gc, err := p.parseGeometryCollection(ctype)
		return gc.AsGeometry(), err
	default:
		return Geometry{}, wkbSyntaxError{fmt.Sprintf("unknown geometry type: %d", gtype)}
	}
}

func (p *wkbParser) parseFloat64() (float64, error) {
	if len(p.body) < 8 {
		return 0, wkbSyntaxError{"unexpected EOF"}
	}

	var u uint64
	if p.bo == 0 {
		u = binary.BigEndian.Uint64(p.body)
	} else {
		u = binary.LittleEndian.Uint64(p.body)
	}

	p.body = p.body[8:]
	return math.Float64frombits(u), nil
}

func (p *wkbParser) parsePoint(ctype CoordinatesType) (Point, error) {
	c := Coordinates{Type: ctype}
	var err error
	c.X, err = p.parseFloat64()
	if err != nil {
		return Point{}, err
	}
	c.Y, err = p.parseFloat64()
	if err != nil {
		return Point{}, err
	}

	if ctype == DimXYZ || ctype == DimXYZM {
		c.Z, err = p.parseFloat64()
		if err != nil {
			return Point{}, err
		}
	}
	if ctype == DimXYM || ctype == DimXYZM {
		c.M, err = p.parseFloat64()
		if err != nil {
			return Point{}, err
		}
	}

	if math.IsNaN(c.X) && math.IsNaN(c.Y) {
		// Empty points are represented as NaN,NaN in WKB.
		return Point{}.ForceCoordinatesType(ctype), nil
	}
	if math.IsNaN(c.X) || math.IsNaN(c.Y) {
		return Point{}, wkbSyntaxError{"point contains mixed NaN values"}
	}
	return NewPoint(c), nil
}

func (p *wkbParser) parseLineString(ctype CoordinatesType) (LineString, error) {
	n, err := p.parseUint32()
	if err != nil {
		return LineString{}, err
	}
	floats := make([]float64, int(n)*ctype.Dimension())

	if len(p.body) < 8*len(floats) {
		return LineString{}, wkbSyntaxError{"unexpected EOF"}
	}

	var seqData []byte
	if p.no {
		seqData = p.body[:8*len(floats)]
	} else {
		seqData = make([]byte, 8*len(floats))
		copy(seqData, p.body)
		flipEndianessStride8(seqData)
	}
	p.body = p.body[8*len(floats):]
	copy(floats, bytesAsFloats(seqData))

	seq := NewSequence(floats, ctype)
	return NewLineString(seq), nil
}

// bytesAsFloats reinterprets the bytes slice as a float64 slice in a similar
// manner to reinterpret_cast in C++.
func bytesAsFloats(byts []byte) []float64 {
	if len(byts) == 0 {
		return nil
	}
	return unsafe.Slice((*float64)(unsafe.Pointer(&byts[0])), len(byts)/8)
}

// flipEndianessStride8 flips the endianess of the input bytes, assuming that
// they represent data items that are 8 bytes long.
func flipEndianessStride8(p []byte) {
	for i := 0; i < len(p); i += 8 {
		p[i+0], p[i+7] = p[i+7], p[i+0]
		p[i+1], p[i+6] = p[i+6], p[i+1]
		p[i+2], p[i+5] = p[i+5], p[i+2]
		p[i+3], p[i+4] = p[i+4], p[i+3]
	}
}

func (p *wkbParser) parsePolygon(ctype CoordinatesType) (Polygon, error) {
	n, err := p.parseUint32()
	if err != nil {
		return Polygon{}, err
	}
	if n == 0 {
		return Polygon{}.ForceCoordinatesType(ctype), nil
	}
	rings := make([]LineString, n)
	for i := range rings {
		rings[i], err = p.parseLineString(ctype)
		if err != nil {
			return Polygon{}, err
		}
	}
	return NewPolygon(rings), nil
}

func (p *wkbParser) parseMultiPoint(ctype CoordinatesType) (MultiPoint, error) {
	n, err := p.parseUint32()
	if err != nil {
		return MultiPoint{}, err
	}
	if n == 0 {
		return MultiPoint{}.ForceCoordinatesType(ctype), nil
	}
	pts := make([]Point, n)
	for i := uint32(0); i < n; i++ {
		geom, err := p.inner()
		if err != nil {
			return MultiPoint{}, err
		}
		if !geom.IsPoint() {
			return MultiPoint{}, wkbSyntaxError{"MultiPoint contains non-Point element"}
		}
		pts[i] = geom.MustAsPoint()
	}
	return NewMultiPoint(pts), nil
}

func (p *wkbParser) parseMultiLineString(ctype CoordinatesType) (MultiLineString, error) {
	n, err := p.parseUint32()
	if err != nil {
		return MultiLineString{}, err
	}
	if n == 0 {
		return MultiLineString{}.ForceCoordinatesType(ctype), nil
	}
	lss := make([]LineString, n)
	for i := uint32(0); i < n; i++ {
		geom, err := p.inner()
		if err != nil {
			return MultiLineString{}, err
		}
		if !geom.IsLineString() {
			return MultiLineString{}, wkbSyntaxError{"MultiLineString contains non-LineString element"}
		}
		lss[i] = geom.MustAsLineString()
	}
	return NewMultiLineString(lss), nil
}

func (p *wkbParser) parseMultiPolygon(ctype CoordinatesType) (MultiPolygon, error) {
	n, err := p.parseUint32()
	if err != nil {
		return MultiPolygon{}, err
	}
	if n == 0 {
		return MultiPolygon{}.ForceCoordinatesType(ctype), nil
	}
	polys := make([]Polygon, n)
	for i := uint32(0); i < n; i++ {
		geom, err := p.inner()
		if err != nil {
			return MultiPolygon{}, err
		}
		if !geom.IsPolygon() {
			return MultiPolygon{}, wkbSyntaxError{"MultiPolygon contains non-Polygon element"}
		}
		polys[i] = geom.MustAsPolygon()
	}
	return NewMultiPolygon(polys), nil
}

func (p *wkbParser) parseGeometryCollection(ctype CoordinatesType) (GeometryCollection, error) {
	n, err := p.parseUint32()
	if err != nil {
		return GeometryCollection{}, err
	}
	if n == 0 {
		return GeometryCollection{}.ForceCoordinatesType(ctype), nil
	}
	geoms := make([]Geometry, n)
	for i := uint32(0); i < n; i++ {
		geoms[i], err = p.inner()
		if err != nil {
			return GeometryCollection{}, err
		}
		if geoms[i].CoordinatesType() != ctype {
			err := mismatchedGeometryCollectionDimsError{
				ctype,
				geoms[i].CoordinatesType(),
			}
			return GeometryCollection{}, err
		}
	}
	return NewGeometryCollection(geoms), nil
}
