package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestPolygonNodeTopology(t *testing.T) {
	tests := []struct {
		name       string
		wktA       string
		wktB       string
		isCrossing bool
	}{
		{
			name:       "Crossing",
			wktA:       "LINESTRING (500 1000, 1000 1000, 1000 1500)",
			wktB:       "LINESTRING (1000 500, 1000 1000, 500 1500)",
			isCrossing: true,
		},
		{
			name:       "NonCrossingQuadrant2",
			wktA:       "LINESTRING (500 1000, 1000 1000, 1000 1500)",
			wktB:       "LINESTRING (300 1200, 1000 1000, 500 1500)",
			isCrossing: false,
		},
		{
			name:       "NonCrossingQuadrant4",
			wktA:       "LINESTRING (500 1000, 1000 1000, 1000 1500)",
			wktB:       "LINESTRING (1000 500, 1000 1000, 1500 1000)",
			isCrossing: false,
		},
		{
			name:       "NonCrossingCollinear",
			wktA:       "LINESTRING (3 1, 5 5, 9 9)",
			wktB:       "LINESTRING (2 1, 5 5, 9 9)",
			isCrossing: false,
		},
		{
			name:       "NonCrossingBothCollinear",
			wktA:       "LINESTRING (3 1, 5 5, 9 9)",
			wktB:       "LINESTRING (3 1, 5 5, 9 9)",
			isCrossing: false,
		},
	}

	reader := jts.Io_NewWKTReader()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := readPts(t, reader, tt.wktA)
			b := readPts(t, reader, tt.wktB)
			// assert: a[1] = b[1]
			got := jts.Algorithm_PolygonNodeTopology_IsCrossing(a[1], a[0], a[2], b[0], b[2])
			if got != tt.isCrossing {
				t.Errorf("IsCrossing() = %v, want %v", got, tt.isCrossing)
			}
		})
	}
}

func TestPolygonNodeTopologyInteriorSegment(t *testing.T) {
	tests := []struct {
		name       string
		wktA       string
		wktB       string
		isInterior bool
	}{
		{
			name:       "InteriorSegment",
			wktA:       "LINESTRING (5 9, 5 5, 9 5)",
			wktB:       "LINESTRING (5 5, 0 0)",
			isInterior: true,
		},
		{
			name:       "ExteriorSegment",
			wktA:       "LINESTRING (5 9, 5 5, 9 5)",
			wktB:       "LINESTRING (5 5, 9 9)",
			isInterior: false,
		},
	}

	reader := jts.Io_NewWKTReader()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := readPts(t, reader, tt.wktA)
			b := readPts(t, reader, tt.wktB)
			// assert: a[1] = b[0]
			got := jts.Algorithm_PolygonNodeTopology_IsInteriorSegment(a[1], a[0], a[2], b[1])
			if got != tt.isInterior {
				t.Errorf("IsInteriorSegment() = %v, want %v", got, tt.isInterior)
			}
		})
	}
}

func readPts(t *testing.T, reader *jts.Io_WKTReader, wkt string) []*jts.Geom_Coordinate {
	t.Helper()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to read WKT: %v", err)
	}
	line := java.Cast[*jts.Geom_LineString](geom)
	return line.GetCoordinates()
}
