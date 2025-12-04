package proj

/*
#cgo LDFLAGS: -lproj
#include <stdlib.h>
#include <proj.h>

const char *proj_context_errno_string_v8_onwards_only(PJ_CONTEXT *ctx, int errno)
{
	#if PROJ_VERSION_MAJOR >= 8
		return proj_context_errno_string(ctx, errno);
	#else
		return NULL;
	#endif
}
*/
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"github.com/peterstace/simplefeatures/geom"
)

// Version returns the major, minor, and patch versions of the PROJ library
// that is being used.
func Version() (major, minor, patch int) {
	return C.PROJ_VERSION_MAJOR, C.PROJ_VERSION_MINOR, C.PROJ_VERSION_PATCH
}

// Transformation transforms coordinates from a source coordinate reference
// system (CRS) to a target CRS. Instances of this type are not thread-safe
// (i.e. a single [Transformation] should not be used by multiple goroutines without
// synchronization).
type Transformation struct {
	ctx *C.PJ_CONTEXT
	pj  *C.PJ
}

// NewTransformation creates a new [Transformation] object that can be used to
// transform coordinates. It retains memory resources that aren't managed by
// the Go runtime. The [Transformation.Release] method must be called to free these resources.
// The returned [Transformation] instance is not thread-safe.
//
// The sourceCRS and targetCRS parameters can be in any format accepted by the
// `proj_create_crs_to_crs` function in the PROJ library. Allowed formats
// include (but are not limited to):
//
//   - A PROJ string, e.g. "+proj=longlat +datum=WGS84".
//   - A WKT string describing the CRS.
//   - An AUTHORITY:CODE string, e.g. "EPSG:4326".
//   - The name of a CRS found in the PROJ database, e.g. "NAD27".
func NewTransformation(sourceCRS, targetCRS string) (*Transformation, error) {
	c := C.proj_context_create()

	cStrSourceCRS := C.CString(sourceCRS)
	cStrTargetCRS := C.CString(targetCRS)
	defer C.free(unsafe.Pointer(cStrSourceCRS))
	defer C.free(unsafe.Pointer(cStrTargetCRS))

	p := C.proj_create_crs_to_crs(c, cStrSourceCRS, cStrTargetCRS, nil)
	if p == nil {
		errno := C.proj_context_errno(c)
		errstr := errnoString(c, errno)
		return nil, fmt.Errorf("proj_create: %s", errstr)
	}
	defer C.proj_destroy(p)

	n := C.proj_normalize_for_visualization(c, p)
	if n == nil {
		errno := C.proj_context_errno(c)
		errstr := errnoString(c, errno)
		return nil, fmt.Errorf("proj_normalize_for_visualization: %s", errstr)
	}
	return &Transformation{c, n}, nil
}

// Release releases the resources held by the [Transformation] object that are
// managed outside the Go runtime. It should be called at least once when the
// [Transformation] is no longer needed.
func (p *Transformation) Release() {
	if p.pj != nil {
		C.proj_destroy(p.pj)
		p.pj = nil
	}
	if p.ctx != nil {
		C.proj_context_destroy(p.ctx)
		p.ctx = nil
	}
}

// Forward does an in-place transform of coordinates from the source CRS to the
// target CRS. Coordinates should be laid out in the order X1, Y1, X2, Y2, ...
// (for XY coordinates), X1, Y1, Z1, X2, Y2, Z2, ... (for XYZ coordinates), X1,
// Y1, M1, X2, Y2, M2, ... (for XYM coordinates), or X1, Y1, Z1, M1, X2, Y2,
// Z2, M2, ... (for XYZM coordinates). The length of the coords slice must be a
// multiple of the dimension of the coordinates type.
func (p *Transformation) Forward(ct geom.CoordinatesType, coords []float64) error {
	return p.transform(C.PJ_FWD, ct, coords)
}

// Inverse does an in-place transform of coordinates from the target CRS to the
// source CRS. Its signature operates in the same way as the [Transformation.Forward] method.
func (p *Transformation) Inverse(ct geom.CoordinatesType, coords []float64) error {
	return p.transform(C.PJ_INV, ct, coords)
}

func (p *Transformation) transform(dir C.PJ_DIRECTION, ct geom.CoordinatesType, coords []float64) error {
	if len(coords)%ct.Dimension() != 0 {
		return errors.New("len(coords) must be a multiple of ct.Dimension()")
	}

	var (
		stride = C.size_t(ct.Dimension()) * C.sizeof_double
		count  = C.size_t(len(coords) / ct.Dimension())
	)

	pX := (*C.double)(&coords[0])
	pY := (*C.double)(&coords[1])

	var pZ, pM *C.double
	switch ct {
	case geom.DimXYZ:
		pZ = (*C.double)(&coords[2])
	case geom.DimXYM:
		pM = (*C.double)(&coords[2])
	case geom.DimXYZM:
		pZ = (*C.double)(&coords[2])
		pM = (*C.double)(&coords[3])
	}

	C.proj_trans_generic(
		p.pj, dir,
		pX, stride, count,
		pY, stride, count,
		pZ, stride, count,
		pM, stride, count,
	)
	errno := C.proj_errno(p.pj)
	if errno != 0 {
		C.proj_errno_reset(p.pj)
		errstr := errnoString(p.ctx, errno)
		return fmt.Errorf("proj_trans_generic: %s", errstr)
	}
	return nil
}

var projErrnoStringMu sync.Mutex

func errnoString(ctx *C.PJ_CONTEXT, errno C.int) string {
	// In PROJ 8 and onwards, we can just use the native
	// proj_context_errno_string which is thread-safe.
	if C.PROJ_VERSION_MAJOR >= 8 {
		return C.GoString(C.proj_context_errno_string_v8_onwards_only(ctx, errno))
	}

	// Pre-PROJ 8, proj_context_errno_string is not available. Instead, we have
	// to use proj_errno_string. However, this function is not thread-safe
	// (hence the mutex).
	projErrnoStringMu.Lock()
	defer projErrnoStringMu.Unlock()
	return C.GoString(C.proj_errno_string(errno))
}
