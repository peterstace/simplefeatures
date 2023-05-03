#include <stdio.h>
#include <geos_c.h>

int main() {
	initGEOS(NULL, NULL);
	GEOSWKTReader *reader = GEOSWKTReader_create();
	GEOSWKTWriter *writer = GEOSWKTWriter_create();
	GEOSWKTWriter_setTrim(writer, 1);

	char *inputWKT = "GEOMETRYCOLLECTION (POLYGON ((0 0, 0 1, 1 0, 0 0)), POLYGON ((0 0, 2 0, 0 -2, 0 0)))";
	GEOSGeometry *input = GEOSWKTReader_read(reader, inputWKT);

	GEOSGeometry *output = GEOSCoverageUnion(input);
	char *outputWKT = GEOSWKTWriter_write(writer, output);

	char isValid = GEOSisValid(output);

	printf("input:  %s\n", inputWKT);
	printf("output: %s\n", outputWKT);
	printf("valid:  %d\n", isValid);

	GEOSGeom_destroy(input);
	GEOSGeom_destroy(output);
	GEOSWKTWriter_destroy(writer);
	GEOSWKTReader_destroy(reader);
	finishGEOS();
}
