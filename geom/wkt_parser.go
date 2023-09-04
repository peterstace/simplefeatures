package geom

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Function names is the parser are chosen to match closely with the BNF
// productions in the WKT grammar.
//
// Convention: functions starting with 'next' consume token(s), and build the
// next production in the grammar.

// UnmarshalWKT parses a Well Known Text (WKT), and returns the corresponding
// Geometry. The result has its geometry constraints validated.
func UnmarshalWKT(wkt string) (Geometry, error) {
	g, err := UnmarshalWKTWithoutValidation(wkt)
	if err != nil {
		return Geometry{}, err
	}
	if err := g.Validate(); err != nil {
		return Geometry{}, err
	}
	return g, nil
}

// UnmarshalWKTWithoutValidation parses a Well Known Text (WKT), and returns
// the corresponding Geometry. The result does not have its geometry
// constraints validated.
func UnmarshalWKTWithoutValidation(wkt string) (Geometry, error) {
	p := newParser(wkt)
	geom, err := p.nextGeometryTaggedText()
	if err != nil {
		return Geometry{}, err
	}

	if tok, err := p.lexer.next(); err == nil {
		return Geometry{}, wantButGot("EOF", tok)
	} else if err != wktUnexpectedEOF {
		return Geometry{}, err
	}
	return geom, nil
}

func wantButGot(wantTok, gotTok string) error {
	return wktSyntaxError{fmt.Sprintf(
		"unexpected token: '%s' (expected %s)",
		gotTok, wantTok,
	)}
}

func newParser(wkt string) *parser {
	return &parser{newWKTLexer(wkt)}
}

type parser struct {
	lexer wktLexer
}

func (p *parser) nextGeometryTaggedText() (Geometry, error) {
	geomType, ctype, err := p.nextGeomTag()
	if err != nil {
		return Geometry{}, err
	}
	switch geomType {
	case "POINT":
		c, ok, err := p.nextPointText(ctype)
		if err != nil {
			return Geometry{}, err
		}
		if !ok {
			return NewEmptyPoint(ctype).AsGeometry(), nil
		}
		return NewPointWithoutValidation(c).AsGeometry(), nil
	case "LINESTRING":
		ls, err := p.nextLineStringText(ctype)
		return ls.AsGeometry(), err
	case "POLYGON":
		p, err := p.nextPolygonText(ctype)
		return p.AsGeometry(), err
	case "MULTIPOINT":
		mp, err := p.nextMultiPointText(ctype)
		return mp.AsGeometry(), err
	case "MULTILINESTRING":
		mls, err := p.nextMultiLineString(ctype)
		return mls.AsGeometry(), err
	case "MULTIPOLYGON":
		mp, err := p.nextMultiPolygonText(ctype)
		return mp.AsGeometry(), err
	case "GEOMETRYCOLLECTION":
		gc, err := p.nextGeometryCollectionText(ctype)
		return gc.AsGeometry(), err
	default:
		return Geometry{}, wantButGot("geometry tag", geomType)
	}
}

func (p *parser) nextGeomTag() (string, CoordinatesType, error) {
	tok, err := p.lexer.next()
	if err != nil {
		return "", 0, err
	}
	geomType := strings.ToUpper(tok)

	tok, err = p.lexer.peek()
	if err != nil {
		return "", 0, err
	}
	dim := DimXY
	switch tok {
	case "Z":
		dim = DimXYZ
	case "M":
		dim = DimXYM
	case "ZM":
		dim = DimXYZM
	}
	if dim != DimXY {
		if _, err := p.lexer.next(); err != nil {
			return "", 0, err
		}
	}

	return geomType, dim, nil
}

func (p *parser) nextEmptySetOrLeftParen() (string, error) {
	tok, err := p.lexer.next()
	if err != nil {
		return "", err
	}
	if tok != "EMPTY" && tok != "(" {
		return "", wantButGot("'EMPTY' or '('", tok)
	}
	return tok, nil
}

func (p *parser) nextRightParen() error {
	tok, err := p.lexer.next()
	if err != nil {
		return err
	}
	if tok != ")" {
		return wantButGot("')'", tok)
	}
	return nil
}

func (p *parser) nextCommaOrRightParen() (string, error) {
	tok, err := p.lexer.next()
	if err != nil {
		return "", err
	}
	if tok != ")" && tok != "," {
		return "", wantButGot("')' or ','", tok)
	}
	return tok, nil
}

func (p *parser) nextPoint(ctype CoordinatesType) (Coordinates, error) {
	var err error
	var c Coordinates
	c.Type = ctype
	c.X, err = p.nextSignedNumericLiteral()
	if err != nil {
		return Coordinates{}, err
	}
	c.Y, err = p.nextSignedNumericLiteral()
	if err != nil {
		return Coordinates{}, err
	}
	if ctype.Is3D() {
		c.Z, err = p.nextSignedNumericLiteral()
		if err != nil {
			return Coordinates{}, err
		}
	}
	if ctype.IsMeasured() {
		c.M, err = p.nextSignedNumericLiteral()
		if err != nil {
			return Coordinates{}, err
		}
	}
	return c, nil
}

func (p *parser) nextPointAppend(dst []float64, ctype CoordinatesType) ([]float64, error) {
	for i := 0; i < ctype.Dimension(); i++ {
		flt, err := p.nextSignedNumericLiteral()
		if err != nil {
			return nil, err
		}
		dst = append(dst, flt)
	}
	return dst, nil
}

func (p *parser) nextSignedNumericLiteral() (float64, error) {
	var negative bool
	tok, err := p.lexer.next()
	if err != nil {
		return 0, err
	}
	if tok == "-" {
		negative = true
		tok, err = p.lexer.next()
		if err != nil {
			return 0, err
		}
	}
	f, err := strconv.ParseFloat(tok, 64)
	if err != nil {
		return 0, wktSyntaxError{err.Error()}
	}
	// NaNs and Infs are not allowed by the WKT grammar.
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, wktSyntaxError{fmt.Sprintf("invalid numeric literal: %v", tok)}
	}
	if negative {
		f *= -1
	}
	return f, nil
}

func (p *parser) nextPointText(ctype CoordinatesType) (Coordinates, bool, error) {
	tok, err := p.nextEmptySetOrLeftParen()
	if err != nil {
		return Coordinates{}, false, err
	}
	if tok == "EMPTY" {
		return Coordinates{}, false, nil
	}
	c, err := p.nextPoint(ctype)
	if err != nil {
		return Coordinates{}, false, err
	}
	if err := p.nextRightParen(); err != nil {
		return Coordinates{}, false, err
	}
	return c, true, nil
}

func (p *parser) nextLineStringText(ctype CoordinatesType) (LineString, error) {
	var floats []float64
	tok, err := p.nextEmptySetOrLeftParen()
	if err != nil {
		return LineString{}, err
	}
	if tok == "(" {
		floats, err = p.nextPointAppend(floats, ctype)
		if err != nil {
			return LineString{}, err
		}
		for {
			tok, err := p.nextCommaOrRightParen()
			if err != nil {
				return LineString{}, err
			}
			if tok == "," {
				floats, err = p.nextPointAppend(floats, ctype)
				if err != nil {
					return LineString{}, err
				}
			} else {
				break
			}
		}
	}
	seq := NewSequence(floats, ctype)
	return NewLineStringWithoutValidation(seq), nil
}

func (p *parser) nextPolygonText(ctype CoordinatesType) (Polygon, error) {
	rings, err := p.nextPolygonOrMultiLineStringText(ctype)
	if err != nil {
		return Polygon{}, err
	}
	if len(rings) == 0 {
		return Polygon{}.ForceCoordinatesType(ctype), nil
	}
	return NewPolygonWithoutValidation(rings), nil
}

func (p *parser) nextMultiLineString(ctype CoordinatesType) (MultiLineString, error) {
	lss, err := p.nextPolygonOrMultiLineStringText(ctype)
	if err != nil {
		return MultiLineString{}, err
	}
	if len(lss) == 0 {
		return MultiLineString{}.ForceCoordinatesType(ctype), nil
	}
	return NewMultiLineString(lss), nil
}

func (p *parser) nextPolygonOrMultiLineStringText(ctype CoordinatesType) ([]LineString, error) {
	tok, err := p.nextEmptySetOrLeftParen()
	if err != nil {
		return nil, err
	}
	if tok == "EMPTY" {
		return nil, nil
	}
	ls, err := p.nextLineStringText(ctype)
	if err != nil {
		return nil, err
	}
	lss := []LineString{ls}
	for {
		tok, err := p.nextCommaOrRightParen()
		if err != nil {
			return nil, err
		}
		if tok == "," {
			ls, err := p.nextLineStringText(ctype)
			if err != nil {
				return nil, err
			}
			lss = append(lss, ls)
		} else {
			break
		}
	}
	return lss, nil
}

func (p *parser) nextMultiPointText(ctype CoordinatesType) (MultiPoint, error) {
	tok, err := p.nextEmptySetOrLeftParen()
	if err != nil {
		return MultiPoint{}, err
	}
	var points []Point
	if tok == "(" {
		for {
			coords, ok, err := p.nextMultiPointStylePoint(ctype)
			if err != nil {
				return MultiPoint{}, err
			}
			if ok {
				pt := NewPointWithoutValidation(coords)
				points = append(points, pt)
			} else {
				points = append(points, NewEmptyPoint(ctype))
			}
			tok, err = p.nextCommaOrRightParen()
			if err != nil {
				return MultiPoint{}, err
			}
			if tok != "," {
				break
			}
		}
	}
	if len(points) == 0 {
		return MultiPoint{}.ForceCoordinatesType(ctype), nil
	}
	return NewMultiPoint(points), nil
}

func (p *parser) nextMultiPointStylePoint(ctype CoordinatesType) (Coordinates, bool, error) {
	// Allowing parentheses to be omitted is an extension of the spec, and is
	// required to handle WKT output from non-complying implementations. In
	// particular, PostGIS doesn't comply to the spec (it outputs points as
	// part of a multipoint without their surrounding parentheses).
	tok, err := p.lexer.peek()
	if err != nil {
		return Coordinates{}, false, err
	}

	var useParens bool
	switch tok {
	case "(":
		if _, err := p.lexer.next(); err != nil {
			return Coordinates{}, false, err
		}
		useParens = true
	case "EMPTY":
		_, err := p.lexer.next()
		return Coordinates{}, false, err
	}

	coords, err := p.nextPoint(ctype)
	if err != nil {
		return Coordinates{}, false, err
	}
	if useParens {
		if err := p.nextRightParen(); err != nil {
			return Coordinates{}, false, err
		}
	}
	return coords, true, nil
}

func (p *parser) nextMultiPolygonText(ctype CoordinatesType) (MultiPolygon, error) {
	var polys []Polygon
	tok, err := p.nextEmptySetOrLeftParen()
	if err != nil {
		return MultiPolygon{}, err
	}
	if tok == "(" {
		for i := 0; true; i++ {
			poly, err := p.nextPolygonText(ctype)
			if err != nil {
				return MultiPolygon{}, err
			}
			polys = append(polys, poly)
			tok, err := p.nextCommaOrRightParen()
			if err != nil {
				return MultiPolygon{}, err
			}
			if tok != "," {
				break
			}
		}
	}
	if len(polys) == 0 {
		return MultiPolygon{}.ForceCoordinatesType(ctype), nil
	}
	return NewMultiPolygonWithoutValidation(polys), nil
}

func (p *parser) nextGeometryCollectionText(ctype CoordinatesType) (GeometryCollection, error) {
	var geoms []Geometry
	tok, err := p.nextEmptySetOrLeftParen()
	if err != nil {
		return GeometryCollection{}, err
	}
	if tok == "(" {
		for {
			g, err := p.nextGeometryTaggedText()
			if err != nil {
				return GeometryCollection{}, err
			}
			geoms = append(geoms, g)
			tok, err := p.nextCommaOrRightParen()
			if err != nil {
				return GeometryCollection{}, err
			}
			if tok != "," {
				break
			}
		}
	}
	if len(geoms) == 0 {
		return GeometryCollection{}.ForceCoordinatesType(ctype), nil
	}
	return NewGeometryCollection(geoms), nil
}
