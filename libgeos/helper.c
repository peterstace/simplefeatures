#include <string.h>
#include "geos_c.h"

// marshal converts the supplied geometry to either WKB (prefered) or WKT
// (fallback). The result is either a WKB or a WKT (or NULL, indicating that
// something went wrong). The size param is set to the number of bytes in the
// returned WKB or WKT. The isWKT param is either set to 0 (WKB) or 1 (WKT),
// indicating the type of marshalling used.
unsigned char *marshal(
	GEOSContextHandle_t handle,
	const GEOSGeometry *g,
	size_t *size,
	char *isWKT
) {
	// Try WKB first. This will fail if the geometry contains an empty Point.
	GEOSWKBWriter *wkbWriter = GEOSWKBWriter_create_r(handle);
	if (!wkbWriter) {
		return NULL;
	}
	unsigned char *wkb = GEOSWKBWriter_write_r(handle, wkbWriter, g, size);
	GEOSWKBWriter_destroy_r(handle, wkbWriter);
	if (wkb) {
		*isWKT = 0;
		return wkb;
	}

	// Try WKT. This should work for all geometries, but will be a bit slower.
	GEOSWKTWriter *wktWriter = GEOSWKTWriter_create_r(handle);
	if (!wktWriter) {
		return NULL;
	}
	unsigned char *wkt = GEOSWKTWriter_write_r(handle, wktWriter, g);
	GEOSWKTWriter_destroy_r(handle, wktWriter);
	if (wkt) {
		*size = strlen(wkt);
		*isWKT = 1;
		return wkt;
	}

	return NULL;
}
