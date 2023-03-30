#include "bridge_geos.h"

#define USE_UNSTABLE_GEOS_CPP_API 1
#include <geos/io/WKBReader.h>
#include <geos/io/WKBWriter.h>
#include <geos/geom/Geometry.h>

#include <cstring>
#include <sstream>

void sf_delete_uchar_buffer(unsigned char *p) {
	delete p;
}

void sf_delete_char_buffer(char *p) {
	delete p;
}

void sf_union(
	unsigned char *g1b,
	size_t g1n,
	unsigned char *g2b,
	size_t g2n,
	unsigned char **g3b,
	size_t *g3n,
	char **estr
) {
	try {
		geos::io::WKBReader reader;
		auto g1 = reader.read(g1b, g1n);
		auto g2 = reader.read(g2b, g2n);

		auto g3 = g1->Union(g2.get());

		std::stringstream ss;
		geos::io::WKBWriter writer;
		writer.write(*g3, ss);

		auto tmp = ss.str();
		unsigned char *buf = new unsigned char[tmp.length()];
		std:memcpy(buf, tmp.c_str(), tmp.length());
		*g3b = buf;
		*g3n = tmp.length();
	}
	catch (std::exception const &e) {
		*estr = new char[strlen(e.what()) + 1];
		strcpy(*estr, e.what());
	}
	catch (...) {
		const char *msg = "unhandled exception";
		*estr = new char[strlen(msg) + 1];
		strcpy(*estr, msg);
	}
}
