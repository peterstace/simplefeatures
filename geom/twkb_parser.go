package geom

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// Tiny Well Known Binary
// See spec https://github.com/TWKB/Specification/blob/master/twkb.md

// UnmarshalTWKB parses a Tiny Well Known Binary (TWKB), returning the
// corresponding Geometry.
func UnmarshalTWKB(twkb []byte, opts ...ConstructorOption) (Geometry, error) {
	p := newtwkbParser(twkb, opts...)
	g, err := p.nextGeometry()
	return g, p.annotateError(err)
}

// UnmarshalTWKBWithHeaders parses a Tiny Well Known Binary (TWKB),
// returning the corresponding Geometry, and any bounding box and any IDs
// listed in its header information.
//
// If there is a bounding box header, the bbox slice will be populated with
// two points, a minimum then a maximum. Otherwise, the slice is empty.
//
// If there is an ID list header, the ids slice will be populated with the
// IDs from that header. Otherwise, the slice is empty.
func UnmarshalTWKBWithHeaders(twkb []byte, opts ...ConstructorOption) (g Geometry, bbox []Point, ids []int64, err error) {
	p := newtwkbParser(twkb, opts...)
	g, err = p.nextGeometry()
	if err != nil {
		return Geometry{}, nil, nil, p.annotateError(err)
	}
	if p.hasBBox {
		bbox, err = UnmarshalTWKBBoundingBoxHeader(twkb)
		if err != nil {
			return Geometry{}, nil, nil, p.annotateError(err)
		}
	}
	return g, bbox, p.idList, p.annotateError(err)
}

// UnmarshalTWKBBoundingBoxHeader checks if the bounding box header
// exists in the Tiny Well Known Binary (TWKB) and if it exists
// returns its minimum and maximum points in the bbox slice
// (otherwise the slice is empty).
//
// Because the results are returned as Points, the X, Y, Z, and M values
// can all be returned. Check the point type to see if the Z and M values
// are valid.
//
// The function returns immediately after parsing the headers.
// Any remaining geometry is not parsed by this function.
func UnmarshalTWKBBoundingBoxHeader(twkb []byte) (bbox []Point, err error) {
	p := newtwkbParser(twkb)
	bbox, err = p.parseBBoxHeader(twkb)
	return bbox, p.annotateError(err)
}

// UnmarshalTWKBEnvelope checks if the bounding box header exists
// in the Tiny Well Known Binary (TWKB) and returns an Envelope
// that is non-empty iff the header exists (thus the envelope
// will not be computed from the data, only from a header).
//
// Note that due to the definition of Envelope, only the X and Y
// coordinates will be returned this way, whereas any Z and M
// coordinates will be silently ignored by this function.
//
// The function returns immediately after parsing the headers.
// Any remaining geometry is not parsed by this function.
func UnmarshalTWKBEnvelope(twkb []byte) (Envelope, error) {
	p := newtwkbParser(twkb)
	if err := p.parseHeaders(); err != nil {
		return Envelope{}, p.annotateError(err)
	}
	if !p.hasBBox {
		return Envelope{}, nil
	}
	return NewEnvelope([]XY{
		{
			p.scalings[0] * float64(p.bbox[0]),
			p.scalings[1] * float64(p.bbox[2]),
		},
		{
			p.scalings[0] * float64(p.bbox[0]+p.bbox[1]),
			p.scalings[1] * float64(p.bbox[2]+p.bbox[3]),
		},
	})
}

// twkbParser holds all state information for interpreting TWKB buffers
// including information such as the last reference point used in coord deltas.
type twkbParser struct {
	twkb []byte
	pos  int
	opts []ConstructorOption

	kind  twkbGeometryType
	ctype CoordinatesType

	dimensions int
	precXY     int
	hasZ       bool
	hasM       bool
	precZ      int
	precM      int
	scalings   [twkbMaxDimensions]float64

	hasBBox bool
	hasSize bool
	hasIDs  bool
	hasExt  bool
	isEmpty bool

	bbox   []int64
	idList []int64

	refpoint [twkbMaxDimensions]int64
}

func newtwkbParser(twkb []byte, opts ...ConstructorOption) twkbParser {
	return twkbParser{
		twkb:       twkb,
		opts:       opts,
		ctype:      DimXY,
		dimensions: 2,
	}
}

// Annotate an error message with the byte offset where it happened.
func (p *twkbParser) annotateError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("TWKB parsing error at byte %d: %w", p.pos, err)
}

// Parse a geometry and return it, the number of bytes consumed, and any error.
func (p *twkbParser) parseGeometry() (Geometry, int, error) {
	g, err := p.nextGeometry()
	return g, p.pos, err
}

// Parse a geometry and return it and any error.
func (p *twkbParser) nextGeometry() (Geometry, error) {
	if err := p.parseHeaders(); err != nil {
		return Geometry{}, err
	}

	switch p.kind {
	case twkbTypePoint:
		pt, err := p.parsePoint()
		if err != nil {
			return Geometry{}, err
		}
		return pt.AsGeometry(), nil
	case twkbTypeLineString:
		ls, err := p.parseLineString()
		if err != nil {
			return Geometry{}, err
		}
		return ls.AsGeometry(), nil
	case twkbTypePolygon:
		poly, err := p.parsePolygon()
		if err != nil {
			return Geometry{}, err
		}
		return poly.AsGeometry(), nil
	case twkbTypeMultiPoint:
		mp, err := p.parseMultiPoint()
		if err != nil {
			return Geometry{}, err
		}
		return mp.AsGeometry(), nil
	case twkbTypeMultiLineString:
		ml, err := p.parseMultiLineString()
		if err != nil {
			return Geometry{}, err
		}
		return ml.AsGeometry(), nil
	case twkbTypeMultiPolygon:
		mp, err := p.parseMultiPolygon()
		if err != nil {
			return Geometry{}, err
		}
		return mp.AsGeometry(), nil
	case twkbTypeGeometryCollection:
		gc, err := p.parseGeometryCollection()
		if err != nil {
			return Geometry{}, err
		}
		return gc.AsGeometry(), nil
	}

	return Geometry{}, fmt.Errorf("unsupported geometry type %d", p.kind)
}

// Parse a geometry's headers.
func (p *twkbParser) parseHeaders() error {
	if err := p.parseTypeAndPrecision(); err != nil {
		return err
	}
	if err := p.parseMetadataHeader(); err != nil {
		return err
	}
	if p.hasExt {
		if err := p.parseExtendedPrecision(); err != nil {
			return err
		}
	}
	if p.hasSize {
		if err := p.parseSize(); err != nil {
			return err
		}
	}
	if p.hasBBox {
		if err := p.parseBBox(); err != nil {
			return err
		}
	}
	return nil
}

func (p *twkbParser) parseTypeAndPrecision() error {
	if len(p.twkb) <= p.pos {
		return errors.New("empty buffer")
	}
	typeprec := p.twkb[p.pos]
	p.pos++

	p.kind = twkbGeometryType(typeprec & 0x0f)
	p.precXY = int(decodeZigZagInt64(uint64(typeprec) >> 4))

	p.scalings[0] = math.Pow10(-p.precXY) // X
	p.scalings[1] = math.Pow10(-p.precXY) // Y
	return nil
}

func (p *twkbParser) parseMetadataHeader() error {
	if len(p.twkb) <= p.pos {
		return errors.New("lacking metadata header")
	}
	metaheader := twkbMetadataHeader(p.twkb[p.pos])
	p.pos++

	p.hasBBox = (metaheader & twkbHasBBox) != 0
	p.hasSize = (metaheader & twkbHasSize) != 0
	p.hasIDs = (metaheader & twkbHasIDs) != 0
	p.hasExt = (metaheader & twkbHasExtPrec) != 0
	p.isEmpty = (metaheader & twkbIsEmpty) != 0
	return nil
}

func (p *twkbParser) parseExtendedPrecision() error {
	if len(p.twkb) <= p.pos {
		return errors.New("lacking extended pecision byte")
	}
	extprec := p.twkb[p.pos]
	p.pos++

	p.ctype = DimXY
	if extprec&1 != 0 {
		p.hasZ = true
		p.precZ = int(extprec >> 2 & 0x07)
	}
	if extprec&2 != 0 {
		p.hasM = true
		p.precM = int(extprec >> 5 & 0x07)
	}
	if p.hasZ && p.hasM {
		p.ctype = DimXYZM
		p.dimensions = 4
		p.scalings[2] = math.Pow10(-p.precZ)
		p.scalings[3] = math.Pow10(-p.precM)
	} else if p.hasZ {
		p.ctype = DimXYZ
		p.dimensions = 3
		p.scalings[2] = math.Pow10(-p.precZ)
	} else if p.hasM {
		p.ctype = DimXYM
		p.dimensions = 3
		p.scalings[2] = math.Pow10(-p.precM)
	}
	return nil
}

func (p *twkbParser) parseSize() error {
	if len(p.twkb) <= p.pos {
		return errors.New("lacking size varint")
	}
	bytesRemaining, err := p.parseUnsignedVarint()
	if err != nil {
		return fmt.Errorf("size varint malformed: %w", err)
	}
	if uint64(p.pos)+bytesRemaining > uint64(len(p.twkb)) {
		return fmt.Errorf("remaining input (%d bytes) smaller than size varint indicates (%d bytes)", len(p.twkb)-p.pos, bytesRemaining)
	}
	return nil
}

func (p *twkbParser) parseBBox() error {
	// Parse the bounding box, but do not otherwise use it.
	for d := 0; d < p.dimensions; d++ {
		minVal, err := p.parseSignedVarint()
		if err != nil {
			return fmt.Errorf("BBox min varint malformed: %w", err)
		}
		p.bbox = append(p.bbox, minVal)

		deltaVal, err := p.parseSignedVarint()
		if err != nil {
			return fmt.Errorf("BBox delta varint malformed: %w", err)
		}
		p.bbox = append(p.bbox, deltaVal)
	}
	return nil
}

func (p *twkbParser) parseBBoxHeader(twkb []byte) (bbox []Point, err error) {
	if err = p.parseHeaders(); err != nil {
		return nil, err
	}
	if !p.hasBBox {
		return nil, nil
	}
	if p.hasZ && p.hasM {
		minX := p.scalings[0] * float64(p.bbox[0])
		minY := p.scalings[1] * float64(p.bbox[2])
		minZ := p.scalings[2] * float64(p.bbox[4])
		minM := p.scalings[3] * float64(p.bbox[6])

		maxX := p.scalings[0] * float64(p.bbox[0]+p.bbox[1])
		maxY := p.scalings[1] * float64(p.bbox[2]+p.bbox[3])
		maxZ := p.scalings[2] * float64(p.bbox[4]+p.bbox[5])
		maxM := p.scalings[3] * float64(p.bbox[6]+p.bbox[7])

		minPt := newUncheckedPoint(Coordinates{XY: XY{minX, minY}, Z: minZ, M: minM, Type: p.ctype})
		maxPt := newUncheckedPoint(Coordinates{XY: XY{maxX, maxY}, Z: maxZ, M: maxM, Type: p.ctype})
		bbox = []Point{minPt, maxPt}

	} else if p.hasZ {
		minX := p.scalings[0] * float64(p.bbox[0])
		minY := p.scalings[1] * float64(p.bbox[2])
		minZ := p.scalings[2] * float64(p.bbox[4])

		maxX := p.scalings[0] * float64(p.bbox[0]+p.bbox[1])
		maxY := p.scalings[1] * float64(p.bbox[2]+p.bbox[3])
		maxZ := p.scalings[2] * float64(p.bbox[4]+p.bbox[5])

		minPt := newUncheckedPoint(Coordinates{XY: XY{minX, minY}, Z: minZ, Type: p.ctype})
		maxPt := newUncheckedPoint(Coordinates{XY: XY{maxX, maxY}, Z: maxZ, Type: p.ctype})
		bbox = []Point{minPt, maxPt}

	} else if p.hasM {
		minX := p.scalings[0] * float64(p.bbox[0])
		minY := p.scalings[1] * float64(p.bbox[2])
		minM := p.scalings[2] * float64(p.bbox[4])

		maxX := p.scalings[0] * float64(p.bbox[0]+p.bbox[1])
		maxY := p.scalings[1] * float64(p.bbox[2]+p.bbox[3])
		maxM := p.scalings[2] * float64(p.bbox[4]+p.bbox[5])

		minPt := newUncheckedPoint(Coordinates{XY: XY{minX, minY}, M: minM, Type: p.ctype})
		maxPt := newUncheckedPoint(Coordinates{XY: XY{maxX, maxY}, M: maxM, Type: p.ctype})
		bbox = []Point{minPt, maxPt}

	} else {
		minX := p.scalings[0] * float64(p.bbox[0])
		minY := p.scalings[1] * float64(p.bbox[2])

		maxX := p.scalings[0] * float64(p.bbox[0]+p.bbox[1])
		maxY := p.scalings[1] * float64(p.bbox[2]+p.bbox[3])

		minPt := newUncheckedPoint(Coordinates{XY: XY{minX, minY}, Type: p.ctype})
		maxPt := newUncheckedPoint(Coordinates{XY: XY{maxX, maxY}, Type: p.ctype})
		bbox = []Point{minPt, maxPt}
	}
	return bbox, nil
}

func (p *twkbParser) parsePoint() (Point, error) {
	if p.isEmpty {
		return NewEmptyPoint(p.ctype), nil
	}
	return p.nextPoint()
}

func (p *twkbParser) nextPoint() (Point, error) {
	var c Coordinates
	c.Type = p.ctype

	coords, err := p.parsePointArray(1)
	if err != nil {
		return Point{}, err
	}

	c.XY.X = coords[0]
	c.XY.Y = coords[1]
	if p.hasZ && p.hasM {
		c.Z = coords[2]
		c.M = coords[3]
	} else if p.hasZ {
		c.Z = coords[2]
	} else if p.hasM {
		c.M = coords[2]
	}

	return NewPoint(c, p.opts...)
}

func (p *twkbParser) parseLineString() (LineString, error) {
	if p.isEmpty {
		return NewLineString(NewSequence(nil, p.ctype), p.opts...)
	}
	return p.nextLineString()
}

func (p *twkbParser) nextLineString() (LineString, error) {
	coords, _, err := p.parsePointCountAndArray()
	if err != nil {
		return LineString{}, err
	}
	return NewLineString(NewSequence(coords, p.ctype), p.opts...)
}

func (p *twkbParser) parsePolygon() (Polygon, error) {
	if p.isEmpty {
		return NewPolygon(nil, p.opts...)
	}
	return p.nextPolygon()
}

func (p *twkbParser) nextPolygon() (Polygon, error) {
	numRings, err := p.parseUnsignedVarint()
	if err != nil {
		return Polygon{}, fmt.Errorf("num rings varint malformed: %w", err)
	}

	var rings []LineString
	for r := uint64(0); r < numRings; r++ {
		coords, numPoints, err := p.parsePointCountAndArray()
		if err != nil {
			return Polygon{}, err
		}
		if numPoints >= 2 {
			// The spec says rings may need to be closed:
			// "rings are assumed to be implicitly closed, so the first
			//  and last point should not be the same"

			// Note, some implementations can generate TWKB with the rings
			// already closed. We wish to gracefully parse these cases too.
			// So check if any of the first and final point's coords differ.
			finalPointDiffersFromFirst := false
			for d := 0; d < p.dimensions; d++ {
				first := coords[d]
				final := coords[d+(numPoints-1)*p.dimensions]
				if first != final {
					finalPointDiffersFromFirst = true
					break
				}
			}

			if finalPointDiffersFromFirst {
				// Append first point again, to close the ring.
				for d := 0; d < p.dimensions; d++ {
					coords = append(coords, coords[d])
				}
			}
		}
		ls, err := NewLineString(NewSequence(coords, p.ctype), p.opts...)
		if err != nil {
			return Polygon{}, err
		}
		rings = append(rings, ls)
	}
	return NewPolygon(rings, p.opts...)
}

func (p *twkbParser) parseMultiPoint() (MultiPoint, error) {
	if p.isEmpty {
		return NewMultiPoint(nil, p.opts...), nil
	}
	return p.nextMultiPoint()
}

func (p *twkbParser) nextMultiPoint() (MultiPoint, error) {
	numPoints, err := p.parseUnsignedVarint()
	if err != nil {
		return MultiPoint{}, fmt.Errorf("num points varint malformed: %w", err)
	}
	if err := p.parseIDList(int(numPoints)); err != nil {
		return MultiPoint{}, err
	}
	var pts []Point
	for i := 0; i < int(numPoints); i++ {
		pt, err := p.nextPoint()
		if err != nil {
			return MultiPoint{}, err
		}
		pts = append(pts, pt)
	}
	return NewMultiPoint(pts, p.opts...), nil
}

func (p *twkbParser) parseMultiLineString() (MultiLineString, error) {
	if p.isEmpty {
		return NewMultiLineString(nil, p.opts...), nil
	}
	return p.nextMultiLineString()
}

func (p *twkbParser) nextMultiLineString() (MultiLineString, error) {
	numLineStrings, err := p.parseUnsignedVarint()
	if err != nil {
		return MultiLineString{}, fmt.Errorf("num linestrings varint malformed: %w", err)
	}
	if err := p.parseIDList(int(numLineStrings)); err != nil {
		return MultiLineString{}, err
	}
	var lines []LineString
	for i := 0; i < int(numLineStrings); i++ {
		ls, err := p.nextLineString()
		if err != nil {
			return MultiLineString{}, err
		}
		lines = append(lines, ls)
	}
	return NewMultiLineString(lines, p.opts...), nil
}

func (p *twkbParser) parseMultiPolygon() (MultiPolygon, error) {
	if p.isEmpty {
		return NewMultiPolygon(nil, p.opts...)
	}
	return p.nextMultiPolygon()
}

func (p *twkbParser) nextMultiPolygon() (MultiPolygon, error) {
	numPolygons, err := p.parseUnsignedVarint()
	if err != nil {
		return MultiPolygon{}, fmt.Errorf("num polygons varint malformed: %w", err)
	}
	if err := p.parseIDList(int(numPolygons)); err != nil {
		return MultiPolygon{}, err
	}
	var polys []Polygon
	for i := 0; i < int(numPolygons); i++ {
		poly, err := p.nextPolygon()
		if err != nil {
			return MultiPolygon{}, err
		}
		polys = append(polys, poly)
	}
	return NewMultiPolygon(polys, p.opts...)
}

func (p *twkbParser) parseGeometryCollection() (GeometryCollection, error) {
	if p.isEmpty {
		return NewGeometryCollection(nil, p.opts...), nil
	}
	return p.nextGeometryCollection()
}

func (p *twkbParser) nextGeometryCollection() (GeometryCollection, error) {
	numGeoms, err := p.parseUnsignedVarint()
	if err != nil {
		return GeometryCollection{}, fmt.Errorf("num polygons varint malformed: %w", err)
	}
	if err := p.parseIDList(int(numGeoms)); err != nil {
		return GeometryCollection{}, err
	}
	var geoms []Geometry
	for i := 0; i < int(numGeoms); i++ {
		subParser := newtwkbParser(p.twkb[p.pos:], p.opts...)
		g, nbytes, err := subParser.parseGeometry()
		if err != nil {
			p.pos += nbytes // Add sub-parser's last known position, for error reporting.
			return GeometryCollection{}, err
		}
		p.pos += nbytes // Sub-parser's geometry has been read, so ensure it is skipped.
		geoms = append(geoms, g)
	}
	return NewGeometryCollection(geoms, p.opts...), nil
}

// Read a number of points then convert that many points from int to float coords.
// Utilise and update the running memory of the previous reference point.
// Return the slice of coords, the number of points, and any error.
func (p *twkbParser) parsePointCountAndArray() ([]float64, int, error) {
	numPoints, err := p.parseUnsignedVarint()
	if err != nil {
		return nil, 0, fmt.Errorf("num points varint malformed: %w", err)
	}

	coords, err := p.parsePointArray(int(numPoints))
	return coords, int(numPoints), err
}

// Convert a given number of points from integer to floating point coordinates.
// Utilise and update the running memory of the previous reference point.
// The returned array will contain numPoints * the number of dimensions values.
func (p *twkbParser) parsePointArray(numPoints int) ([]float64, error) {
	var coords = make([]float64, numPoints*int(p.dimensions))
	c := 0
	for i := 0; i < numPoints; i++ {
		for d := 0; d < p.dimensions; d++ {
			val, err := p.parseSignedVarint()
			if err != nil {
				return nil, fmt.Errorf("point %d of %d coord %d varint malformed: %w", i, numPoints, d, err)
			}

			p.refpoint[d] += val // Reverse coord differencing to find the true value.
			coords[c] = float64(p.refpoint[d]) * p.scalings[d]
			c++
		}
	}
	return coords, nil
}

func (p *twkbParser) parseIDList(numIDs int) error {
	if !p.hasIDs {
		return nil
	}
	p.idList = make([]int64, numIDs)
	for i := 0; i < numIDs; i++ {
		id, err := p.parseSignedVarint()
		if err != nil {
			return fmt.Errorf("ID list varint %d of %d malformed: %w", i, numIDs, err)
		}
		p.idList[i] = id
	}
	return nil
}

func (p *twkbParser) parseUnsignedVarint() (uint64, error) {
	// LEB128.
	val, n := binary.Uvarint(p.twkb[p.pos:])
	if n == 0 {
		return 0, errors.New("problem parsing unsigned varint: buffer too small")
	}
	if n < 0 {
		return 0, errors.New("problem parsing unsigned varint: numeric overflow")
	}
	p.pos += n // Have now read the varint.
	return val, nil
}

func (p *twkbParser) parseSignedVarint() (int64, error) {
	// LEB128 Zig-Zag encoded.
	val, n := binary.Varint(p.twkb[p.pos:])
	if n == 0 {
		return 0, errors.New("problem parsing signed varint: buffer too small")
	}
	if n < 0 {
		return 0, errors.New("problem parsing signed varint: numeric overflow")
	}
	p.pos += n // Have now read the varint.
	return val, nil
}
