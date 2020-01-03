package geom_test

import (
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestNotEquals(t *testing.T) {
	wkts := []string{
		"POINT EMPTY",

		"POINT(1 2)",
		"POINT(2 1)",

		"MULTIPOINT(1 2,2 1)",
		"MULTIPOINT(1 2,3 1)",

		"LINESTRING(0 0,1 1)",

		"POLYGON((0 0,0 1,1 0,0 0))",
	}
	for i := range wkts {
		for j := range wkts {
			if i == j {
				continue
			}
			t.Run(wkts[i]+"=="+wkts[j], func(t *testing.T) {
				g1, err := UnmarshalWKT(strings.NewReader(wkts[i]))
				if err != nil {
					t.Fatal(err)
				}
				g2, err := UnmarshalWKT(strings.NewReader(wkts[j]))
				if err != nil {
					t.Fatal(err)
				}
				eq, err := g1.Equals(g2)
				if err != nil {
					t.Fatal(err)
				}
				if eq {
					t.Errorf("expected not be be equal")
				}
			})
		}
	}
}

func TestEquals(t *testing.T) {
	for _, equalSet := range [][]string{
		{"POINT (1 1)"},
		{"POINT (2 3)"},
		{
			"MULTIPOINT(1 2,2 1)",
			"MULTIPOINT(2 1,1 2)",
			"MULTIPOINT(2 1,1 2,2 1)",
		},
		{
			"POINT EMPTY",
			"LINESTRING EMPTY",
			"POLYGON EMPTY",
			"MULTIPOINT EMPTY",
			"MULTILINESTRING EMPTY",
			"MULTIPOLYGON EMPTY",
			"GEOMETRYCOLLECTION EMPTY",
			"GEOMETRYCOLLECTION(POINT EMPTY)",
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY)",
		},
	} {
		for i := range equalSet {
			for j := range equalSet {
				t.Run(equalSet[i]+"=="+equalSet[j], func(t *testing.T) {
					g1, err := UnmarshalWKT(strings.NewReader(equalSet[i]))
					if err != nil {
						t.Fatal(err)
					}
					g2, err := UnmarshalWKT(strings.NewReader(equalSet[j]))
					if err != nil {
						t.Fatal(err)
					}
					eq, err := g1.Equals(g2)
					if err != nil {
						t.Fatal(err)
					}
					if !eq {
						t.Errorf("expected to be equal")
					}
				})
			}
		}
	}
}
