package simplefeatures_test

import (
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

type OptionalBool struct {
	Set bool
	Val bool
}

var (
	Yes = OptionalBool{true, true}
	No  = OptionalBool{true, false}
)

type corpusTestCase struct {
	WKT      string
	IsSimple OptionalBool
}

var wktCorpus = map[int]corpusTestCase{
	// Point
	101: {
		WKT: "POINT EMPTY",
	},
	102: {
		WKT: "POINT(0 0)",
	},
	// Make sure floats that can't be precisely represented work.
	103: {
		WKT: "POINT(0.1 0.1)",
	},

	// LineString
	201: {
		WKT:      "LINESTRING EMPTY",
		IsSimple: Yes,
	},
	202: {
		WKT:      "LINESTRING(0 0,1 2)",
		IsSimple: Yes,
	},
	203: {
		WKT:      "LINESTRING(0 0,1 1,1 1)",
		IsSimple: Yes,
	},
	204: {
		WKT:      "LINESTRING(0 0,0 0,1 1)",
		IsSimple: Yes,
	},
	205: {
		WKT:      "LINESTRING(0 0,1 1,0 0)",
		IsSimple: No,
	},
	206: {
		WKT:      "LINESTRING(0 0,1 1,0 1)",
		IsSimple: Yes,
	},
	207: {
		WKT:      "LINESTRING(0 0,1 1,0 1,0 0)",
		IsSimple: Yes,
	},
	208: {
		WKT:      "LINESTRING(0 0,1 1,0 1,1 0)",
		IsSimple: No,
	},
	209: {
		WKT:      "LINESTRING(0 0,1 1,0 1,1 0,0 0)",
		IsSimple: No,
	},
	210: {
		WKT:      "LINESTRING(0 0,1 1,0 1,1 0,2 0)",
		IsSimple: No,
	},
	211: {
		WKT:      "LINESTRING(0 0,1 1,0 1,0 0,1 1)",
		IsSimple: No,
	},
	212: {
		WKT:      "LINESTRING(0 0,1 1,0 1,0 0,2 2)",
		IsSimple: No,
	},
	213: {
		WKT:      "LINESTRING(1 1,2 2,0 0)",
		IsSimple: No,
	},
	214: {
		WKT:      "LINESTRING(1 1,2 2,3 2,3 3,0 0)",
		IsSimple: No,
	},

	// Polygon
	301: {
		WKT: "POLYGON EMPTY",
	},
	302: {
		WKT: "POLYGON((0 0,1 1,1 0,0 0))",
	},
	303: {
		WKT: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
	},
	304: {
		WKT: "POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))",
	},
	305: {
		WKT: "POLYGON((0 0,5 0,5 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1),(1 3,1 4,2 4,2 3,1 3))",
	},

	// MultiPoint
	401: {
		WKT: "MULTIPOINT EMPTY",
	},
	402: {
		WKT: "MULTIPOINT((0 0))",
	},
	403: {
		WKT: "MULTIPOINT((1 1))",
	},
	404: {
		WKT: "MULTIPOINT((1 1),(1 1))",
	},
	405: {
		WKT: "MULTIPOINT((1 1),(2 2))",
	},
	406: {
		WKT: "MULTIPOINT((1 1),(2 2),(3 3))",
	},
	407: {
		WKT: "MULTIPOINT((1 1),EMPTY,(3 3))",
	},
	408: {
		WKT: "MULTIPOINT(EMPTY)",
	},

	// MultiLineString
	501: {
		WKT: "MULTILINESTRING EMPTY",
	},
	502: {
		WKT: "MULTILINESTRING(EMPTY)",
	},
	503: {
		WKT: "MULTILINESTRING(EMPTY,EMPTY)",
	},
	504: {
		WKT: "MULTILINESTRING((0 0,1 1))",
	},
	505: {
		WKT: "MULTILINESTRING((0 0,1 1),(0 0,2 2,3 3))",
	},

	// MultiPolygon
	601: {
		WKT: "MULTIPOLYGON EMPTY",
	},
	602: {
		WKT: "MULTIPOLYGON(EMPTY)",
	},
	603: {
		WKT: "MULTIPOLYGON(EMPTY,EMPTY)",
	},
	604: {
		WKT: "MULTIPOLYGON(EMPTY,((0 0,1 1,1 0,0 0)))",
	},
	605: {
		WKT: "MULTIPOLYGON(((0 0,1 1,1 0,0 0)),((0 0,1 0,1 1,0 1,0 0)))",
	},
	606: {
		WKT: "MULTIPOLYGON(((0 0,5 0,5 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1),(1 3,1 4,2 4,2 3,1 3)))",
	},

	// GeometryCollection
	701: {
		WKT: "GEOMETRYCOLLECTION EMPTY",
	},
	702: {
		WKT: "GEOMETRYCOLLECTION(POINT EMPTY)",
	},
	703: {
		WKT: "GEOMETRYCOLLECTION(POINT EMPTY,POINT(0 0))",
	},
	704: {
		WKT: "GEOMETRYCOLLECTION(POINT EMPTY,POINT(0 0),POLYGON EMPTY)",
	},
}

func TestWKTIdentity(t *testing.T) {
	for id, tt := range wktCorpus {
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			geom, err := UnmarshalWKT(strings.NewReader(tt.WKT))
			if err != nil {
				t.Fatalf("could not unmarshal WKT: %v", err)
			}
			out := geom.AsText()
			if string(out) != tt.WKT {
				t.Errorf("WKTs are different:\ninput:  %s\noutput: %s", tt.WKT, string(out))
			}
		})
	}
}

func TestIsSimple(t *testing.T) {
	for id, tt := range wktCorpus {
		if !tt.IsSimple.Set {
			continue
		}
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			geom, err := UnmarshalWKT(strings.NewReader(tt.WKT))
			if err != nil {
				t.Fatalf("could not unmarshal WKT: %v", err)
			}
			got := geom.IsSimple()
			if got != tt.IsSimple.Val {
				t.Errorf("got=%v want=%v", got, tt.IsSimple.Val)
			}
		})
	}
}
