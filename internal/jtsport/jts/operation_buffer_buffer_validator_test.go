package jts

import (
	"fmt"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// BufferValidator is a test helper for validating buffer operations.

type operationBuffer_BufferValidator_Test interface {
	GetName() string
	Test(t *testing.T, bv *operationBuffer_BufferValidator) error
	GetPriority() int
}

type operationBuffer_BufferValidator_baseTest struct {
	name     string
	priority int
}

func (bt *operationBuffer_BufferValidator_baseTest) GetName() string {
	return bt.name
}

func (bt *operationBuffer_BufferValidator_baseTest) GetPriority() int {
	return bt.priority
}

type operationBuffer_BufferValidator struct {
	original       *Geom_Geometry
	bufferDistance float64
	nameToTestMap  map[string]operationBuffer_BufferValidator_Test
	buffer         *Geom_Geometry
	wkt            string
	geomFact       *Geom_GeometryFactory
	wktWriter      *Io_WKTWriter
	wktReader      *Io_WKTReader
	t              *testing.T
}

const operationBuffer_BufferValidator_QUADRANT_SEGMENTS_1 = 100
const operationBuffer_BufferValidator_QUADRANT_SEGMENTS_2 = 50

func operationBuffer_NewBufferValidator(bufferDistance float64, wkt string) *operationBuffer_BufferValidator {
	return operationBuffer_NewBufferValidatorWithContainsTest(bufferDistance, wkt, true)
}

func operationBuffer_NewBufferValidatorWithContainsTest(bufferDistance float64, wkt string, addContainsTest bool) *operationBuffer_BufferValidator {
	bv := &operationBuffer_BufferValidator{
		bufferDistance: bufferDistance,
		wkt:            wkt,
		nameToTestMap:  make(map[string]operationBuffer_BufferValidator_Test),
		geomFact:       Geom_NewGeometryFactoryDefault(),
		wktWriter:      Io_NewWKTWriter(),
	}
	// SRID = 888 is to test that SRID is preserved in computed buffers
	bv.SetFactory(Geom_NewPrecisionModel(), 888)
	if addContainsTest {
		bv.addContainsTest()
	}
	//bv.addBufferResultValidatorTest()
	return bv
}

func (bv *operationBuffer_BufferValidator) Test(t *testing.T) {
	bv.t = t
	for _, test := range bv.nameToTestMap {
		err := test.Test(t, bv)
		if err != nil {
			t.Errorf("%s", bv.supplement(err.Error()))
		}
	}
}

func (bv *operationBuffer_BufferValidator) supplement(message string) string {
	newMessage := "\n" + message + "\n"
	original := bv.getOriginal()
	newMessage += "Original: " + bv.wktWriter.WriteFormatted(original) + "\n"
	newMessage += fmt.Sprintf("Buffer Distance: %v\n", bv.bufferDistance)
	buffer := bv.getBuffer()
	newMessage += "Buffer: " + bv.wktWriter.WriteFormatted(buffer) + "\n"
	return newMessage[:len(newMessage)-1]
}

func (bv *operationBuffer_BufferValidator) addTest(test operationBuffer_BufferValidator_Test) *operationBuffer_BufferValidator {
	bv.nameToTestMap[test.GetName()] = test
	return bv
}

func (bv *operationBuffer_BufferValidator) SetExpectedArea(expectedArea float64) *operationBuffer_BufferValidator {
	return bv.addTest(&operationBuffer_BufferValidator_areaTest{
		operationBuffer_BufferValidator_baseTest: operationBuffer_BufferValidator_baseTest{name: "Area Test", priority: 2},
		expectedArea:                             expectedArea,
	})
}

type operationBuffer_BufferValidator_areaTest struct {
	operationBuffer_BufferValidator_baseTest
	expectedArea float64
}

func (at *operationBuffer_BufferValidator_areaTest) Test(t *testing.T, bv *operationBuffer_BufferValidator) error {
	tolerance := bv.getBuffer().GetArea() - bv.getOriginal().BufferWithQuadrantSegments(
		bv.bufferDistance,
		operationBuffer_BufferValidator_QUADRANT_SEGMENTS_1-operationBuffer_BufferValidator_QUADRANT_SEGMENTS_2,
	).GetArea()
	if tolerance < 0 {
		tolerance = -tolerance
	}
	actual := bv.getBuffer().GetArea()
	if actual < at.expectedArea-tolerance || actual > at.expectedArea+tolerance {
		return fmt.Errorf("%s: expected %v, got %v (tolerance %v)", at.GetName(), at.expectedArea, actual, tolerance)
	}
	return nil
}

func (bv *operationBuffer_BufferValidator) SetEmptyBufferExpected(emptyBufferExpected bool) *operationBuffer_BufferValidator {
	return bv.addTest(&operationBuffer_BufferValidator_emptyTest{
		operationBuffer_BufferValidator_baseTest: operationBuffer_BufferValidator_baseTest{name: "Empty Buffer Test", priority: 1},
		emptyBufferExpected:                      emptyBufferExpected,
	})
}

type operationBuffer_BufferValidator_emptyTest struct {
	operationBuffer_BufferValidator_baseTest
	emptyBufferExpected bool
}

func (et *operationBuffer_BufferValidator_emptyTest) Test(t *testing.T, bv *operationBuffer_BufferValidator) error {
	buffer := bv.getBuffer()
	isEmpty := buffer.IsEmpty()
	if isEmpty != et.emptyBufferExpected {
		expectedStr := ""
		if !et.emptyBufferExpected {
			expectedStr = "not "
		}
		return fmt.Errorf("Expected buffer %sto be empty", expectedStr)
	}
	return nil
}

func (bv *operationBuffer_BufferValidator) SetBufferHolesExpected(bufferHolesExpected bool) *operationBuffer_BufferValidator {
	return bv.addTest(&operationBuffer_BufferValidator_holesTest{
		operationBuffer_BufferValidator_baseTest: operationBuffer_BufferValidator_baseTest{name: "Buffer Holes Test", priority: 2},
		bufferHolesExpected:                      bufferHolesExpected,
	})
}

type operationBuffer_BufferValidator_holesTest struct {
	operationBuffer_BufferValidator_baseTest
	bufferHolesExpected bool
}

func (ht *operationBuffer_BufferValidator_holesTest) Test(t *testing.T, bv *operationBuffer_BufferValidator) error {
	buffer := bv.getBuffer()
	hasHoles := ht.hasHoles(buffer)
	if hasHoles != ht.bufferHolesExpected {
		expectedStr := ""
		if !ht.bufferHolesExpected {
			expectedStr = "not "
		}
		return fmt.Errorf("Expected buffer %sto have holes", expectedStr)
	}
	return nil
}

func (ht *operationBuffer_BufferValidator_holesTest) hasHoles(buffer *Geom_Geometry) bool {
	if buffer.IsEmpty() {
		return false
	}
	if java.InstanceOf[*Geom_Polygon](buffer) {
		return java.Cast[*Geom_Polygon](buffer).GetNumInteriorRing() > 0
	}
	multiPolygon := java.Cast[*Geom_MultiPolygon](buffer)
	for i := 0; i < multiPolygon.GetNumGeometries(); i++ {
		if ht.hasHoles(multiPolygon.GetGeometryN(i)) {
			return true
		}
	}
	return false
}

func (bv *operationBuffer_BufferValidator) getOriginal() *Geom_Geometry {
	if bv.original == nil {
		geom, err := bv.wktReader.Read(bv.wkt)
		if err != nil {
			panic(fmt.Sprintf("failed to read WKT: %v", err))
		}
		bv.original = geom
	}
	return bv.original
}

func (bv *operationBuffer_BufferValidator) SetPrecisionModel(precisionModel *Geom_PrecisionModel) *operationBuffer_BufferValidator {
	bv.wktReader = Io_NewWKTReaderWithFactory(Geom_NewGeometryFactoryWithPrecisionModel(precisionModel))
	return bv
}

func (bv *operationBuffer_BufferValidator) SetFactory(precisionModel *Geom_PrecisionModel, srid int) *operationBuffer_BufferValidator {
	bv.wktReader = Io_NewWKTReaderWithFactory(Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, srid))
	return bv
}

func (bv *operationBuffer_BufferValidator) getBuffer() *Geom_Geometry {
	if bv.buffer == nil {
		bv.buffer = bv.getOriginal().BufferWithQuadrantSegments(bv.bufferDistance, operationBuffer_BufferValidator_QUADRANT_SEGMENTS_1)
		_, isGeomCollection := bv.buffer.GetChild().(*Geom_GeometryCollection)
		if isGeomCollection && bv.buffer.IsEmpty() {
			// #contains doesn't work with GeometryCollections [Jon Aquino 10/29/2003]
			geom, err := bv.wktReader.Read("POINT EMPTY")
			if err != nil {
				Util_Assert_ShouldNeverReachHere()
			}
			bv.buffer = geom
		}
	}
	return bv.buffer
}

func (bv *operationBuffer_BufferValidator) addContainsTest() {
	bv.addTest(&operationBuffer_BufferValidator_containsTest{
		operationBuffer_BufferValidator_baseTest: operationBuffer_BufferValidator_baseTest{name: "Contains Test", priority: 2},
	})
}

type operationBuffer_BufferValidator_containsTest struct {
	operationBuffer_BufferValidator_baseTest
}

func (ct *operationBuffer_BufferValidator_containsTest) Test(t *testing.T, bv *operationBuffer_BufferValidator) error {
	original := bv.getOriginal()
	// Skip for GeometryCollection
	if _, isGeomCollection := original.GetChild().(*Geom_GeometryCollection); isGeomCollection {
		return nil
	}
	if !original.IsValid() {
		return fmt.Errorf("original geometry is not valid")
	}
	buffer := bv.getBuffer()
	if bv.bufferDistance > 0 {
		if !ct.contains(buffer, original) {
			return fmt.Errorf("Expected buffer to contain original")
		}
	} else {
		if !ct.contains(original, buffer) {
			return fmt.Errorf("Expected original to contain buffer")
		}
	}
	return nil
}

func (ct *operationBuffer_BufferValidator_containsTest) contains(a, b *Geom_Geometry) bool {
	// JTS doesn't currently handle empty geometries correctly [Jon Aquino 10/29/2003]
	if b.IsEmpty() {
		return true
	}
	return a.Contains(b)
}

// addBufferResultValidatorTest is commented out in Java
// func (bv *operationBuffer_BufferValidator) addBufferResultValidatorTest() { ... }
