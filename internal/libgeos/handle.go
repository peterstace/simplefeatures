package libgeos

import (
	"bytes"

	"github.com/peterstace/simplefeatures/geom"
)

/*
#cgo linux CFLAGS: -I/usr/include
#cgo linux LDFLAGS: -L/usr/lib -lgeos_c
#include "geos_c.h"
#include <stdarg.h>
#include <stdio.h>

void sf_notice_handler(const char *format, ...) {
    va_list args;
    va_start(args, format);
    fprintf(stderr, "NOTICE: ");
    vfprintf(stderr, format, args);
    va_end(args);
}

//XXX: error handler

GEOSContextHandle_t sf_initGEOS() {
	return initGEOS_r(sf_notice_handler, sf_notice_handler);
}

*/
import "C"

type Handle struct {
	h C.GEOSContextHandle_t
}

func NewHandle() *Handle {
	handle := &Handle{C.GEOS_init_r()}
	//handle.h
	// TODO: set notice and error handlers
}

func (h Handle) Close() {
	C.GEOS_finish_r(h.h)
}

func (h Handle) createGeomHandle(g geom.Geometry) (*C.GEOSGeometry, func()) {
	var wkb bytes.Buffer
	if err := g.AsBinary(&wkb); err != nil {
		panic(err) // can't fail writing to a buffer
	}

	reader := C.GEOSWKBReader_create_r(h.h)
	defer C.GEOSWKBReader_destroy_r(h.h, reader)
	gh := C.GEOSWKBReader_read_r(
		h.h,
		reader,
		(*C.uchar)(&wkb.Bytes()[0]),
		C.ulong(wkb.Len()),
	)
	return gh, func() { C.GEOSGeom_destroy(gh) }
}

func (h Handle) AsText(g geom.Geometry) (string, error) {
	gh, destroy := h.createGeomHandle(g)
	defer destroy()

	writer := C.GEOSWKTWriter_create_r(h.h)
	defer C.GEOSWKTWriter_destroy_r(h.h, writer)
	wkt := C.GEOSWKTWriter_write_r(h.h, writer, gh)
	return C.GoString(wkt), nil
}
