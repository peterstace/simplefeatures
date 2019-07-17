package simplefeatures

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
)

// UnmarshalWKB reads the Well Known Binary (WKB), and returns the
// corresponding Geometry.
func UnmarshalWKB(r io.Reader) (Geometry, error) {
	p := wkbParser{r: r}
	p.parseByteOrder()
	p.parseGeomType()
	geom := p.parseGeomRoot()
	// TODO: check for trailing bytes
	return geom, p.err
}

type coordType byte

const (
	coordTypeXY   coordType = 1
	coordTypeXYZ  coordType = 2
	coordTypeXYM  coordType = 3
	coordTypeXYZM coordType = 4
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
	if p.err != nil {
		return
	}
	var buf [1]byte
	if _, err := io.ReadFull(p.r, buf[:]); err != nil {
		p.err = err
		return
	}
	switch buf[0] {
	case 0:
		p.bo = binary.BigEndian
	case 1:
		p.bo = binary.LittleEndian
	default:
		p.err = fmt.Errorf("invalid byte order: %x", buf[0])
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
	if p.err != nil {
		return nil
	}
	switch p.geomType {
	case 1:
		return NewPointFromCoords(p.parsePoint())
	case 2:
		log.Println("case 2")
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
	default:
		p.setErr(fmt.Errorf("unknown geometry type: %d", p.geomType))
		return nil
	}
}

func (p *wkbParser) read(ptr interface{}) {
	if p.err != nil {
		return
	}
	p.err = binary.Read(p.r, p.bo, ptr)
}

func (p *wkbParser) parsePoint() Coordinates {
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
		p.err = errors.New("unknown coord type")
		return Coordinates{}
	}

	// Only XY is supported so far.
	_, _ = z, m
	xs, err := NewScalar(strconv.FormatFloat(x, 'f', -1, 64))
	if err != nil {
		p.err = err
		return Coordinates{}
	}
	ys, err := NewScalar(strconv.FormatFloat(y, 'f', -1, 64))
	if err != nil {
		p.err = err
		return Coordinates{}
	}
	return Coordinates{XY{xs, ys}}
}

func (p *wkbParser) parseLineString() []Coordinates {
	n := p.parseUint32()
	coords := make([]Coordinates, n)
	for i := range coords {
		coords[i] = p.parsePoint()
	}
	return coords
}

func (p *wkbParser) parseFloat64() float64 {
	if p.err != nil {
		return 0
	}
	var f float64
	p.read(&f)
	return f
}
