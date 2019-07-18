package simplefeatures

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
)

// UnmarshalWKB reads the Well Known Binary (WKB), and returns the
// corresponding Geometry.
//
// TODO: should check consistent coordinate types inside compound geometries.
func UnmarshalWKB(r io.Reader) (Geometry, error) {
	p := wkbParser{r: r}
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
	geomType  int
	coordType coordType
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
	p.geomType = int(geomCode % 1000)
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

func (p *wkbParser) parseGeomRoot() Geometry {
	switch p.geomType {
	case 1:
		coords := p.parsePoint()
		if coords.Empty {
			return NewEmptyPoint()
		} else {
			return NewPointFromCoords(coords.Value)
		}
	case 2:
		coords := p.parseLineString()
		switch len(coords) {
		case 0:
			return NewEmptyLineString()
		case 2:
			ln, err := NewLine(coords[0], coords[1])
			p.setErr(err)
			return ln
		default:
			ls, err := NewLineString(coords)
			p.setErr(err)
			return ls
		}
	case 3:
		coords := p.parsePolygon()
		if len(coords) == 0 {
			return NewEmptyPolygon()
		} else {
			poly, err := NewPolygonFromCoords(coords)
			p.setErr(err)
			return poly
		}
	case 4:
		return p.parseMultiPoint()
	case 5:
		return p.parseMultiLineString()
	case 6:
		return p.parseMultiPolygon()
	case 7:
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
	xs, err := NewScalar(strconv.FormatFloat(x, 'f', -1, 64))
	p.setErr(err)
	ys, err := NewScalar(strconv.FormatFloat(y, 'f', -1, 64))
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
	return NewMultiPoint(pts)
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
			ls, err := NewLineString([]Coordinates{c1, c2})
			p.setErr(err)
			lss = append(lss, ls)
		default:
			p.setErr(errors.New("non-LineString found in MultiLineString"))
		}
	}
	return NewMultiLineString(lss)
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
	mpoly, err := NewMultiPolygon(polys)
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
	return NewGeometryCollection(geoms)
}
