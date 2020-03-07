package geom

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// TODO: When 3D/measure are supported, we will need to check for consistent
// coordinate types inside compound geometries.

// UnmarshalWKB reads the Well Known Binary (WKB), and returns the
// corresponding Geometry.
func UnmarshalWKB(r io.Reader, opts ...ConstructorOption) (Geometry, error) {
	p := wkbParser{r: r, opts: opts}
	p.parseByteOrder()
	p.parseGeomType()
	geom := p.parseGeomRoot()
	return geom, p.err
}

type wkbParser struct {
	err       error
	r         io.Reader
	bo        binary.ByteOrder
	geomType  uint32
	coordType CoordinatesType // TODO: rename to ctype for consistency
	opts      []ConstructorOption
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

func (p *wkbParser) parseGeomType() {
	geomCode := p.parseUint32()
	p.geomType = geomCode % 1000
	switch geomCode / 1000 {
	case 0:
		p.coordType = XYOnly
	case 1:
		p.coordType = XYZ
	case 2:
		p.coordType = XYM
	case 3:
		p.coordType = XYZM
	default:
		p.setErr(errors.New("cannot determine coordinate type"))
	}
}

const (
	wkbGeomTypePoint              = uint32(1)
	wkbGeomTypeLineString         = uint32(2)
	wkbGeomTypePolygon            = uint32(3)
	wkbGeomTypeMultiPoint         = uint32(4)
	wkbGeomTypeMultiLineString    = uint32(5)
	wkbGeomTypeMultiPolygon       = uint32(6)
	wkbGeomTypeGeometryCollection = uint32(7)
)

func (p *wkbParser) parseGeomRoot() Geometry {
	switch p.geomType {
	case wkbGeomTypePoint:
		c, ok := p.parsePoint()
		if !ok {
			return NewEmptyPoint(p.coordType).AsGeometry()
		} else {
			return NewPointC(c, p.coordType, p.opts...).AsGeometry()
		}
	case wkbGeomTypeLineString:
		ls := p.parseLineString()
		seq := ls.Coordinates()
		if seq.Length() == 2 {
			ln, err := NewLineC(seq.Get(0), seq.Get(1), p.coordType, p.opts...)
			p.setErr(err)
			return ln.AsGeometry()
		}
		return ls.AsGeometry()
	case wkbGeomTypePolygon:
		return p.parsePolygon().AsGeometry()
	case wkbGeomTypeMultiPoint:
		return p.parseMultiPoint().AsGeometry()
	case wkbGeomTypeMultiLineString:
		return p.parseMultiLineString().AsGeometry()
	case wkbGeomTypeMultiPolygon:
		return p.parseMultiPolygon().AsGeometry()
	case wkbGeomTypeGeometryCollection:
		return p.parseGeometryCollection().AsGeometry()
	default:
		p.setErr(fmt.Errorf("unknown geometry type: %d", p.geomType))
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

func (p *wkbParser) parsePoint() (Coordinates, bool) {
	var c Coordinates
	c.X = p.parseFloat64()
	c.Y = p.parseFloat64()
	switch p.coordType {
	case XYOnly:
	case XYZ:
		c.Z = p.parseFloat64()
	case XYM:
		c.M = p.parseFloat64()
	case XYZM:
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

func (p *wkbParser) parseLineString() LineString {
	n := p.parseUint32()
	floats := make([]float64, 0, int(n)*p.coordType.Dimension())
	for i := uint32(0); i < n; i++ {
		c, ok := p.parsePoint()
		if !ok {
			p.setErr(errors.New("empty point not allowed in LineString"))
		}
		floats = append(floats, c.X, c.Y)
		if p.coordType.Is3D() {
			floats = append(floats, c.Z)
		}
		if p.coordType.IsMeasured() {
			floats = append(floats, c.M)
		}
	}
	seq := NewSequenceNoCopy(floats, p.coordType)
	poly, err := NewLineStringFromSequence(seq, p.opts...)
	p.setErr(err)
	return poly
}

func (p *wkbParser) parsePolygon() Polygon {
	n := p.parseUint32()
	rings := make([]LineString, n)
	for i := range rings {
		rings[i] = p.parseLineString()
	}
	poly, err := NewPolygon(rings, p.coordType, p.opts...)
	p.setErr(err)
	return poly
}

func (p *wkbParser) parseMultiPoint() MultiPoint {
	n := p.parseUint32()
	var pts []Point
	for i := uint32(0); i < n; i++ {
		geom, err := UnmarshalWKB(p.r)
		p.setErr(err)
		if !geom.IsPoint() {
			p.setErr(errors.New("non-Point found in MultiPoint"))
		}
		pts = append(pts, geom.AsPoint())
	}
	mp, err := NewMultiPoint(pts, p.coordType, p.opts...)
	p.setErr(err)
	return mp
}

func (p *wkbParser) parseMultiLineString() MultiLineString {
	n := p.parseUint32()
	var lss []LineString
	for i := uint32(0); i < n; i++ {
		geom, err := UnmarshalWKB(p.r)
		p.setErr(err)
		switch {
		case geom.IsLineString():
			lss = append(lss, geom.AsLineString())
		case geom.IsLine():
			ln := geom.AsLine()
			ls := ln.AsLineString()
			lss = append(lss, ls)
		default:
			p.setErr(errors.New("non-LineString found in MultiLineString"))
		}
	}
	mls, err := NewMultiLineString(lss, p.coordType, p.opts...)
	p.setErr(err)
	return mls
}

func (p *wkbParser) parseMultiPolygon() MultiPolygon {
	n := p.parseUint32()
	var polys []Polygon
	for i := uint32(0); i < n; i++ {
		geom, err := UnmarshalWKB(p.r)
		p.setErr(err)
		if !geom.IsPolygon() {
			p.setErr(errors.New("non-Polygon found in MultiPolygon"))
		}
		polys = append(polys, geom.AsPolygon())
	}
	mpoly, err := NewMultiPolygon(polys, p.coordType, p.opts...)
	p.setErr(err)
	return mpoly
}

func (p *wkbParser) parseGeometryCollection() GeometryCollection {
	n := p.parseUint32()
	var geoms []Geometry
	for i := uint32(0); i < n; i++ {
		geom, err := UnmarshalWKB(p.r)
		p.setErr(err)
		geoms = append(geoms, geom)
	}
	gc, err := NewGeometryCollection(geoms, p.coordType, p.opts...)
	p.setErr(err)
	return gc
}
