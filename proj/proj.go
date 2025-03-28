package proj

/*
#cgo LDFLAGS: -lproj
#include <proj.h>
*/
import "C"
import (
	"fmt"

	"github.com/peterstace/simplefeatures/geom"
)

type Transformation struct {
	ctx *C.PJ_CONTEXT
	pj  *C.PJ
}

func NewTransformation(sourceCRS, targetCRS string) (*Transformation, error) {
	c := C.proj_context_create()

	p := C.proj_create_crs_to_crs(c, C.CString(sourceCRS), C.CString(targetCRS), nil)
	if p == nil {
		errno := C.proj_context_errno(c)
		errstr := C.GoString(C.proj_context_errno_string(c, errno))
		return nil, fmt.Errorf("proj_create failed: %s", errstr)
	}
	defer C.proj_destroy(p)

	n := C.proj_normalize_for_visualization(c, p)
	if n == nil {
		errno := C.proj_context_errno(c)
		errstr := C.GoString(C.proj_context_errno_string(c, errno))
		return nil, fmt.Errorf("proj_normalize_for_visualization failed: %s", errstr)
	}
	return &Transformation{c, n}, nil
}

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

func (p *Transformation) Forward(ct geom.CoordinatesType, coords []float64) error {
	return p.transform(C.PJ_FWD, ct, coords)
}

func (p *Transformation) Inverse(ct geom.CoordinatesType, coords []float64) error {
	return p.transform(C.PJ_INV, ct, coords)
}

func (p *Transformation) transform(dir C.PJ_DIRECTION, ct geom.CoordinatesType, coords []float64) error {
	if len(coords)%int(ct.Dimension()) != 0 {
		return fmt.Errorf("len(coords) must be a multiple of ct.Dimension()")
	}

	var (
		stride = C.size_t(ct.Dimension())
		count  = C.size_t(len(coords) / int(ct.Dimension()))
	)

	p0 := (*C.double)(&coords[0]) // Always X.
	p1 := (*C.double)(&coords[1]) // Always Y.

	var p2 *C.double
	if ct.Dimension() >= 3 {
		p2 = (*C.double)(&coords[2]) // Z (if has Z), otherwise M.
	}

	var p3 *C.double
	if ct.Dimension() == 4 {
		p3 = (*C.double)(&coords[3]) // Always M (must have Z as well).
	}

	C.proj_trans_generic(
		p.pj, dir,
		p0, stride, count,
		p1, stride, count,
		p2, stride, count,
		p3, stride, count,
	)
	errno := C.proj_errno(p.pj)
	if errno != 0 {
		C.proj_errno_reset(p.pj)
		errstr := C.GoString(C.proj_context_errno_string(p.ctx, errno))
		return fmt.Errorf("proj_trans_generic failed: %s", errstr)
	}
	return nil
}
