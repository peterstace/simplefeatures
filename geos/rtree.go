package geos

/*
#cgo LDFLAGS: -lgeos_c
#cgo CFLAGS: -Wall
#include "geos_c.h"
*/
import "C"

type RTree struct {
	h  *handle
	tr *C.GEOSSTRtree
}

func NewRTree(nodeCapacity int) (*RTree, error) {
	h, err := newHandle()
	if err != nil {
		return nil, err
	}
	// TODO: validate node capacity? Check for negative for when converting to size_t?
	tr := C.GEOSSTRtree_create_r(h.context, C.size_t(nodeCapacity))
	// TODO: check for NULL return?
	return &RTree{h, tr}, nil
}

func (r *RTree) Close() error {
	// TODO: need to clean up anything _inside_ the tree as well.
	C.GEOSSTRtree_destroy_r(r.h.context, r.tr)
	r.h.release()
	return nil
}
