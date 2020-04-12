package geom

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// UnmarshalWKB reads the Well Known Binary (WKB), and returns the
// corresponding Geometry.
func UnmarshalWKB(r io.Reader, opts ...ConstructorOption) (Geometry, error) {
	p := wkbParser{r: r, opts: opts}
	p.parseByteOrder()
	gtype, ctype := p.parseGeomAndCoordType()
	geom := p.parseGeomRoot(gtype, ctype)
	return geom, p.err
}

type wkbParser struct {
	err  error
	r    io.Reader
	bo   binary.ByteOrder
	opts []ConstructorOption
}

func (p *wkbParser) setErr(err error) {
	if p.err == nil {
		p.err = err
	}
}

func (p *wkbParser) parseByteOrder() {
	var buf [1]byte
	_, err := io.ReadFull(p.r, buf[:])
	p.setErr(err)
	switch buf[0] {
	case 0:
		p.bo = binary.BigEndian
	case 1:
		p.bo = binary.LittleEndian
	default:
		p.setErr(fmt.Errorf("invalid byte order: %x", buf[0]))
	}
}

func (p *wkbParser) parseUint32() uint32 {
	var x uint32
	p.read(&x)
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
		c, ok := p.parsePoint(ctype)
		if !ok {
			return NewEmptyPoint(ctype).AsGeometry()
		}
		return NewPoint(c, p.opts...).AsGeometry()
	case TypeLineString:
		ls := p.parseLineString(ctype)
		return ls.AsGeometry()
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

func (p *wkbParser) read(ptr interface{}) {
	if p.bo == nil {
		return // an error will have already been set
	}
	p.setErr(binary.Read(p.r, p.bo, ptr))
}

func (p *wkbParser) parseFloat64() float64 {
	var f float64
	p.read(&f)
	return f
}

func (p *wkbParser) parsePoint(ctype CoordinatesType) (Coordinates, bool) {
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
		return Coordinates{}, false
	}

	if math.IsNaN(c.X) && math.IsNaN(c.Y) {
		// Empty points are represented as NaN,NaN is WKB.
		return Coordinates{}, false
	}
	if math.IsNaN(c.X) || math.IsNaN(c.Y) {
		p.setErr(errors.New("point contains mixed NaN values"))
		return Coordinates{}, false
	}
	return c, true
}

func (p *wkbParser) parseLineString(ctype CoordinatesType) LineString {
	n := p.parseUint32()
	floats := make([]float64, 0, int(n)*ctype.Dimension())
	for i := uint32(0); i < n; i++ {
		c, ok := p.parsePoint(ctype)
		if !ok {
			p.setErr(errors.New("empty point not allowed in LineString"))
		}
		floats = append(floats, c.X, c.Y)
		if ctype.Is3D() {
			floats = append(floats, c.Z)
		}
		if ctype.IsMeasured() {
			floats = append(floats, c.M)
		}
	}
	seq := NewSequence(floats, ctype)
	poly, err := NewLineString(seq, p.opts...)
	p.setErr(err)
	return poly
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
		geom, err := UnmarshalWKB(p.r)
		p.setErr(err)
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
		geom, err := UnmarshalWKB(p.r)
		p.setErr(err)
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
		geom, err := UnmarshalWKB(p.r)
		p.setErr(err)
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
		geom, err := UnmarshalWKB(p.r)
		p.setErr(err)
		geoms = append(geoms, geom)
	}
	return NewGeometryCollection(geoms, p.opts...)
}
