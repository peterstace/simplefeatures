package geom

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"unsafe"
)

// UnmarshalWKB reads the Well Known Binary (WKB), and returns the
// corresponding Geometry.
func UnmarshalWKB(wkb []byte, opts ...ConstructorOption) (Geometry, error) {
	p := wkbParser{body: wkb, opts: opts}
	g := p.run()
	return g, p.err
}

type wkbParser struct {
	err  error
	body []byte
	bo   binary.ByteOrder
	opts []ConstructorOption
}

func (p *wkbParser) run() Geometry {
	p.parseByteOrder()
	gtype, ctype := p.parseGeomAndCoordType()
	geom := p.parseGeomRoot(gtype, ctype)
	return geom
}

func (p *wkbParser) inner() Geometry {
	inner := wkbParser{err: p.err, body: p.body, opts: p.opts}
	g := inner.run()
	p.body = inner.body
	p.setErr(inner.err)
	return g
}

func (p *wkbParser) setErr(err error) {
	if p.err == nil {
		p.err = err
	}
}

func (p *wkbParser) parseByteOrder() {
	switch b := p.readByte(); b {
	case 0:
		p.bo = binary.BigEndian
	case 1:
		p.bo = binary.LittleEndian
	default:
		p.setErr(fmt.Errorf("invalid byte order: %x", b))
	}
}

func (p *wkbParser) readByte() byte {
	if len(p.body) == 0 {
		p.setErr(io.ErrUnexpectedEOF)
		return 0
	}
	b := p.body[0]
	p.body = p.body[1:]
	return b
}

func (p *wkbParser) parseUint32() uint32 {
	if len(p.body) < 4 {
		p.setErr(io.ErrUnexpectedEOF)
		return 0
	}
	x := p.bo.Uint32(p.body)
	p.body = p.body[4:]
	return x
}

func (p *wkbParser) parseGeomAndCoordType() (GeometryType, CoordinatesType) {
	geomCode := p.parseUint32()
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
		p.setErr(fmt.Errorf("invalid geometry type in geom code: %v", geomCode))
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
		p.setErr(fmt.Errorf("invalid coordinates type in geom code: %v", geomCode))
	}

	return gtype, ctype
}

func (p *wkbParser) parseGeomRoot(gtype GeometryType, ctype CoordinatesType) Geometry {
	switch gtype {
	case TypePoint:
		return p.parsePoint(ctype).AsGeometry()
	case TypeLineString:
		return p.parseLineString(ctype).AsGeometry()
	case TypePolygon:
		return p.parsePolygon(ctype).AsGeometry()
	case TypeMultiPoint:
		return p.parseMultiPoint(ctype).AsGeometry()
	case TypeMultiLineString:
		return p.parseMultiLineString(ctype).AsGeometry()
	case TypeMultiPolygon:
		return p.parseMultiPolygon(ctype).AsGeometry()
	case TypeGeometryCollection:
		return p.parseGeometryCollection(ctype).AsGeometry()
	default:
		p.setErr(fmt.Errorf("unknown geometry type: %d", gtype))
		return Geometry{}
	}
}

func (p *wkbParser) parseFloat64() float64 {
	if len(p.body) < 8 {
		p.setErr(io.ErrUnexpectedEOF)
		return 0
	}
	u := p.bo.Uint64(p.body)
	p.body = p.body[8:]
	return math.Float64frombits(u)
}

func (p *wkbParser) parsePoint(ctype CoordinatesType) Point {
	var c Coordinates
	c.Type = ctype
	c.X = p.parseFloat64()
	c.Y = p.parseFloat64()
	switch c.Type {
	case DimXY:
	case DimXYZ:
		c.Z = p.parseFloat64()
	case DimXYM:
		c.M = p.parseFloat64()
	case DimXYZM:
		c.Z = p.parseFloat64()
		c.M = p.parseFloat64()
	default:
		p.setErr(errors.New("unknown coord type"))
		return Point{}
	}

	if math.IsNaN(c.X) && math.IsNaN(c.Y) {
		// Empty points are represented as NaN,NaN is WKB.
		return Point{}
	}
	if math.IsNaN(c.X) || math.IsNaN(c.Y) {
		p.setErr(errors.New("point contains mixed NaN values"))
		return Point{}
	}
	return NewPoint(c, p.opts...)
}

func (p *wkbParser) parseLineString(ctype CoordinatesType) LineString {
	n := p.parseUint32()
	floats := make([]float64, int(n)*ctype.Dimension())

	if len(p.body) < 8*len(floats) {
		p.setErr(io.ErrUnexpectedEOF)
		return LineString{}
	}

	var seqData []byte
	if p.bo == nativeOrder {
		seqData = p.body[:8*len(floats)]
	} else {
		seqData = make([]byte, 8*len(floats))
		copy(seqData, p.body)
		flipEndianessStride8(seqData)
	}
	p.body = p.body[8*len(floats):]
	copy(floats, bytesAsFloats(seqData))

	seq := NewSequence(floats, ctype)
	poly, err := NewLineString(seq, p.opts...)
	p.setErr(err)
	return poly
}

// bytesAsFloats reinterprets the bytes slice as a float64 slice in a similar
// manner to reinterpret_cast in C++.
func bytesAsFloats(byts []byte) []float64 {
	bytsHeader := (*reflect.SliceHeader)(unsafe.Pointer(&byts))
	floatsHeader := reflect.SliceHeader{
		Data: bytsHeader.Data,
		Len:  len(byts) / 8,
		Cap:  cap(byts) / 8,
	}
	return *(*[]float64)(unsafe.Pointer(&floatsHeader))
}

// flipEndianessStride8 flips the endianess of the input bytes, assuming that
// they represent data items that are 8 bytes long.
func flipEndianessStride8(p []byte) {
	for i := 0; i < len(p)/8; i += 8 {
		p[i+0], p[i+7] = p[i+7], p[i+0]
		p[i+1], p[i+6] = p[i+6], p[i+1]
		p[i+2], p[i+5] = p[i+5], p[i+2]
		p[i+3], p[i+4] = p[i+4], p[i+3]
	}
}

func (p *wkbParser) parsePolygon(ctype CoordinatesType) Polygon {
	n := p.parseUint32()
	if n == 0 {
		return Polygon{}.ForceCoordinatesType(ctype)
	}
	rings := make([]LineString, n)
	for i := range rings {
		rings[i] = p.parseLineString(ctype)
	}
	poly, err := NewPolygonFromRings(rings, p.opts...)
	p.setErr(err)
	return poly
}

func (p *wkbParser) parseMultiPoint(ctype CoordinatesType) MultiPoint {
	n := p.parseUint32()
	if n == 0 {
		return MultiPoint{}.ForceCoordinatesType(ctype)
	}
	var pts []Point
	for i := uint32(0); i < n; i++ {
		geom := p.inner()
		if !geom.IsPoint() {
			p.setErr(errors.New("non-Point found in MultiPoint"))
		}
		pts = append(pts, geom.AsPoint())
	}
	return NewMultiPointFromPoints(pts, p.opts...)
}

func (p *wkbParser) parseMultiLineString(ctype CoordinatesType) MultiLineString {
	n := p.parseUint32()
	if n == 0 {
		return MultiLineString{}.ForceCoordinatesType(ctype)
	}
	var lss []LineString
	for i := uint32(0); i < n; i++ {
		geom := p.inner()
		if !geom.IsLineString() {
			p.setErr(errors.New("non-LineString found in MultiLineString"))
		} else {
			lss = append(lss, geom.AsLineString())
		}
	}
	return NewMultiLineStringFromLineStrings(lss, p.opts...)
}

func (p *wkbParser) parseMultiPolygon(ctype CoordinatesType) MultiPolygon {
	n := p.parseUint32()
	if n == 0 {
		return MultiPolygon{}.ForceCoordinatesType(ctype)
	}
	var polys []Polygon
	for i := uint32(0); i < n; i++ {
		geom := p.inner()
		if !geom.IsPolygon() {
			p.setErr(errors.New("non-Polygon found in MultiPolygon"))
		}
		polys = append(polys, geom.AsPolygon())
	}
	mpoly, err := NewMultiPolygonFromPolygons(polys, p.opts...)
	p.setErr(err)
	return mpoly
}

func (p *wkbParser) parseGeometryCollection(ctype CoordinatesType) GeometryCollection {
	n := p.parseUint32()
	if n == 0 {
		return GeometryCollection{}.ForceCoordinatesType(ctype)
	}
	var geoms []Geometry
	for i := uint32(0); i < n; i++ {
		geom := p.inner()
		geoms = append(geoms, geom)
	}
	return NewGeometryCollection(geoms, p.opts...)
}
