package geom

import (
	"strconv"
	"testing"
)

func TestReNode(t *testing.T) {
	for i, tt := range []struct {
		inputA, inputB   string
		outputA, outputB string
	}{
		{
			inputA:  "LINESTRING(0 0,1 1)",
			inputB:  "LINESTRING(0 1,1 0)",
			outputA: "LINESTRING(0 0,0.5 0.5,1 1)",
			outputB: "LINESTRING(0 1,0.5 0.5,1 0)",
		},
		{
			inputA:  "LINESTRING(0 0,0.5 0.5)",
			inputB:  "LINESTRING(0 0,1 1)",
			outputA: "LINESTRING(0 0,0.5 0.5)",
			outputB: "LINESTRING(0 0,0.5 0.5,1 1)",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			inA, err := UnmarshalWKT(tt.inputA)
			if err != nil {
				t.Fatal(err)
			}
			inB, err := UnmarshalWKT(tt.inputB)
			if err != nil {
				t.Fatal(err)
			}
			wantA, err := UnmarshalWKT(tt.outputA)
			if err != nil {
				t.Fatal(err)
			}
			wantB, err := UnmarshalWKT(tt.outputB)
			if err != nil {
				t.Fatal(err)
			}
			gotA, gotB := reNodeGeometries(inA, inB)
			if !gotA.EqualsExact(wantA) || !gotB.EqualsExact(wantB) {
				t.Logf("INPUT A: %v\n", inA.AsText())
				t.Logf("INPUT B: %v\n", inB.AsText())
				t.Logf("WANT  A: %v\n", wantA.AsText())
				t.Logf("WANT  B: %v\n", wantB.AsText())
				t.Logf("GOT   A: %v\n", gotA.AsText())
				t.Logf("GOT   B: %v\n", gotB.AsText())
				t.Error("mismatch")
			}
		})
	}
}
