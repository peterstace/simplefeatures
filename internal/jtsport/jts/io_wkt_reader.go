package jts

import (
	"io"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const (
	wktComma     = ","
	wktLParen    = "("
	wktRParen    = ")"
	wktNaNSymbol = "NaN"
)

// Io_WKTReader converts a geometry in Well-Known Text format to a Geometry.
//
// WKTReader supports extracting Geometry objects from either io.Readers or
// Strings. This allows it to function as a parser to read Geometry objects
// from text blocks embedded in other data formats (e.g. XML).
//
// A WKTReader is parameterized by a GeometryFactory, to allow it to create
// Geometry objects of the appropriate implementation. In particular, the
// GeometryFactory determines the PrecisionModel and SRID that is used.
//
// The WKTReader converts all input numbers to the precise internal representation.
//
// As of version 1.15, JTS can read (but not write) WKT syntax which specifies
// coordinate dimension Z, M or ZM as modifiers (e.g. POINT Z) or in the name
// of the geometry type (e.g. LINESTRINGZM). If the coordinate dimension is
// specified it will be set in the created geometry. If the coordinate dimension
// is not specified, the default behaviour is to create XYZ geometry (this is
// backwards compatible with older JTS versions). This can be altered to create
// XY geometry by calling SetIsOldJtsCoordinateSyntaxAllowed(false).
//
// A reader can be set to ensure the input is structurally valid by calling
// SetFixStructure(true). This ensures that geometry can be constructed without
// errors due to missing coordinates. The created geometry may still be
// topologically invalid.
//
// Notes:
//   - Keywords are case-insensitive.
//   - The reader supports non-standard "LINEARRING" tags.
//   - The reader uses strconv.ParseFloat to perform the conversion of ASCII
//     numbers to floating point. This means it supports Go syntax for floating
//     point literals (including scientific notation).
type Io_WKTReader struct {
	geometryFactory *Geom_GeometryFactory
	csFactory       Geom_CoordinateSequenceFactory
	precisionModel  *Geom_PrecisionModel

	isAllowOldJtsCoordinateSyntax bool
	isAllowOldJtsMultipointSyntax bool
	isFixStructure                bool
}

// Io_NewWKTReader creates a reader that creates objects using the default GeometryFactory.
func Io_NewWKTReader() *Io_WKTReader {
	return Io_NewWKTReaderWithFactory(Geom_NewGeometryFactoryDefault())
}

// Io_NewWKTReaderWithFactory creates a reader that creates objects using the given GeometryFactory.
func Io_NewWKTReaderWithFactory(geometryFactory *Geom_GeometryFactory) *Io_WKTReader {
	return &Io_WKTReader{
		geometryFactory:               geometryFactory,
		csFactory:                     geometryFactory.GetCoordinateSequenceFactory(),
		precisionModel:                geometryFactory.GetPrecisionModel(),
		isAllowOldJtsCoordinateSyntax: true,
		isAllowOldJtsMultipointSyntax: true,
		isFixStructure:                false,
	}
}

// SetIsOldJtsCoordinateSyntaxAllowed sets a flag indicating that coordinates may have 3 ordinate
// values even though no Z or M ordinate indicator is present. The default value is true.
func (r *Io_WKTReader) SetIsOldJtsCoordinateSyntaxAllowed(value bool) {
	r.isAllowOldJtsCoordinateSyntax = value
}

// SetIsOldJtsMultiPointSyntaxAllowed sets a flag indicating that point coordinates in a
// MultiPoint geometry must not be enclosed in parentheses. The default value is true.
func (r *Io_WKTReader) SetIsOldJtsMultiPointSyntaxAllowed(value bool) {
	r.isAllowOldJtsMultipointSyntax = value
}

// SetFixStructure sets a flag indicating that the structure of input geometry should be fixed
// so that the geometry can be constructed without error. This involves adding coordinates if
// the input coordinate sequence is shorter than required.
func (r *Io_WKTReader) SetFixStructure(isFixStructure bool) {
	r.isFixStructure = isFixStructure
}

// Read reads a Well-Known Text representation of a Geometry from a string.
func (r *Io_WKTReader) Read(wellKnownText string) (*Geom_Geometry, error) {
	reader := strings.NewReader(wellKnownText)
	return r.ReadFromReader(reader)
}

// ReadFromReader reads a Well-Known Text representation of a Geometry from a Reader.
func (r *Io_WKTReader) ReadFromReader(reader io.Reader) (*Geom_Geometry, error) {
	tokenizer := newWktTokenizer(reader)
	return r.readGeometryTaggedText(tokenizer)
}

// wktTokenizer is a simple tokenizer for WKT parsing.
type wktTokenizer struct {
	reader        io.RuneReader
	peeked        rune
	hasPeeked     bool
	currentVal    string
	pushedBackVal string
	hasPushedBack bool
	lineNo        int
	eof           bool
}

func newWktTokenizer(r io.Reader) *wktTokenizer {
	var runeReader io.RuneReader
	if rr, ok := r.(io.RuneReader); ok {
		runeReader = rr
	} else {
		runeReader = &byteRuneReader{r: r}
	}
	return &wktTokenizer{
		reader: runeReader,
		lineNo: 1,
	}
}

// byteRuneReader wraps an io.Reader to implement io.RuneReader.
type byteRuneReader struct {
	r io.Reader
}

func (b *byteRuneReader) ReadRune() (rune, int, error) {
	var buf [1]byte
	n, err := b.r.Read(buf[:])
	if err != nil {
		return 0, 0, err
	}
	if n == 0 {
		return 0, 0, io.EOF
	}
	return rune(buf[0]), 1, nil
}

func (t *wktTokenizer) readRune() (rune, error) {
	if t.hasPeeked {
		t.hasPeeked = false
		return t.peeked, nil
	}
	r, _, err := t.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			t.eof = true
		}
		return 0, err
	}
	if r == '\n' {
		t.lineNo++
	}
	return r, nil
}

func (t *wktTokenizer) peekRune() (rune, error) {
	if t.hasPeeked {
		return t.peeked, nil
	}
	r, _, err := t.reader.ReadRune()
	if err != nil {
		return 0, err
	}
	t.peeked = r
	t.hasPeeked = true
	return r, nil
}

func (t *wktTokenizer) skipWhitespace() error {
	for {
		r, err := t.peekRune()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if r == '#' {
			// Skip comment until end of line.
			for {
				r, err = t.readRune()
				if err == io.EOF {
					return nil
				}
				if err != nil {
					return err
				}
				if r == '\n' {
					break
				}
			}
			continue
		}
		if !unicode.IsSpace(r) {
			return nil
		}
		_, _ = t.readRune()
	}
}

// nextToken returns the next token, or io.EOF if end of stream.
func (t *wktTokenizer) nextToken() (string, error) {
	// Check for pushed back token first.
	if t.hasPushedBack {
		t.hasPushedBack = false
		t.currentVal = t.pushedBackVal
		return t.currentVal, nil
	}

	if err := t.skipWhitespace(); err != nil {
		return "", err
	}

	r, err := t.readRune()
	if err == io.EOF {
		return "", io.EOF
	}
	if err != nil {
		return "", err
	}

	// Single character tokens.
	if r == '(' || r == ')' || r == ',' {
		t.currentVal = string(r)
		return t.currentVal, nil
	}

	// Word token (including numbers).
	if t.isWordChar(r) {
		var sb strings.Builder
		sb.WriteRune(r)
		for {
			r, err := t.peekRune()
			if err == io.EOF {
				break
			}
			if err != nil {
				return "", err
			}
			if !t.isWordChar(r) {
				break
			}
			_, _ = t.readRune()
			sb.WriteRune(r)
		}
		t.currentVal = sb.String()
		return t.currentVal, nil
	}

	return "", Io_NewParseException("Unexpected character: " + string(r))
}

func (t *wktTokenizer) isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '+' || r == '.'
}

func (t *wktTokenizer) pushBack(token string) {
	t.pushedBackVal = token
	t.hasPushedBack = true
}

func (t *wktTokenizer) lineNo_() int {
	return t.lineNo
}

// isNumberNext tests if the next token in the stream is a number.
func (t *wktTokenizer) isNumberNext() (bool, error) {
	if err := t.skipWhitespace(); err != nil {
		return false, err
	}
	r, err := t.peekRune()
	if err == io.EOF {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return t.isWordChar(r), nil
}

// isOpenerNext tests if the next token in the stream is a left parenthesis.
func (t *wktTokenizer) isOpenerNext() (bool, error) {
	if err := t.skipWhitespace(); err != nil {
		return false, err
	}
	r, err := t.peekRune()
	if err == io.EOF {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return r == '(', nil
}

// Parsing methods for WKTReader.

func (r *Io_WKTReader) getCoordinate(tokenizer *wktTokenizer, ordinateFlags *Io_OrdinateSet, tryParen bool) (*Geom_Coordinate, error) {
	opened := false
	if tryParen {
		isOpener, err := tokenizer.isOpenerNext()
		if err != nil {
			return nil, err
		}
		if isOpener {
			_, err = tokenizer.nextToken()
			if err != nil {
				return nil, err
			}
			opened = true
		}
	}

	offsetM := 0
	if ordinateFlags.Contains(Io_Ordinate_Z) {
		offsetM = 1
	}
	coord := r.createCoordinate(ordinateFlags)

	x, err := r.getNextNumber(tokenizer)
	if err != nil {
		return nil, err
	}
	coord.SetOrdinate(Geom_CoordinateSequence_X, r.precisionModel.MakePrecise(x))

	y, err := r.getNextNumber(tokenizer)
	if err != nil {
		return nil, err
	}
	coord.SetOrdinate(Geom_CoordinateSequence_Y, r.precisionModel.MakePrecise(y))

	if ordinateFlags.Contains(Io_Ordinate_Z) {
		z, err := r.getNextNumber(tokenizer)
		if err != nil {
			return nil, err
		}
		coord.SetOrdinate(Geom_CoordinateSequence_Z, z)
	}
	if ordinateFlags.Contains(Io_Ordinate_M) {
		m, err := r.getNextNumber(tokenizer)
		if err != nil {
			return nil, err
		}
		coord.SetOrdinate(Geom_CoordinateSequence_Z+offsetM, m)
	}

	if ordinateFlags.Size() == 2 && r.isAllowOldJtsCoordinateSyntax {
		isNumber, err := tokenizer.isNumberNext()
		if err != nil {
			return nil, err
		}
		if isNumber {
			z, err := r.getNextNumber(tokenizer)
			if err != nil {
				return nil, err
			}
			coord.SetOrdinate(Geom_CoordinateSequence_Z, z)
		}
	}

	if opened {
		_, err = r.getNextCloser(tokenizer)
		if err != nil {
			return nil, err
		}
	}

	return coord, nil
}

func (r *Io_WKTReader) createCoordinate(ordinateFlags *Io_OrdinateSet) *Geom_Coordinate {
	hasZ := ordinateFlags.Contains(Io_Ordinate_Z)
	hasM := ordinateFlags.Contains(Io_Ordinate_M)
	if hasZ && hasM {
		return Geom_NewCoordinateXYZM4DWithXYZM(0, 0, 0, 0).Geom_Coordinate
	}
	if hasM {
		return Geom_NewCoordinateXYM3DWithXYM(0, 0, 0).Geom_Coordinate
	}
	if hasZ || r.isAllowOldJtsCoordinateSyntax {
		return Geom_NewCoordinate()
	}
	return Geom_NewCoordinateXY2DWithXY(0, 0).Geom_Coordinate
}

func (r *Io_WKTReader) getCoordinateSequence(tokenizer *wktTokenizer, ordinateFlags *Io_OrdinateSet, minSize int, isRing bool) (Geom_CoordinateSequence, error) {
	nextWord, err := r.getNextEmptyOrOpener(tokenizer)
	if err != nil {
		return nil, err
	}
	if nextWord == Io_WKTConstants_EMPTY {
		return r.createCoordinateSequenceEmpty(ordinateFlags), nil
	}

	var coordinates []*Geom_Coordinate
	for {
		coord, err := r.getCoordinate(tokenizer, ordinateFlags, false)
		if err != nil {
			return nil, err
		}
		coordinates = append(coordinates, coord)

		nextToken, err := r.getNextCloserOrComma(tokenizer)
		if err != nil {
			return nil, err
		}
		if nextToken != wktComma {
			break
		}
	}

	if r.isFixStructure {
		coordinates = r.fixStructure(coordinates, minSize, isRing)
	}

	return r.csFactory.CreateFromCoordinates(coordinates), nil
}

func (r *Io_WKTReader) fixStructure(coords []*Geom_Coordinate, minSize int, isRing bool) []*Geom_Coordinate {
	if len(coords) == 0 {
		return coords
	}
	if isRing && !r.isClosed(coords) {
		coords = append(coords, coords[0].Copy())
	}
	for len(coords) < minSize {
		coords = append(coords, coords[len(coords)-1].Copy())
	}
	return coords
}

func (r *Io_WKTReader) isClosed(coords []*Geom_Coordinate) bool {
	if len(coords) == 0 {
		return true
	}
	if len(coords) == 1 || !coords[0].Equals2D(coords[len(coords)-1]) {
		return false
	}
	return true
}

func (r *Io_WKTReader) createCoordinateSequenceEmpty(ordinateFlags *Io_OrdinateSet) Geom_CoordinateSequence {
	measures := 0
	if ordinateFlags.Contains(Io_Ordinate_M) {
		measures = 1
	}
	return r.csFactory.CreateWithSizeAndDimensionAndMeasures(0, r.toDimension(ordinateFlags), measures)
}

func (r *Io_WKTReader) getCoordinateSequenceOldMultiPoint(tokenizer *wktTokenizer, ordinateFlags *Io_OrdinateSet) (Geom_CoordinateSequence, error) {
	var coordinates []*Geom_Coordinate
	for {
		coord, err := r.getCoordinate(tokenizer, ordinateFlags, true)
		if err != nil {
			return nil, err
		}
		coordinates = append(coordinates, coord)

		nextToken, err := r.getNextCloserOrComma(tokenizer)
		if err != nil {
			return nil, err
		}
		if nextToken != wktComma {
			break
		}
	}

	return r.csFactory.CreateFromCoordinates(coordinates), nil
}

func (r *Io_WKTReader) toDimension(ordinateFlags *Io_OrdinateSet) int {
	dimension := 2
	if ordinateFlags.Contains(Io_Ordinate_Z) {
		dimension++
	}
	if ordinateFlags.Contains(Io_Ordinate_M) {
		dimension++
	}
	if dimension == 2 && r.isAllowOldJtsCoordinateSyntax {
		dimension++
	}
	return dimension
}

func (r *Io_WKTReader) getNextNumber(tokenizer *wktTokenizer) (float64, error) {
	token, err := tokenizer.nextToken()
	if err == io.EOF {
		return 0, r.parseErrorExpected(tokenizer, "number")
	}
	if err != nil {
		return 0, err
	}

	if strings.EqualFold(token, wktNaNSymbol) {
		return math.NaN(), nil
	}

	val, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return 0, r.parseErrorWithLine(tokenizer, "Invalid number: "+token)
	}
	return val, nil
}

func (r *Io_WKTReader) getNextEmptyOrOpener(tokenizer *wktTokenizer) (string, error) {
	nextWord, err := r.getNextWord(tokenizer)
	if err != nil {
		return "", err
	}

	upperWord := strings.ToUpper(nextWord)
	if upperWord == Io_WKTConstants_Z {
		nextWord, err = r.getNextWord(tokenizer)
		if err != nil {
			return "", err
		}
	} else if upperWord == Io_WKTConstants_M {
		nextWord, err = r.getNextWord(tokenizer)
		if err != nil {
			return "", err
		}
	} else if upperWord == Io_WKTConstants_ZM {
		nextWord, err = r.getNextWord(tokenizer)
		if err != nil {
			return "", err
		}
	}

	if nextWord == Io_WKTConstants_EMPTY || nextWord == wktLParen {
		return nextWord, nil
	}
	return "", r.parseErrorExpected(tokenizer, Io_WKTConstants_EMPTY+" or "+wktLParen)
}

func (r *Io_WKTReader) getNextOrdinateFlags(tokenizer *wktTokenizer) (*Io_OrdinateSet, error) {
	result := Io_Ordinate_CreateXY()

	nextWord, err := r.lookAheadWord(tokenizer)
	if err != nil {
		return result, nil
	}

	upperWord := strings.ToUpper(nextWord)
	if upperWord == Io_WKTConstants_Z {
		_, _ = tokenizer.nextToken()
		result.Add(Io_Ordinate_Z)
	} else if upperWord == Io_WKTConstants_M {
		_, _ = tokenizer.nextToken()
		result.Add(Io_Ordinate_M)
	} else if upperWord == Io_WKTConstants_ZM {
		_, _ = tokenizer.nextToken()
		result.Add(Io_Ordinate_Z)
		result.Add(Io_Ordinate_M)
	}
	return result, nil
}

func (r *Io_WKTReader) lookAheadWord(tokenizer *wktTokenizer) (string, error) {
	nextWord, err := r.getNextWord(tokenizer)
	if err != nil {
		return "", err
	}
	tokenizer.pushBack(nextWord)
	return nextWord, nil
}

func (r *Io_WKTReader) getNextCloserOrComma(tokenizer *wktTokenizer) (string, error) {
	nextWord, err := r.getNextWord(tokenizer)
	if err != nil {
		return "", err
	}
	if nextWord == wktComma || nextWord == wktRParen {
		return nextWord, nil
	}
	return "", r.parseErrorExpected(tokenizer, wktComma+" or "+wktRParen)
}

func (r *Io_WKTReader) getNextCloser(tokenizer *wktTokenizer) (string, error) {
	nextWord, err := r.getNextWord(tokenizer)
	if err != nil {
		return "", err
	}
	if nextWord == wktRParen {
		return nextWord, nil
	}
	return "", r.parseErrorExpected(tokenizer, wktRParen)
}

func (r *Io_WKTReader) getNextWord(tokenizer *wktTokenizer) (string, error) {
	token, err := tokenizer.nextToken()
	if err == io.EOF {
		return "", r.parseErrorExpected(tokenizer, "word")
	}
	if err != nil {
		return "", err
	}

	if strings.EqualFold(token, Io_WKTConstants_EMPTY) {
		return Io_WKTConstants_EMPTY, nil
	}
	if token == "(" {
		return wktLParen, nil
	}
	if token == ")" {
		return wktRParen, nil
	}
	if token == "," {
		return wktComma, nil
	}
	return token, nil
}

func (r *Io_WKTReader) parseErrorExpected(tokenizer *wktTokenizer, expected string) error {
	return r.parseErrorWithLine(tokenizer, "Expected "+expected+" but found '"+tokenizer.currentVal+"'")
}

func (r *Io_WKTReader) parseErrorWithLine(tokenizer *wktTokenizer, msg string) error {
	return Io_NewParseException(msg + " (line " + strconv.Itoa(tokenizer.lineNo_()) + ")")
}

func (r *Io_WKTReader) readGeometryTaggedText(tokenizer *wktTokenizer) (*Geom_Geometry, error) {
	ordinateFlags := Io_Ordinate_CreateXY()

	typeWord, err := r.getNextWord(tokenizer)
	if err != nil {
		return nil, err
	}
	typeWord = strings.ToUpper(typeWord)

	if strings.HasSuffix(typeWord, Io_WKTConstants_ZM) {
		ordinateFlags.Add(Io_Ordinate_Z)
		ordinateFlags.Add(Io_Ordinate_M)
	} else if strings.HasSuffix(typeWord, Io_WKTConstants_Z) {
		ordinateFlags.Add(Io_Ordinate_Z)
	} else if strings.HasSuffix(typeWord, Io_WKTConstants_M) {
		ordinateFlags.Add(Io_Ordinate_M)
	}
	return r.readGeometryTaggedTextWithType(tokenizer, typeWord, ordinateFlags)
}

func (r *Io_WKTReader) readGeometryTaggedTextWithType(tokenizer *wktTokenizer, typeWord string, ordinateFlags *Io_OrdinateSet) (*Geom_Geometry, error) {
	if ordinateFlags.Size() == 2 {
		flags, err := r.getNextOrdinateFlags(tokenizer)
		if err != nil {
			return nil, err
		}
		ordinateFlags = flags
	}

	// Check if we can create a sequence with the required dimension.
	// If not, use the XYZM factory.
	measures := 0
	if ordinateFlags.Contains(Io_Ordinate_M) {
		measures = 1
	}
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				r.geometryFactory = Geom_NewGeometryFactory(
					r.geometryFactory.GetPrecisionModel(),
					r.geometryFactory.GetSRID(),
					GeomImpl_CoordinateArraySequenceFactory_Instance(),
				)
				r.csFactory = r.geometryFactory.GetCoordinateSequenceFactory()
			}
		}()
		r.csFactory.CreateWithSizeAndDimensionAndMeasures(0, r.toDimension(ordinateFlags), measures)
	}()

	isType, err := r.isTypeName(tokenizer, typeWord, Io_WKTConstants_POINT)
	if err != nil {
		return nil, err
	}
	if isType {
		return r.readPointText(tokenizer, ordinateFlags)
	}

	isType, err = r.isTypeName(tokenizer, typeWord, Io_WKTConstants_LINESTRING)
	if err != nil {
		return nil, err
	}
	if isType {
		return r.readLineStringText(tokenizer, ordinateFlags)
	}

	isType, err = r.isTypeName(tokenizer, typeWord, Io_WKTConstants_LINEARRING)
	if err != nil {
		return nil, err
	}
	if isType {
		return r.readLinearRingText(tokenizer, ordinateFlags)
	}

	isType, err = r.isTypeName(tokenizer, typeWord, Io_WKTConstants_POLYGON)
	if err != nil {
		return nil, err
	}
	if isType {
		return r.readPolygonText(tokenizer, ordinateFlags)
	}

	isType, err = r.isTypeName(tokenizer, typeWord, Io_WKTConstants_MULTIPOINT)
	if err != nil {
		return nil, err
	}
	if isType {
		return r.readMultiPointText(tokenizer, ordinateFlags)
	}

	isType, err = r.isTypeName(tokenizer, typeWord, Io_WKTConstants_MULTILINESTRING)
	if err != nil {
		return nil, err
	}
	if isType {
		return r.readMultiLineStringText(tokenizer, ordinateFlags)
	}

	isType, err = r.isTypeName(tokenizer, typeWord, Io_WKTConstants_MULTIPOLYGON)
	if err != nil {
		return nil, err
	}
	if isType {
		return r.readMultiPolygonText(tokenizer, ordinateFlags)
	}

	isType, err = r.isTypeName(tokenizer, typeWord, Io_WKTConstants_GEOMETRYCOLLECTION)
	if err != nil {
		return nil, err
	}
	if isType {
		return r.readGeometryCollectionText(tokenizer, ordinateFlags)
	}

	return nil, r.parseErrorWithLine(tokenizer, "Unknown geometry type: "+typeWord)
}

func (r *Io_WKTReader) isTypeName(tokenizer *wktTokenizer, typeWord string, typeName string) (bool, error) {
	if !strings.HasPrefix(typeWord, typeName) {
		return false, nil
	}

	modifiers := typeWord[len(typeName):]
	isValidMod := len(modifiers) <= 2 &&
		(len(modifiers) == 0 ||
			modifiers == Io_WKTConstants_Z ||
			modifiers == Io_WKTConstants_M ||
			modifiers == Io_WKTConstants_ZM)
	if !isValidMod {
		return false, r.parseErrorWithLine(tokenizer, "Invalid dimension modifiers: "+typeWord)
	}

	return true, nil
}

func (r *Io_WKTReader) readPointText(tokenizer *wktTokenizer, ordinateFlags *Io_OrdinateSet) (*Geom_Geometry, error) {
	seq, err := r.getCoordinateSequence(tokenizer, ordinateFlags, 1, false)
	if err != nil {
		return nil, err
	}
	point := r.geometryFactory.CreatePointFromCoordinateSequence(seq)
	return point.Geom_Geometry, nil
}

func (r *Io_WKTReader) readLineStringText(tokenizer *wktTokenizer, ordinateFlags *Io_OrdinateSet) (*Geom_Geometry, error) {
	seq, err := r.getCoordinateSequence(tokenizer, ordinateFlags, Geom_LineString_MINIMUM_VALID_SIZE, false)
	if err != nil {
		return nil, err
	}
	ls := r.geometryFactory.CreateLineStringFromCoordinateSequence(seq)
	return ls.Geom_Geometry, nil
}

func (r *Io_WKTReader) readLinearRingText(tokenizer *wktTokenizer, ordinateFlags *Io_OrdinateSet) (*Geom_Geometry, error) {
	seq, err := r.getCoordinateSequence(tokenizer, ordinateFlags, Geom_LinearRing_MinimumValidSize, true)
	if err != nil {
		return nil, err
	}
	ring := r.geometryFactory.CreateLinearRingFromCoordinateSequence(seq)
	return ring.Geom_Geometry, nil
}

func (r *Io_WKTReader) readPolygonText(tokenizer *wktTokenizer, ordinateFlags *Io_OrdinateSet) (*Geom_Geometry, error) {
	nextToken, err := r.getNextEmptyOrOpener(tokenizer)
	if err != nil {
		return nil, err
	}
	if nextToken == Io_WKTConstants_EMPTY {
		poly := r.geometryFactory.CreatePolygonFromCoordinateSequence(r.createCoordinateSequenceEmpty(ordinateFlags))
		return poly.Geom_Geometry, nil
	}

	var holes []*Geom_LinearRing
	shellGeom, err := r.readLinearRingText(tokenizer, ordinateFlags)
	if err != nil {
		return nil, err
	}
	shell := java.Cast[*Geom_LinearRing](shellGeom)

	nextToken, err = r.getNextCloserOrComma(tokenizer)
	if err != nil {
		return nil, err
	}
	for nextToken == wktComma {
		holeGeom, err := r.readLinearRingText(tokenizer, ordinateFlags)
		if err != nil {
			return nil, err
		}
		hole := java.Cast[*Geom_LinearRing](holeGeom)
		holes = append(holes, hole)
		nextToken, err = r.getNextCloserOrComma(tokenizer)
		if err != nil {
			return nil, err
		}
	}

	poly := r.geometryFactory.CreatePolygonWithLinearRingAndHoles(shell, holes)
	return poly.Geom_Geometry, nil
}

func (r *Io_WKTReader) readMultiPointText(tokenizer *wktTokenizer, ordinateFlags *Io_OrdinateSet) (*Geom_Geometry, error) {
	nextToken, err := r.getNextEmptyOrOpener(tokenizer)
	if err != nil {
		return nil, err
	}
	if nextToken == Io_WKTConstants_EMPTY {
		mp := r.geometryFactory.CreateMultiPointFromPoints([]*Geom_Point{})
		return mp.Geom_Geometry, nil
	}

	// Check for old-style JTS syntax (no parentheses surrounding Point coordinates).
	if r.isAllowOldJtsMultipointSyntax {
		nextWord, err := r.lookAheadWord(tokenizer)
		if err == nil && nextWord != wktLParen && nextWord != Io_WKTConstants_EMPTY {
			seq, err := r.getCoordinateSequenceOldMultiPoint(tokenizer, ordinateFlags)
			if err != nil {
				return nil, err
			}
			mp := r.geometryFactory.CreateMultiPointFromCoordinateSequence(seq)
			return mp.Geom_Geometry, nil
		}
	}

	var points []*Geom_Point
	pointGeom, err := r.readPointText(tokenizer, ordinateFlags)
	if err != nil {
		return nil, err
	}
	points = append(points, java.Cast[*Geom_Point](pointGeom))

	nextToken, err = r.getNextCloserOrComma(tokenizer)
	if err != nil {
		return nil, err
	}
	for nextToken == wktComma {
		pointGeom, err = r.readPointText(tokenizer, ordinateFlags)
		if err != nil {
			return nil, err
		}
		points = append(points, java.Cast[*Geom_Point](pointGeom))
		nextToken, err = r.getNextCloserOrComma(tokenizer)
		if err != nil {
			return nil, err
		}
	}

	mp := r.geometryFactory.CreateMultiPointFromPoints(points)
	return mp.Geom_Geometry, nil
}

func (r *Io_WKTReader) readMultiLineStringText(tokenizer *wktTokenizer, ordinateFlags *Io_OrdinateSet) (*Geom_Geometry, error) {
	nextToken, err := r.getNextEmptyOrOpener(tokenizer)
	if err != nil {
		return nil, err
	}
	if nextToken == Io_WKTConstants_EMPTY {
		mls := r.geometryFactory.CreateMultiLineString()
		return mls.Geom_Geometry, nil
	}

	var lineStrings []*Geom_LineString
	for {
		lsGeom, err := r.readLineStringText(tokenizer, ordinateFlags)
		if err != nil {
			return nil, err
		}
		lineStrings = append(lineStrings, java.Cast[*Geom_LineString](lsGeom))
		nextToken, err = r.getNextCloserOrComma(tokenizer)
		if err != nil {
			return nil, err
		}
		if nextToken != wktComma {
			break
		}
	}

	mls := r.geometryFactory.CreateMultiLineStringFromLineStrings(lineStrings)
	return mls.Geom_Geometry, nil
}

func (r *Io_WKTReader) readMultiPolygonText(tokenizer *wktTokenizer, ordinateFlags *Io_OrdinateSet) (*Geom_Geometry, error) {
	nextToken, err := r.getNextEmptyOrOpener(tokenizer)
	if err != nil {
		return nil, err
	}
	if nextToken == Io_WKTConstants_EMPTY {
		mpoly := r.geometryFactory.CreateMultiPolygon()
		return mpoly.Geom_Geometry, nil
	}

	var polygons []*Geom_Polygon
	for {
		polyGeom, err := r.readPolygonText(tokenizer, ordinateFlags)
		if err != nil {
			return nil, err
		}
		polygons = append(polygons, java.Cast[*Geom_Polygon](polyGeom))
		nextToken, err = r.getNextCloserOrComma(tokenizer)
		if err != nil {
			return nil, err
		}
		if nextToken != wktComma {
			break
		}
	}

	mpoly := r.geometryFactory.CreateMultiPolygonFromPolygons(polygons)
	return mpoly.Geom_Geometry, nil
}

func (r *Io_WKTReader) readGeometryCollectionText(tokenizer *wktTokenizer, ordinateFlags *Io_OrdinateSet) (*Geom_Geometry, error) {
	nextToken, err := r.getNextEmptyOrOpener(tokenizer)
	if err != nil {
		return nil, err
	}
	if nextToken == Io_WKTConstants_EMPTY {
		gc := r.geometryFactory.CreateGeometryCollection()
		return gc.Geom_Geometry, nil
	}

	var geometries []*Geom_Geometry
	for {
		geom, err := r.readGeometryTaggedText(tokenizer)
		if err != nil {
			return nil, err
		}
		geometries = append(geometries, geom)
		nextToken, err = r.getNextCloserOrComma(tokenizer)
		if err != nil {
			return nil, err
		}
		if nextToken != wktComma {
			break
		}
	}

	gc := r.geometryFactory.CreateGeometryCollectionFromGeometries(geometries)
	return gc.Geom_Geometry, nil
}
