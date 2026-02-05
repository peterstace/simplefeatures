package geom

import (
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

// Buffer returns a geometry that contains all points within the given radius
// of the input geometry.
//
// In GIS, the positive (or negative) buffer of a geometry is defined as
// the Minkowski sum (or difference) of the geometry with a circle of radius
// equal to the absolute value of the buffer distance.
//
// The buffer operation always returns a polygonal result.
// The negative or zero-distance buffer of lines and points is always an empty
// Polygon.
//
// Since true buffer curves may contain circular arcs, computed buffer polygons
// are only approximations to the true geometry. The user can control the
// accuracy of the approximation by specifying the number of segments used to
// approximate quarter circles (via BufferQuadSegments).
//
// An error may be returned in pathological cases of numerical degeneracy.
func Buffer(g Geometry, radius float64, opts ...BufferOption) (Geometry, error) {
	optSet := newBufferOptionSet(opts)

	params := jts.OperationBuffer_NewBufferParameters()
	params.SetQuadrantSegments(optSet.quadSegments)
	params.SetEndCapStyle(optSet.endCapStyle)
	params.SetJoinStyle(optSet.joinStyle)
	params.SetMitreLimit(optSet.mitreLimit)
	params.SetSingleSided(optSet.isSingleSided)
	params.SetSimplifyFactor(optSet.simplifyFactor)

	var result Geometry
	err := catch(func() error {
		wkbReader := jts.Io_NewWKBReader()
		jtsG, err := wkbReader.ReadBytes(g.AsBinary())
		if err != nil {
			return wrap(err, "converting geometry to JTS")
		}
		jtsResult := jts.OperationBuffer_BufferOp_BufferOpWithParams(jtsG, radius, params)
		wkbWriter := jts.Io_NewWKBWriter()
		result, err = UnmarshalWKB(wkbWriter.Write(jtsResult), NoValidate{})
		return wrap(err, "converting JTS buffer result to simplefeatures")
	})
	return result, err
}

// BufferOption allows the behaviour of the Buffer operation to be modified.
type BufferOption func(*bufferOptionSet)

type bufferOptionSet struct {
	quadSegments   int
	endCapStyle    int
	joinStyle      int
	mitreLimit     float64
	isSingleSided  bool
	simplifyFactor float64
}

func newBufferOptionSet(opts []BufferOption) bufferOptionSet {
	bos := bufferOptionSet{
		quadSegments:   jts.OperationBuffer_BufferParameters_DEFAULT_QUADRANT_SEGMENTS,
		endCapStyle:    jts.OperationBuffer_BufferParameters_CAP_ROUND,
		joinStyle:      jts.OperationBuffer_BufferParameters_JOIN_ROUND,
		mitreLimit:     jts.OperationBuffer_BufferParameters_DEFAULT_MITRE_LIMIT,
		isSingleSided:  false,
		simplifyFactor: jts.OperationBuffer_BufferParameters_DEFAULT_SIMPLIFY_FACTOR,
	}
	for _, opt := range opts {
		opt(&bos)
	}
	return bos
}

// BufferQuadSegments sets the number of segments used to approximate a quarter
// circle. It defaults to 8.
func BufferQuadSegments(quadSegments int) BufferOption {
	return func(bos *bufferOptionSet) {
		bos.quadSegments = quadSegments
	}
}

// BufferEndCapRound sets the end cap style to 'round'. It is 'round' by
// default.
func BufferEndCapRound() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.endCapStyle = jts.OperationBuffer_BufferParameters_CAP_ROUND
	}
}

// BufferEndCapFlat sets the end cap style to 'flat'. It is 'round' by default.
func BufferEndCapFlat() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.endCapStyle = jts.OperationBuffer_BufferParameters_CAP_FLAT
	}
}

// BufferEndCapSquare sets the end cap style to 'square'. It is 'round' by
// default.
func BufferEndCapSquare() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.endCapStyle = jts.OperationBuffer_BufferParameters_CAP_SQUARE
	}
}

// BufferJoinStyleRound sets the join style to 'round'. It is 'round' by
// default.
func BufferJoinStyleRound() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.joinStyle = jts.OperationBuffer_BufferParameters_JOIN_ROUND
	}
}

// BufferJoinStyleMitre sets the join style to 'mitre'. It is 'round' by
// default. The mitreLimit controls how far a mitre join can extend from the
// join point. Corners with a ratio which exceed the limit will be beveled.
func BufferJoinStyleMitre(mitreLimit float64) BufferOption {
	return func(bos *bufferOptionSet) {
		bos.joinStyle = jts.OperationBuffer_BufferParameters_JOIN_MITRE
		bos.mitreLimit = mitreLimit
	}
}

// BufferJoinStyleBevel sets the join style to 'bevel'. It is 'round' by
// default.
func BufferJoinStyleBevel() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.joinStyle = jts.OperationBuffer_BufferParameters_JOIN_BEVEL
	}
}

// BufferSingleSided sets whether the computed buffer should be single-sided.
// A single-sided buffer is constructed on only one side of each input line.
// The side is determined by the sign of the buffer distance: positive
// indicates the left-hand side, negative indicates the right-hand side.
// The end cap style is ignored for single-sided buffers and forced to flat.
func BufferSingleSided() BufferOption {
	return func(bos *bufferOptionSet) {
		bos.isSingleSided = true
	}
}

// BufferSimplifyFactor sets the factor used to determine the simplify distance
// tolerance for input simplification. The factor is multiplied by the buffer
// distance to get the simplification tolerance. Simplifying can increase the
// performance of computing buffers. Values between 0.01 and 0.1 produce
// relatively good accuracy for the generated buffer. Larger values sacrifice
// accuracy in return for performance. The default is 0.01.
func BufferSimplifyFactor(factor float64) BufferOption {
	return func(bos *bufferOptionSet) {
		if factor < 0 {
			bos.simplifyFactor = 0
		} else {
			bos.simplifyFactor = factor
		}
	}
}
