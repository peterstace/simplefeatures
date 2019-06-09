package simplefeatures_test

import (
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

var wktCorpus = map[int]string{
	// Point
	101: "POINT EMPTY",
	102: "POINT(0 0)",
	// Make sure floats that can't be precisely represented work.
	103: "POINT(0.1 0.1)",

	// LineString
	201: "LINESTRING EMPTY",
	202: "LINESTRING(0 0,1 1)",

	// Polygon
	301: "POLYGON EMPTY",
	302: "POLYGON((0 0,1 1,1 0,0 0))",
	303: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
	304: "POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))",
	305: "POLYGON((0 0,5 0,5 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1),(1 3,1 4,2 4,2 3,1 3))",

	// MultiPoint
	401: "MULTIPOINT EMPTY",
	402: "MULTIPOINT((0 0))",
	403: "MULTIPOINT((1 1))",
	404: "MULTIPOINT((1 1),(1 1))",
	405: "MULTIPOINT((1 1),(2 2))",
	406: "MULTIPOINT((1 1),(2 2),(3 3))",
	407: "MULTIPOINT((1 1),EMPTY,(3 3))",
	408: "MULTIPOINT(EMPTY)",

	// MultiLineString
	501: "MULTILINESTRING EMPTY",
	502: "MULTILINESTRING(EMPTY)",
	503: "MULTILINESTRING(EMPTY,EMPTY)",
	504: "MULTILINESTRING((0 0,1 1))",
	505: "MULTILINESTRING((0 0,1 1),(0 0,2 2,3 3))",

	// MultiPolygon
	601: "MULTIPOLYGON EMPTY",
	602: "MULTIPOLYGON(EMPTY)",
	603: "MULTIPOLYGON(EMPTY,EMPTY)",
	604: "MULTIPOLYGON(EMPTY,((0 0,1 1,1 0,0 0)))",
	605: "MULTIPOLYGON(((0 0,1 1,1 0,0 0)),((0 0,1 0,1 1,0 1,0 0)))",
	606: "MULTIPOLYGON(((0 0,5 0,5 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1),(1 3,1 4,2 4,2 3,1 3)))",

	701: "GEOMETRYCOLLECTION EMPTY",
	702: "GEOMETRYCOLLECTION(POINT EMPTY)",
	703: "GEOMETRYCOLLECTION(POINT EMPTY,POINT(0 0))",
	704: "GEOMETRYCOLLECTION(POINT EMPTY,POINT(0 0),POLYGON EMPTY)",
}

func TestWKTIdentity(t *testing.T) {
	for id, wkt := range wktCorpus {
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			geom, err := UnmarshalWKT(strings.NewReader(wkt))
			if err != nil {
				t.Fatalf("could not unmarshal WKT: %v", err)
			}
			out := geom.AsText()
			if string(out) != wkt {
				t.Errorf("WKTs are different:\ninput:  %s\noutput: %s", wkt, string(out))
			}
		})
	}
}
