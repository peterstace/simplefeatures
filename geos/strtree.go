package geos

/*
#cgo LDFLAGS: -lgeos_c
#cgo CFLAGS: -Wall
#include "geos_c.h"
#include "stdint.h"
#include "stdlib.h"
#include "string.h"

extern void queryCallback(void *item, void *userdata);
*/
import "C"

import (
	"fmt"
	"runtime/cgo"
	"unsafe"

	"github.com/peterstace/simplefeatures/geom"
)

type STRTree struct {
	h       *handle
	tree    *C.GEOSSTRtree
	geoms   []*C.GEOSGeometry
	indices *C.char
	closed  bool
}

func NewSTRTree(nodeCapacity int, boxes []geom.Envelope) (*STRTree, error) {
	h, err := newHandle()
	if err != nil {
		return nil, err
	}

	tree := &STRTree{
		h:       h,
		tree:    nil, // Populated below.
		geoms:   nil, // Populated below.
		indices: nil, // Populated below.
		closed:  false,
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
		tree.Close()
		return nil, wrap(h.err(), "executing GEOSSTRtree_create_r")
	}

	tree.indices = (*C.char)(C.malloc(C.sizeof_char * C.size_t(len(boxes))))
	C.memset(unsafe.Pointer(tree.indices), 0, C.sizeof_char*C.size_t(len(boxes)))

	for i, b := range boxes {
		// TODO: what if the envelope is empty?
		gh, err := h.createGeometryHandle(b.BoundingDiagonal())
		if err != nil {
			tree.Close()
			return nil, wrap(err, "creating entry bbox")
		}
		userData := unsafe.Pointer(uintptr(unsafe.Pointer(tree.indices)) + C.sizeof_char*uintptr(i))
		C.GEOSSTRtree_insert_r(tree.h.context, tree.tree, gh, userData)
		tree.geoms = append(tree.geoms, gh) // TODO: only need to keep geoms if >= 3.9.0
	}

	return tree, nil
}

// TODO: add tests to ensure no memory leaks
func (t *STRTree) Close() error {
	if t.closed {
		return fmt.Errorf("already closed")
	}

	if t.tree != nil {
		C.GEOSSTRtree_destroy_r(t.h.context, t.tree)
	}
	if t.indices != nil {
		C.free(unsafe.Pointer(t.indices))
	}
	for _, gh := range t.geoms {
		C.GEOSGeom_destroy(gh)
	}
	t.geoms = nil
	t.h.release()

	t.closed = true
	return nil
}

func (t *STRTree) Iterate(callback func(ridx int)) {
	if t.closed {
		panic("STRTree is closed")
	}

	userData := cgo.NewHandle(callbackUserData{
		indices:  t.indices,
		callback: callback,
	})
	defer userData.Delete()

	C.GEOSSTRtree_iterate_r(
		t.h.context,
		t.tree,
		(C.GEOSQueryCallback)(C.queryCallback),
		unsafe.Pointer(userData),
	)
}

type callbackUserData struct {
	indices  *C.char
	callback func(int)
}

//export queryCallback
func queryCallback(item, userData unsafe.Pointer) {
	ud := (cgo.Handle(userData).Value()).(callbackUserData)
	ridx := uintptr(item) - uintptr(unsafe.Pointer(ud.indices))
	ud.callback(int(ridx))
}
