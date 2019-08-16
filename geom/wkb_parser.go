package geom

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
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

type coordType byte

const (
	coordTypeXY coordType = 1 + iota
	coordTypeXYZ
	coordTypeXYM
	coordTypeXYZM
)

type wkbParser struct {
	err       error
	r         io.Reader
	bo        binary.ByteOrder
	geomType  uint32
	coordType coordType
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
		p.coordType = coordTypeXY
	case 1:
		p.coordType = coordTypeXYZ
	case 2:
		p.coordType = coordTypeXYM
	case 3:
		p.coordType = coordTypeXYZM
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
		coords := p.parsePoint()
		if coords.Empty {
			return NewEmptyPoint(p.opts...)
		} else {
			return NewPointC(coords.Value, p.opts...)
		}
	case wkbGeomTypeLineString:
		coords := p.parseLineString()
		switch len(coords) {
		case 0:
			return NewEmptyLineString(p.opts...)
		case 2:
			ln, err := NewLineC(coords[0], coords[1], p.opts...)
			p.setErr(err)
			return ln
		default:
			ls, err := NewLineStringC(coords, p.opts...)
			p.setErr(err)
			return ls
		}
	case wkbGeomTypePolygon:
		coords := p.parsePolygon()
		if len(coords) == 0 {
			return NewEmptyPolygon(p.opts...)
		} else {
			poly, err := NewPolygonC(coords, p.opts...)
			p.setErr(err)
			return poly
		}
	case wkbGeomTypeMultiPoint:
		return p.parseMultiPoint()
	case wkbGeomTypeMultiLineString:
		return p.parseMultiLineString()
	case wkbGeomTypeMultiPolygon:
		return p.parseMultiPolygon()
	case wkbGeomTypeGeometryCollection:
		return p.parseGeometryCollection()
	default:
		p.setErr(fmt.Errorf("unknown geometry type: %d", p.geomType))
		return nil
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

func (p *wkbParser) parsePoint() OptionalCoordinates {
	x := p.parseFloat64()
	y := p.parseFloat64()
	var z, m float64
	switch p.coordType {
	case coordTypeXY:
	case coordTypeXYZ:
		z = p.parseFloat64()
	case coordTypeXYM:
		m = p.parseFloat64()
	case coordTypeXYZM:
		z = p.parseFloat64()
		m = p.parseFloat64()
	default:
		p.setErr(errors.New("unknown coord type"))
		return OptionalCoordinates{}
	}

	if math.IsNaN(x) && math.IsNaN(y) {
		// Empty points are represented as NaN,NaN is WKB.
		return OptionalCoordinates{Empty: true}
	}
	if math.IsNaN(x) || math.IsNaN(y) {
		p.setErr(errors.New("point contains mixed NaN values"))
		return OptionalCoordinates{}
	}

	// Only XY is supported so far.
	_, _ = z, m
	xs, err := NewScalarS(strconv.FormatFloat(x, 'f', -1, 64))
	p.setErr(err)
	ys, err := NewScalarS(strconv.FormatFloat(y, 'f', -1, 64))
	p.setErr(err)
	return OptionalCoordinates{Value: Coordinates{XY{xs, ys}}}
}

func (p *wkbParser) parseLineString() []Coordinates {
	n := p.parseUint32()
	coords := make([]Coordinates, 0, n)
	for i := uint32(0); i < n; i++ {
		pt := p.parsePoint()
		if !pt.Empty {
			coords = append(coords, pt.Value)
		}
	}
	return coords
}

func (p *wkbParser) parsePolygon() [][]Coordinates {
	n := p.parseUint32()
	coords := make([][]Coordinates, n)
	for i := range coords {
		coords[i] = p.parseLineString()
	}
	return coords
}

func (p *wkbParser) parseMultiPoint() MultiPoint {
	n := p.parseUint32()
	var pts []Point
	for i := uint32(0); i < n; i++ {
		geom, err := UnmarshalWKB(p.r)
		p.setErr(err)
		if geom != nil && geom.IsEmpty() {
			continue
		}
		pt, ok := geom.(Point)
		if !ok {
			p.setErr(errors.New("non-Point found in MultiPoint"))
		}
		pts = append(pts, pt)
	}
	return NewMultiPoint(pts, p.opts...)
}

func (p *wkbParser) parseMultiLineString() MultiLineString {
	n := p.parseUint32()
	var lss []LineString
	for i := uint32(0); i < n; i++ {
		geom, err := UnmarshalWKB(p.r)
		p.setErr(err)
		if geom != nil && geom.IsEmpty() {
			continue
		}
		switch geom := geom.(type) {
		case LineString:
			lss = append(lss, geom)
		case Line:
			c1 := geom.StartPoint().Coordinates()
			c2 := geom.EndPoint().Coordinates()
			ls, err := NewLineStringC([]Coordinates{c1, c2}, p.opts...)
			p.setErr(err)
			lss = append(lss, ls)
		default:
			p.setErr(errors.New("non-LineString found in MultiLineString"))
		}
	}
	return NewMultiLineString(lss, p.opts...)
}

func (p *wkbParser) parseMultiPolygon() MultiPolygon {
	n := p.parseUint32()
	var polys []Polygon
	for i := uint32(0); i < n; i++ {
		geom, err := UnmarshalWKB(p.r)
		p.setErr(err)
		if geom != nil && geom.IsEmpty() {
			continue
		}
		poly, ok := geom.(Polygon)
		if !ok {
			p.setErr(errors.New("non-Polygon found in MultiPolygon"))
		}
		polys = append(polys, poly)
	}
	mpoly, err := NewMultiPolygon(polys, p.opts...)
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
	return NewGeometryCollection(geoms, p.opts...)
}
