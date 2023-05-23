package geos

/*
#cgo LDFLAGS: -lgeos_c
#cgo CFLAGS: -Wall
#include "geos_c.h"
#include "stdint.h"
#include "stdlib.h"

int32_t *allocate_indices(size_t n) {
	int32_t *ind = malloc(4 * n);
	if (ind == NULL) {
		return NULL;
	}
	for (int i = 0; i < n; i++) {
		ind[i] = i;
	}
	return ind;
}

int32_t *offset_pointer(int32_t *indices, size_t offset) {
	return &indices[offset];
}
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/peterstace/simplefeatures/geom"
)

type STRTree struct {
	h       *handle
	tree    *C.GEOSSTRtree
	geoms   []*C.GEOSGeometry
	indices *C.int32_t
	closed  bool
}

type STRTreeEntry struct {
	BBox geom.Envelope
	Item interface{} // TODO: Use generics here?
}

func NewSTRTree(nodeCapacity int, entries []STRTreeEntry) (*STRTree, error) {
	h, err := newHandle()
	if err != nil {
		return nil, err
	}

	tree := &STRTree{
		h:     h,
		tree:  nil, // Populated below.
		geoms: nil, // Populated below.
	}

	// The upper limit is artificial, but it's a good idea to constrain it to
	// something sensible.
	if nodeCapacity < 2 || nodeCapacity > 64 {
		return nil, fmt.Errorf(
			"node capacity must be between 2 and 64 (inclusive) but was %d",
			nodeCapacity,
		)
	}

	tree.tree = C.GEOSSTRtree_create_r(h.context, C.size_t(nodeCapacity))
	if tree.tree == nil {
		return nil, wrap(err, "executing GEOSSTRtree_create_r")
	}

	tree.indices = C.allocate_indices(C.size_t(nodeCapacity))
	if tree.indices == nil {
		tree.Close()
		return nil, fmt.Errorf("allocating memory for indices")
	}
	for i, e := range entries {
		gh, err := h.createGeometryHandle(e.BBox.BoundingDiagonal())
		if err != nil {
			tree.Close()
			return nil, wrap(err, "creating entry bbox")
		}
		tree.geoms = append(tree.geoms, gh)
		userData := C.offset_pointer(tree.indices, C.size_t(i))
		C.GEOSSTRtree_insert_r(tree.h.context, tree.tree, gh, unsafe.Pointer(userData))
	}

	return tree, nil
}

func (t *STRTree) Close() error {
	if t.closed {
		return fmt.Errorf("already closed")
	}

	for _, gh := range t.geoms {
		C.GEOSGeom_destroy(gh)
	}
	t.geoms = nil

	if t.indices != nil {
		C.free(unsafe.Pointer(t.indices))
	}
	if t.tree != nil {
		C.GEOSSTRtree_destroy_r(t.h.context, t.tree)
	}
	t.h.release()
	t.closed = true
	return nil
}
