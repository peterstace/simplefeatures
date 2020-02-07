package libgeos

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
import (
	"bytes"
	"sync"

	"github.com/peterstace/simplefeatures/geom"
)

var (
	handle C.GEOSContextHandle_t
	mu     sync.Mutex
)

func init() {
	handle = C.sf_initGEOS()
}

func AsText(g geom.Geometry) (string, error) {
	var wkb bytes.Buffer
	if err := g.AsBinary(&wkb); err != nil {
		return "", err
	}

	mu.Lock()
	defer mu.Unlock()

	reader := C.GEOSWKBReader_create_r(handle)
	defer C.GEOSWKBReader_destroy_r(handle, reader)
	geomHandle := C.GEOSWKBReader_read_r(
		handle,
		reader,
		(*C.uchar)(&wkb.Bytes()[0]),
		C.ulong(wkb.Len()),
	)
	defer C.GEOSGeom_destroy(geomHandle)

	writer := C.GEOSWKTWriter_create_r(handle)
	defer C.GEOSWKTWriter_destroy_r(handle, writer)
	wkt := C.GEOSWKTWriter_write_r(handle, writer, geomHandle)
	return C.GoString(wkt), nil
}
