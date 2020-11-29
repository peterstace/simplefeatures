package geom

import (
	"fmt"
	"reflect"
	"testing"
)

type DCELSpec struct {
	NumVerts int
	NumEdges int
	NumFaces int
	Vertices []VertexSpec
	Edges    []EdgeSpec
	Faces    []FaceSpec
}

type FaceSpec struct {
	// Origin and destination of an edge that is incident to the face.
	First  XY
	Second XY
	Cycle  []XY
	Label  uint8
}

type EdgeSpec struct {
	EdgeLabel uint8
	FaceLabel uint8
	Sequence  []XY
}

type VertexSpec struct {
	Label    uint8
	Vertices []XY
}

func CheckDCEL(t *testing.T, dcel *doublyConnectedEdgeList, spec DCELSpec) {
	t.Helper()

	if spec.NumVerts != len(dcel.vertices) {
		t.Fatalf("vertices: want=%d got=%d", spec.NumVerts, len(dcel.vertices))
	}
	var vertsInSpec int
	for _, v := range spec.Vertices {
		vertsInSpec += len(v.Vertices)
	}
	if spec.NumVerts != vertsInSpec {
		t.Fatalf("NumVerts doesn't match vertsInSpec: %d vs %d", spec.NumVerts, vertsInSpec)
	}

	if spec.NumEdges != len(dcel.halfEdges) {
		t.Fatalf("edges: want=%d got=%d", spec.NumEdges, len(dcel.halfEdges))
	}
	if spec.NumEdges != len(spec.Edges) {
		t.Fatalf("NumEdges doesn't match len(spec.Edges): %d vs %d", spec.NumEdges, len(spec.Edges))
	}

	if spec.NumFaces != len(dcel.faces) {
		t.Fatalf("faces: want=%d got=%d", spec.NumFaces, len(dcel.faces))
	}
	if spec.NumFaces != len(spec.Faces) {
		t.Fatalf("NumFaces doesn't match len(spec.Faces): %d vs %d", spec.NumFaces, len(spec.Faces))
	}

	t.Run("vertex_labels", func(t *testing.T) {
		unchecked := make(map[*vertexRecord]bool)
		for _, vert := range dcel.vertices {
			unchecked[vert] = true
		}
		for _, want := range spec.Vertices {
			for _, wantXY := range want.Vertices {
				vert, ok := dcel.vertices[wantXY]
				if !ok {
					t.Errorf("no vertex %v", wantXY)
					continue
				}
				if vert.label != want.Label {
					t.Errorf("vertex label mismatch for %v: want=%b got=%b", wantXY, want.Label, vert.label)
				}
				delete(unchecked, vert)
			}
		}
		if len(unchecked) > 0 {
			for vert := range unchecked {
				t.Logf("unchecked vertex: %v", vert)
			}
			t.Errorf("some vertex labels not checked: %d", len(unchecked))
		}
	})

	t.Run("vertex_incidents", func(t *testing.T) {
		for _, vr := range dcel.vertices {
			bruteForceEdgeSet := make(map[*halfEdgeRecord]struct{})
			for _, er := range dcel.halfEdges {
				if er.origin.coords == vr.coords {
					bruteForceEdgeSet[er] = struct{}{}
				}
			}
			incidentsSet := make(map[*halfEdgeRecord]struct{})
			for _, e := range vr.incidents {
				incidentsSet[e] = struct{}{}
			}
			if !reflect.DeepEqual(bruteForceEdgeSet, incidentsSet) {
				t.Fatalf("vertex record at %v doesn't have correct incidents: "+
					"bruteForceEdgeSet=%v incidentsSet=%v", vr.coords, bruteForceEdgeSet, incidentsSet)
			}
		}
	})

	t.Run("edges", func(t *testing.T) {
		for _, e := range dcel.halfEdges {
			// Find an edge spec that matches e.
			var found bool
			var want EdgeSpec
			for i := range spec.Edges {
				seq := spec.Edges[i].Sequence
				if seq[0] != e.origin.coords {
					continue
				}
				if seq[len(seq)-1] != e.next.origin.coords {
					continue
				}
				if !xysEqual(e.intermediate, seq[1:len(seq)-1]) {
					continue
				}
				want = spec.Edges[i]
				found = true
			}
			if !found {
				t.Fatalf("could not find edge spec matching edge: %v", e)
			}

			if e.edgeLabel != want.EdgeLabel {
				t.Errorf("incorrect edge label for edge %v: "+
					"want=%b got=%b", e, want.EdgeLabel, e.edgeLabel)
			}
			if e.faceLabel != want.FaceLabel {
				t.Errorf("incorrect face label for edge %v: "+
					"want=%b got=%b", e, want.FaceLabel, e.faceLabel)
			}
		}
	})

	for i, want := range spec.Faces {
		t.Run(fmt.Sprintf("face_%d", i), func(t *testing.T) {
			got := findEdge(t, dcel, want.First, want.Second).incident
			CheckCycle(t, got, got.cycle, want.Cycle)
			if want.Label != got.label {
				t.Errorf("face label doesn't match: want=%b got=%b", want.Label, got.label)
			}
		})
	}
}

func xysEqual(a, b []XY) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func findEdge(t *testing.T, dcel *doublyConnectedEdgeList, first, second XY) *halfEdgeRecord {
	for _, e := range dcel.halfEdges {
		if e.origin.coords == first && e.secondXY() == second {
			return e
		}
	}
	t.Fatalf("could not find edge with first %v and second %v", first, second)
	return nil
}

func CheckCycle(t *testing.T, f *faceRecord, start *halfEdgeRecord, want []XY) {
	// Check component matches forward order when following 'next' pointer.
	e := start
	var got []XY
	for {
		if e.incident == nil {
			t.Errorf("half edge has no incident face set")
		} else if e.incident != f {
			t.Errorf("half edge has incorrect incident face")
		}
		if e.origin == nil {
			t.Errorf("edge origin not set")
		}
		got = append(got, e.origin.coords)
		got = append(got, e.intermediate...)
		e = e.next
		if e == start {
			got = append(got, e.origin.coords)
			break
		}
	}
	CheckXYs(t, got, want)

	// Check component matches reverse order when following 'prev' pointer.
	for i := 0; i < len(want)/2; i++ {
		j := len(want) - i - 1
		want[i], want[j] = want[j], want[i]
	}
	var i int
	start = start.prev
	e = start
	got = nil
	for {
		i++
		if i == 100 {
			t.Fatal("inf loop")
		}

		got = append(got, e.next.origin.coords)
		for j := len(e.intermediate) - 1; j >= 0; j-- {
			got = append(got, e.intermediate[j])
		}

		e = e.prev
		if e == start {
			got = append(got, e.next.origin.coords)
			break
		}
	}
	CheckXYs(t, got, want)

	// Check 'twin' assertions.
	e = start
	for {
		if e.twin == nil {
			t.Fatalf("twin not populated")
		}
		if e.twin.twin != e {
			t.Fatalf("twin's twin is not itself")
		}
		if e.origin != e.twin.next.origin {
			t.Fatalf("edge's origin doesn't match twin's next origin")
		}
		if e.next.origin != e.twin.origin {
			t.Fatalf("edge's next origin doesn't match twin's origin ")
		}
		e = e.next
		if e == start {
			break
		}
	}
}

func CheckXYs(t *testing.T, got, want []XY) {
	t.Helper()
	if len(want) != len(got) {
		t.Errorf("XY sequences don't match:\ngot:  %v\nwant: %v", got, want)
		return
	}

	if len(got) < 3 || got[0] != got[len(got)-1] {
		t.Errorf("got not a cycle")
	}
	if len(want) < 3 || want[0] != want[len(want)-1] {
		t.Errorf("want not a cycle")
	}

	// Strip of the part of the cycle that joints back to the
	// start, so that we can run the cycles through offsets
	// looking for a match.
	got = got[:len(got)-1]
	want = want[:len(want)-1]

	n := len(want)
outer:
	for offset := 0; offset < n; offset++ {
		for i := 0; i < n; i++ {
			j := (i + offset) % n
			if got[i] != want[j] {
				continue outer
			}
		}
		return // success, we found an offset that results in the XYs being equal
	}
	t.Errorf("XY sequences don't match:\ngot:  %v\nwant: %v", got, want)
}

func createOverlayFromWKTs(t *testing.T, wktA, wktB string) *doublyConnectedEdgeList {
	gA, err := UnmarshalWKT(wktA)
	if err != nil {
		t.Fatal(err)
	}
	gB, err := UnmarshalWKT(wktB)
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := createOverlay(gA, gB)
	if err != nil {
		t.Fatal(err)
	}
	return overlay
}

func TestGraphTriangle(t *testing.T) {
	poly, err := UnmarshalWKT("POLYGON((0 0,0 1,1 0,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	dcel := newDCELFromMultiPolygon(poly.AsPolygon().AsMultiPolygon(), inputAMask, findInteractionPoints([]Geometry{poly}))

	/*

	  V2 *
	     |\
	     | \
	     |  \
	     |   \
	     |    \   f0
	     |     \
	     | f1   \
	     |       \
	  V0 @--------* V1

	*/

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{0, 1}

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 1,
		NumEdges: 2,
		NumFaces: 0,
		Vertices: []VertexSpec{{
			Label:    inputAPopulated | inputAInSet,
			Vertices: []XY{v0},
		}},
		Edges: []EdgeSpec{
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v0, v1, v2, v0},
			},
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v0, v2, v1, v0},
			},
		},
		Faces: nil,
	})
}

func TestGraphWithHoles(t *testing.T) {
	poly, err := UnmarshalWKT("POLYGON((0 0,5 0,5 5,0 5,0 0),(1 1,2 1,2 2,1 2,1 1),(3 3,4 3,4 4,3 4,3 3))")
	if err != nil {
		t.Fatal(err)
	}

	/*
	         f0

	  v3-------------------v2
	   |                   |
	   |          v9---v10 |
	   |    f1     |f3 |   |
	   |          v8---v11 |
	   |                   |
	   |  v5---v6          |
	   |   |f2 |           |
	   |  v4---v7          |
	   |                   |
	  v0-------------------v1

	*/

	dcel := newDCELFromMultiPolygon(poly.AsPolygon().AsMultiPolygon(), inputBMask, findInteractionPoints([]Geometry{poly}))

	v0 := XY{0, 0}
	v1 := XY{5, 0}
	v2 := XY{5, 5}
	v3 := XY{0, 5}
	v4 := XY{1, 1}
	v5 := XY{1, 2}
	v6 := XY{2, 2}
	v7 := XY{2, 1}
	v8 := XY{3, 3}
	v9 := XY{3, 4}
	v10 := XY{4, 4}
	v11 := XY{4, 3}

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 3,
		NumEdges: 6,
		NumFaces: 0,
		Vertices: []VertexSpec{{
			Label:    inputBPopulated | inputBInSet,
			Vertices: []XY{v0, v4, v8},
		}},
		Edges: []EdgeSpec{
			{
				EdgeLabel: inputBPopulated | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v0, v1, v2, v3, v0},
			},
			{
				EdgeLabel: inputBPopulated | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v0, v3, v2, v1, v0},
			},
			{
				EdgeLabel: inputBPopulated | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v4, v5, v6, v7, v4},
			},
			{
				EdgeLabel: inputBPopulated | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v4, v7, v6, v5, v4},
			},
			{
				EdgeLabel: inputBPopulated | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v8, v9, v10, v11, v8},
			},
			{
				EdgeLabel: inputBPopulated | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v8, v11, v10, v9, v8},
			},
		},
		Faces: nil,
	})
}

func TestGraphWithMultiPolygon(t *testing.T) {
	mp, err := UnmarshalWKT("MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))")
	if err != nil {
		t.Fatal(err)
	}

	/*
	            f0
	  v3-----v2   v7-----v6
	   | f1  |     | f2  |
	   |     |     |     |
	  v0-----v1   v4-----v5
	*/

	dcel := newDCELFromMultiPolygon(mp.AsMultiPolygon(), inputBMask, findInteractionPoints([]Geometry{mp}))

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{1, 1}
	v3 := XY{0, 1}
	v4 := XY{2, 0}
	v5 := XY{3, 0}
	v6 := XY{3, 1}
	v7 := XY{2, 1}

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 2,
		NumEdges: 4,
		NumFaces: 0,
		Vertices: []VertexSpec{{
			Label:    inputBPopulated | inputBInSet,
			Vertices: []XY{v0, v4},
		}},
		Edges: []EdgeSpec{
			{
				EdgeLabel: inputBPopulated | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v0, v1, v2, v3, v0},
			},
			{
				EdgeLabel: inputBPopulated | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v0, v3, v2, v1, v0},
			},
			{
				EdgeLabel: inputBPopulated | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v4, v5, v6, v7, v4},
			},
			{
				EdgeLabel: inputBPopulated | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v4, v7, v6, v5, v4},
			},
		},
		Faces: nil,
	})
}

func TestGraphMultiLineString(t *testing.T) {
	mls, err := UnmarshalWKT("MULTILINESTRING((1 0,0 1,1 2),(2 0,3 1,2 2))")
	if err != nil {
		t.Fatal(err)
	}
	dcel := newDCELFromGeometry(mls, MultiLineString{}, inputAMask, findInteractionPoints([]Geometry{mls}))

	/*
	        v2    v3
	       /        \
	      /          \
	     /            \
	   v1              v4
	     \            /
	      \          /
	       \        /
	        v0    v5
	*/

	v0 := XY{1, 0}
	v1 := XY{0, 1}
	v2 := XY{1, 2}
	v3 := XY{2, 2}
	v4 := XY{3, 1}
	v5 := XY{2, 0}

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 4,
		NumEdges: 4,
		NumFaces: 0,
		Vertices: []VertexSpec{{
			Label:    inputAPopulated | inputAInSet,
			Vertices: []XY{v0, v2, v3, v5},
		}},
		Edges: []EdgeSpec{
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v0, v1, v2},
			},
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v2, v1, v0},
			},
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v3, v4, v5},
			},
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v5, v4, v3},
			},
		},
		Faces: nil,
	})
}

func TestGraphSelfOverlappingLineString(t *testing.T) {
	ls, err := UnmarshalWKT("LINESTRING(0 0,0 1,1 1,1 0,0 1,1 1,2 1)")
	if err != nil {
		t.Fatal(err)
	}
	dcel := newDCELFromGeometry(ls, MultiLineString{}, inputAMask, findInteractionPoints([]Geometry{ls}))

	/*
	   v1----v2----v4
	    |\   |
	    | \  |
	    |  \ |
	    |   \|
	   v0    v3
	*/

	v0 := XY{0, 0}
	v1 := XY{0, 1}
	v2 := XY{1, 1}
	v3 := XY{1, 0}
	v4 := XY{2, 1}

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 4,
		NumEdges: 8,
		NumFaces: 0,
		Vertices: []VertexSpec{{
			Label:    inputAPopulated | inputAInSet,
			Vertices: []XY{v0, v1, v2, v4},
		}},
		Edges: []EdgeSpec{
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v0, v1},
			},
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v1, v0},
			},
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v1, v2},
			},
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v2, v1},
			},
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v1, v3, v2},
			},
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v2, v3, v1},
			},
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v2, v4},
			},
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v4, v2},
			},
		},
		Faces: nil,
	})
}

func TestGraphGhostDeduplication(t *testing.T) {
	ls, err := UnmarshalWKT("LINESTRING(0 0,1 0)")
	if err != nil {
		t.Fatal(err)
	}

	// Ghost contains duplicated lines. This could happen in the scenario where
	// one of the inputs is a multipolygon and the ghost joins the rings, but
	// then the line joining the two input geometries is the same line segment.
	ghost, err := UnmarshalWKT("MULTILINESTRING((0 0,0 1,0 0))")
	if err != nil {
		t.Fatal(err)
	}

	dcel := newDCELFromGeometry(ls, ghost.AsMultiLineString(), inputAMask, findInteractionPoints([]Geometry{ls, ghost}))

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{0, 1}

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 3,
		NumEdges: 4,
		NumFaces: 0,
		Vertices: []VertexSpec{
			{Label: inputAPopulated | inputAInSet, Vertices: []XY{v0, v1}},
			{Label: 0, Vertices: []XY{v2}},
		},
		Edges: []EdgeSpec{
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v0, v1},
			},
			{
				EdgeLabel: inputAPopulated | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v1, v0},
			},
			{
				EdgeLabel: inputAPopulated,
				FaceLabel: 0,
				Sequence:  []XY{v0, v2},
			},
			{
				EdgeLabel: inputAPopulated,
				FaceLabel: 0,
				Sequence:  []XY{v2, v0},
			},
		},
	})
}

func TestGraphOverlayDisjoint(t *testing.T) {
	overlay := createOverlayFromWKTs(t,
		"POLYGON((0 0,1 0,1 1,0 1,0 0))",
		"POLYGON((2 2,2 3,3 3,3 2,2 2))",
	)

	/*
	                v7------v6
	                |        |
	                |   f3   |
	                |        |
	                |        |
	               ,v4------v5
	             ,`
	   v3------v2
	   |     ,` |
	   |f2  `   |       f0
	   |  ,` f1 |
	   | ,      |
	   v0------v1

	*/

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{1, 1}
	v3 := XY{0, 1}
	v4 := XY{2, 2}
	v5 := XY{3, 2}
	v6 := XY{3, 3}
	v7 := XY{2, 3}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 3,
		NumEdges: 10,
		NumFaces: 4,
		Vertices: []VertexSpec{
			{
				Label:    populatedMask | inputAInSet,
				Vertices: []XY{v0, v2},
			},
			{
				Label:    populatedMask | inputBInSet,
				Vertices: []XY{v4},
			},
		},
		Edges: []EdgeSpec{
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v0, v1, v2},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v2, v1, v0},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v2, v3, v0},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v0, v3, v2},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v4, v5, v6, v7, v4},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v4, v7, v6, v5, v4},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: 0,
				Sequence:  []XY{v0, v2},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: 0,
				Sequence:  []XY{v2, v0},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v2, v4},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v4, v2},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v2,
				Second: v1,
				Cycle:  []XY{v2, v1, v0, v3, v2, v4, v7, v6, v5, v4, v2},
				Label:  populatedMask,
			},
			{
				// f1
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v0},
				Label:  populatedMask | inputAInSet,
			},
			{
				// f2
				First:  v2,
				Second: v3,
				Cycle:  []XY{v2, v3, v0, v2},
				Label:  populatedMask | inputAInSet,
			},
			{
				// f3
				First:  v4,
				Second: v5,
				Cycle:  []XY{v4, v5, v6, v7, v4},
				Label:  populatedMask | inputBInSet,
			},
		},
	})
}

func TestGraphOverlayIntersecting(t *testing.T) {
	overlay := createOverlayFromWKTs(t,
		"POLYGON((0 0,1 2,2 0,0 0))",
		"POLYGON((0 1,2 1,1 3,0 1))",
	)

	/*
	           v7
	          /  \
	         /    \
	        /  f2  \
	       /        \
	      /    v3    \
	     /    /  \    \
	    /    / f3 \    \
	   v5--v4------v2--v6
	   |   /        \
	   |f4/    f1    \    f0
	   | /            \
	   |/              \
	   v0--------------v1

	*/

	v0 := XY{0, 0}
	v1 := XY{2, 0}
	v2 := XY{1.5, 1}
	v3 := XY{1, 2}
	v4 := XY{0.5, 1}
	v5 := XY{0, 1}
	v6 := XY{2, 1}
	v7 := XY{1, 3}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 4,
		NumEdges: 14,
		NumFaces: 5,
		Vertices: []VertexSpec{
			{
				Label:    populatedMask | inputAInSet,
				Vertices: []XY{v0},
			},
			{
				Label:    populatedMask | inputBInSet,
				Vertices: []XY{v5},
			},
			{
				Label:    populatedMask | inSetMask,
				Vertices: []XY{v2, v4},
			},
		},
		Edges: []EdgeSpec{
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v4, v0},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v0, v1, v2},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v2, v1, v0},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v0, v4},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v2, v6, v7, v5},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v5, v4},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v4, v5},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v5, v7, v6, v2},
			},
			{
				EdgeLabel: populatedMask | inSetMask,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v2, v4},
			},
			{
				EdgeLabel: populatedMask | inSetMask,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v4, v2},
			},
			{
				EdgeLabel: populatedMask | inSetMask,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v4, v3, v2},
			},
			{
				EdgeLabel: populatedMask | inSetMask,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v2, v3, v4},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v5, v0},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v0, v5},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v2,
				Second: v1,
				Cycle:  []XY{v2, v1, v0, v5, v7, v6, v2},
				Label:  populatedMask,
			},
			{
				// f1
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v4, v0},
				Label:  populatedMask | inputAInSet,
			},
			{
				// f2
				First:  v2,
				Second: v6,
				Cycle:  []XY{v2, v6, v7, v5, v4, v3, v2},
				Label:  populatedMask | inputBInSet,
			},
			{
				// f3
				First:  v4,
				Second: v2,
				Cycle:  []XY{v4, v2, v3, v4},
				Label:  populatedMask | inputAInSet | inputBInSet,
			},
			{
				// f4
				First:  v0,
				Second: v4,
				Cycle:  []XY{v0, v4, v5, v0},
				Label:  populatedMask,
			},
		},
	})
}

func TestGraphOverlayInside(t *testing.T) {
	overlay := createOverlayFromWKTs(t,
		"POLYGON((0 0,3 0,3 3,0 3,0 0))",
		"POLYGON((1 1,2 1,2 2,1 2,1 1))",
	)

	/*
	  v3-----------------v2
	   |                 |
	   |                 |
	   |    v7-----v6    |
	   |     | f2  |     |
	   |     |     |     |
	   |    v4-----v5    |  f0
	   |  ,`             |
	   |,`     f1        |
	  v0-----------------v1

	*/

	v0 := XY{0, 0}
	v1 := XY{3, 0}
	v2 := XY{3, 3}
	v3 := XY{0, 3}
	v4 := XY{1, 1}
	v5 := XY{2, 1}
	v6 := XY{2, 2}
	v7 := XY{1, 2}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 2,
		NumEdges: 6,
		NumFaces: 3,
		Vertices: []VertexSpec{
			{
				Label:    populatedMask | inputAInSet,
				Vertices: []XY{v0},
			},
			{
				Label:    populatedMask | inSetMask,
				Vertices: []XY{v4},
			},
		},
		Edges: []EdgeSpec{
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v0, v3, v2, v1, v0},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v0, v1, v2, v3, v0},
			},
			{
				EdgeLabel: populatedMask | inSetMask,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v4, v7, v6, v5, v4},
			},
			{
				EdgeLabel: populatedMask | inSetMask,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v4, v5, v6, v7, v4},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: 0,
				Sequence:  []XY{v0, v4},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: 0,
				Sequence:  []XY{v4, v0},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v0,
				Second: v3,
				Cycle:  []XY{v0, v3, v2, v1, v0},
				Label:  populatedMask,
			},
			{
				// f1
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v3, v0, v4, v7, v6, v5, v4, v0},
				Label:  populatedMask | inputAInSet,
			},
			{
				// f2
				First:  v4,
				Second: v5,
				Cycle:  []XY{v4, v5, v6, v7, v4},
				Label:  populatedMask | inputAInSet | inputBInSet,
			},
		},
	})
}

func TestGraphOverlayReproduceHorizontalHoleLinkageBug(t *testing.T) {
	overlay := createOverlayFromWKTs(t,
		"MULTIPOLYGON(((4 0,4 1,5 1,5 0,4 0)),((1 0,1 2,3 2,3 0,1 0)))",
		"MULTIPOLYGON(((0 4,0 5,1 5,1 4,0 4)),((0 1,0 3,2 3,2 1,0 1)))",
	)

	/*
	  v16---v15
	   | f2  |
	   |     |
	  v13---v14
	   | `,
	   |f9 `,
	  v12---v19---v11
	   |  f4   `,f8|
	   |         `,|
	   |    v4----v18----v3
	   |     | f5  | `,f7|    f0
	   |     |     |   `,|
	  v9----v17---v10   v20   v8-----v7
	         |           | `,  | f1  |
	         |  f3       |f6 `,|     |
	   o    v1-----------v2---v5-----v6
	*/

	v1 := XY{1, 0}
	v2 := XY{3, 0}
	v3 := XY{3, 2}
	v4 := XY{1, 2}
	v5 := XY{4, 0}
	v6 := XY{5, 0}
	v7 := XY{5, 1}
	v8 := XY{4, 1}
	v9 := XY{0, 1}
	v10 := XY{2, 1}
	v11 := XY{2, 3}
	v12 := XY{0, 3}
	v13 := XY{0, 4}
	v14 := XY{1, 4}
	v15 := XY{1, 5}
	v16 := XY{0, 5}
	v17 := XY{1, 1}
	v18 := XY{2, 2}
	v19 := XY{1, 3}
	v20 := XY{3, 1}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 10,
		NumEdges: 36,
		NumFaces: 10,
		Vertices: []VertexSpec{
			{
				Label:    populatedMask | inputAInSet,
				Vertices: []XY{v1, v20, v2, v5},
			},
			{
				Label:    populatedMask | inSetMask,
				Vertices: []XY{v17, v18},
			},
			{
				Label:    populatedMask | inputBInSet,
				Vertices: []XY{v9, v19, v12, v13},
			},
		},
		Edges: []EdgeSpec{
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v20, v5},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v5, v2},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v5, v20},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v2, v5},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v12, v13},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v13, v19},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v13, v12},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v19, v13},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v5, v6, v7, v8, v5},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v5, v8, v7, v6, v5},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v13, v14, v15, v16, v13},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v13, v16, v15, v14, v13},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v18, v3, v20},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v20, v2},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v2, v1},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v1, v17},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v17, v9},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v9, v12},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v12, v19},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v19, v11, v18},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v17, v1},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v1, v2},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v2, v20},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v20, v3, v18},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v18, v11, v19},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v19, v12},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v12, v9},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v9, v17},
			},
			{
				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v18, v10, v17},
			},
			{
				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v17, v4, v18},
			},
			{
				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v17, v10, v18},
			},
			{
				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v18, v4, v17},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: 0,
				Sequence:  []XY{v20, v18},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: 0,
				Sequence:  []XY{v18, v20},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: 0,
				Sequence:  []XY{v19, v18},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: 0,
				Sequence:  []XY{v18, v19},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v19,
				Second: v11,
				Cycle: []XY{
					v19, v11, v18, v3, v20, v5, v8, v7,
					v6, v5, v2, v1, v17, v9, v12,
					v13, v16, v15, v14, v13, v19,
				},
				Label: inputBPopulated | inputAPopulated,
			},
			{
				// f1
				First:  v5,
				Second: v6,
				Cycle:  []XY{v5, v6, v7, v8, v5},
				Label:  inputBPopulated | inputAPopulated | inputAInSet,
			},
			{
				// f2
				First:  v13,
				Second: v14,
				Cycle:  []XY{v13, v14, v15, v16, v13},
				Label:  inputBPopulated | inputAPopulated | inputBInSet,
			},
			{
				// f3
				First:  v1,
				Second: v2,
				Cycle:  []XY{v1, v2, v20, v18, v10, v17, v1},
				Label:  inputBPopulated | inputAPopulated | inputAInSet,
			},
			{
				// f4
				First:  v17,
				Second: v4,
				Cycle:  []XY{v17, v4, v18, v19, v12, v9, v17},
				Label:  inputBPopulated | inputAPopulated | inputBInSet,
			},
			{
				// f5
				First:  v17,
				Second: v10,
				Cycle:  []XY{v17, v10, v18, v4, v17},
				Label:  inputBPopulated | inputAPopulated | inputBInSet | inputAInSet,
			},
			{
				// f6
				First:  v2,
				Second: v5,
				Cycle:  []XY{v2, v5, v20, v2},
				Label:  populatedMask,
			},
			{
				// f7
				First:  v20,
				Second: v3,
				Cycle:  []XY{v20, v3, v18, v20},
				Label:  inputBPopulated | inputAPopulated | inputAInSet,
			},
			{
				// f8
				First:  v18,
				Second: v11,
				Cycle:  []XY{v18, v11, v19, v18},
				Label:  inputBPopulated | inputAPopulated | inputBInSet,
			},
			{
				// f9
				First:  v12,
				Second: v19,
				Cycle:  []XY{v12, v19, v13, v12},
				Label:  populatedMask,
			},
		},
	})
}

func TestGraphOverlayFullyOverlappingEdge(t *testing.T) {
	overlay := createOverlayFromWKTs(t,
		"POLYGON((0 0,0 1,1 1,1 0,0 0))",
		"POLYGON((1 0,1 1,2 1,2 0,1 0))",
	)

	/*
	  v5-----v4----v3
	   |  f1 |  f2 |  f0
	   |     |     |
	  v0----v1-----v2
	*/

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{2, 0}
	v3 := XY{2, 1}
	v4 := XY{1, 1}
	v5 := XY{0, 1}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 3,
		NumEdges: 8,
		NumFaces: 3,
		Vertices: []VertexSpec{
			{
				Vertices: []XY{v0},
				Label:    populatedMask | inputAInSet,
			},
			{
				Vertices: []XY{v1, v4},
				Label:    populatedMask | inSetMask,
			},
		},
		Edges: []EdgeSpec{
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v1, v0},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v0, v5, v4},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v4, v5, v0},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v0, v1},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v4, v3, v2, v1},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v1, v2, v3, v4},
			},
			{
				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
				FaceLabel: populatedMask | inputAInSet,
				Sequence:  []XY{v1, v4},
			},
			{
				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
				FaceLabel: populatedMask | inputBInSet,
				Sequence:  []XY{v4, v1},
			},
		},
		Faces: []FaceSpec{
			{
				First:  v1,
				Second: v0,
				Cycle:  []XY{v0, v5, v4, v3, v2, v1, v0},
				Label:  inputAPopulated | inputBPopulated,
			},
			{
				First:  v1,
				Second: v4,
				Cycle:  []XY{v1, v4, v5, v0, v1},
				Label:  inputAPopulated | inputBPopulated | inputAInSet,
			},
			{
				First:  v1,
				Second: v2,
				Cycle:  []XY{v1, v2, v3, v4, v1},
				Label:  inputAPopulated | inputBPopulated | inputBInSet,
			},
		},
	})
}

func TestGraphOverlayPartiallyOverlappingEdge(t *testing.T) {
	overlay := createOverlayFromWKTs(t,
		"POLYGON((0 1,0 3,2 3,2 1,0 1))",
		"POLYGON((2 0,2 2,4 2,4 0,2 0))",
	)

	/*
	  v7-------v6    f0
	   |       |
	   | f1   v5-------v4
	   |       |       |
	  v0------v1   f2  |
	    `-, f3 |       |
	       `-,v2-------v3
	*/

	v0 := XY{0, 1}
	v1 := XY{2, 1}
	v2 := XY{2, 0}
	v3 := XY{4, 0}
	v4 := XY{4, 2}
	v5 := XY{2, 2}
	v6 := XY{2, 3}
	v7 := XY{0, 3}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 4,
		NumEdges: 12,
		NumFaces: 4,
		Vertices: []VertexSpec{
			{
				Vertices: []XY{v0},
				Label:    populatedMask | inputAInSet,
			},
			{
				Vertices: []XY{v2},
				Label:    populatedMask | inputBInSet,
			},
			{
				Vertices: []XY{v1, v5},
				Label:    populatedMask | inSetMask,
			},
		},
		Edges: []EdgeSpec{
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v1, v0},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
				Sequence:  []XY{v0, v7, v6, v5},
			},

			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v5, v6, v7, v0},
			},
			{
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated | inputAInSet,
				Sequence:  []XY{v0, v1},
			},

			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v5, v4, v3, v2},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
				Sequence:  []XY{v2, v1},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v1, v2},
			},
			{
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
				Sequence:  []XY{v2, v3, v4, v5},
			},
			{
				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
				FaceLabel: populatedMask | inputAInSet,
				Sequence:  []XY{v1, v5},
			},
			{
				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
				FaceLabel: populatedMask | inputBInSet,
				Sequence:  []XY{v5, v1},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v2, v0},
			},
			{
				EdgeLabel: populatedMask,
				FaceLabel: 0,
				Sequence:  []XY{v0, v2},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v0,
				Second: v7,
				Cycle:  []XY{v0, v7, v6, v5, v4, v3, v2, v0},
				Label:  populatedMask,
			},
			{
				// f1
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v5, v6, v7, v0},
				Label:  populatedMask | inputAInSet,
			},
			{
				// f2
				First:  v1,
				Second: v2,
				Cycle:  []XY{v1, v2, v3, v4, v5, v1},
				Label:  populatedMask | inputBInSet,
			},
			{
				// f3
				First:  v2,
				Second: v1,
				Cycle:  []XY{v2, v1, v0, v2},
				Label:  populatedMask,
			},
		},
	})
}

func TestGraphOverlayFullyOverlappingCycle(t *testing.T) {
	overlay := createOverlayFromWKTs(t,
		"POLYGON((0 0,0 1,1 1,1 0,0 0))",
		"POLYGON((0 0,0 1,1 1,1 0,0 0))",
	)

	/*
	  v3-------v2
	   |       |
	   |  f1   |  f0
	   |       |
	  v0-------v1
	*/

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{1, 1}
	v3 := XY{0, 1}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 1,
		NumEdges: 2,
		NumFaces: 2,
		Vertices: []VertexSpec{{
			Label:    populatedMask | inSetMask,
			Vertices: []XY{v0},
		}},
		Edges: []EdgeSpec{
			{
				EdgeLabel: populatedMask | inSetMask,
				FaceLabel: populatedMask | inSetMask,
				Sequence:  []XY{v0, v1, v2, v3, v0},
			},
			{
				EdgeLabel: populatedMask | inSetMask,
				FaceLabel: populatedMask,
				Sequence:  []XY{v0, v3, v2, v1, v0},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v0,
				Second: v3,
				Cycle:  []XY{v0, v3, v2, v1, v0},
				Label:  inputAPopulated | inputBPopulated,
			},
			{
				// f1
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v3, v0},
				Label:  inputAPopulated | inputBPopulated | inputAInSet | inputBInSet,
			},
		},
	})
}

func TestGraphOverlayTwoLineStringsIntersectingAtEndpoints(t *testing.T) {
	overlay := createOverlayFromWKTs(t,
		"LINESTRING(0 0,1 0)",
		"LINESTRING(0 0,0 1)",
	)

	/*
	  v0 B
	   |
	   |
	  v1----v2 A
	*/

	v0 := XY{0, 1}
	v1 := XY{0, 0}
	v2 := XY{1, 0}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 3,
		NumEdges: 4,
		NumFaces: 1,
		Vertices: []VertexSpec{
			{Vertices: []XY{v2}, Label: populatedMask | inputAInSet},
			{Vertices: []XY{v0}, Label: populatedMask | inputBInSet},
			{Vertices: []XY{v1}, Label: populatedMask | inSetMask},
		},
		Edges: []EdgeSpec{
			{
				Sequence:  []XY{v1, v2},
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
			},
			{
				Sequence:  []XY{v2, v1},
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
			},
			{
				Sequence:  []XY{v0, v1},
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
			},
			{
				Sequence:  []XY{v1, v0},
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
			},
		},
		Faces: []FaceSpec{{
			First:  v0,
			Second: v1,
			Cycle:  []XY{v0, v1, v2, v1, v0},
			Label:  populatedMask,
		}},
	})
}

func TestGraphOverlayReproduceFaceAllocationBug(t *testing.T) {
	overlay := createOverlayFromWKTs(t,
		"LINESTRING(0 1,1 0)",
		"MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))",
	)

	/*
	  v3------v2    v7------v6
	   |`, f2 |      |      |
	   |  `,  |  f0  |  f3  |
	   | f1 `,|      |      |
	  v0------v1----v4------v5
	*/

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{1, 1}
	v3 := XY{0, 1}
	v4 := XY{2, 0}
	v5 := XY{3, 0}
	v6 := XY{3, 1}
	v7 := XY{2, 1}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 4,
		NumEdges: 12,
		NumFaces: 4,
		Vertices: []VertexSpec{
			{
				Vertices: []XY{v1, v3},
				Label:    populatedMask | inSetMask,
			},
			{
				Vertices: []XY{v0, v4},
				Label:    populatedMask | inputBInSet,
			},
		},
		Edges: []EdgeSpec{
			{
				Sequence:  []XY{v1, v3},
				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
				FaceLabel: inputAPopulated,
			},
			{
				Sequence:  []XY{v3, v1},
				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
				FaceLabel: inputAPopulated,
			},
			{
				Sequence:  []XY{v0, v1},
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
			},
			{
				Sequence:  []XY{v1, v2, v3},
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
			},
			{
				Sequence:  []XY{v3, v0},
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
			},
			{
				Sequence:  []XY{v0, v3},
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
			},
			{
				Sequence:  []XY{v3, v2, v1},
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
			},
			{
				Sequence:  []XY{v1, v0},
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
			},
			{
				Sequence:  []XY{v4, v5, v6, v7, v4},
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
			},
			{
				Sequence:  []XY{v4, v7, v6, v5, v4},
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
			},
			{
				Sequence:  []XY{v4, v1},
				EdgeLabel: populatedMask,
				FaceLabel: 0,
			},
			{
				Sequence:  []XY{v1, v4},
				EdgeLabel: populatedMask,
				FaceLabel: 0,
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v1,
				Second: v0,
				Cycle:  []XY{v1, v0, v3, v2, v1, v4, v7, v6, v5, v4, v1},
				Label:  populatedMask,
			},
			{
				// f1
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v3, v0},
				Label:  populatedMask | inputBInSet,
			},
			{
				// f2
				First:  v1,
				Second: v2,
				Cycle:  []XY{v1, v2, v3, v1},
				Label:  populatedMask | inputBInSet,
			},
			{
				// f3
				First:  v4,
				Second: v5,
				Cycle:  []XY{v4, v5, v6, v7, v4},
				Label:  populatedMask | inputBInSet,
			},
		},
	})
}

func TestGraphOverlayReproducePointOnLineStringPrecisionBug(t *testing.T) {
	overlay := createOverlayFromWKTs(t,
		"LINESTRING(0 0,1 1)",
		"POINT(0.35355339059327373 0.35355339059327373)",
	)

	/*
	      v2
	      /
	    v1
	    /
	  v0
	*/

	v0 := XY{0, 0}
	v1 := XY{0.35355339059327373, 0.35355339059327373}
	v2 := XY{1, 1}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 3,
		NumEdges: 4,
		NumFaces: 1,
		Vertices: []VertexSpec{
			{
				Vertices: []XY{v0, v2},
				Label:    populatedMask | inputAInSet,
			},
			{
				Vertices: []XY{v1},
				Label:    populatedMask | inputAInSet | inputBInSet,
			},
		},
		Edges: []EdgeSpec{
			{
				Sequence:  []XY{v0, v1},
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
			},
			{
				Sequence:  []XY{v1, v2},
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
			},
			{
				Sequence:  []XY{v2, v1},
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
			},
			{
				Sequence:  []XY{v1, v0},
				EdgeLabel: populatedMask | inputAInSet,
				FaceLabel: inputAPopulated,
			},
		},
		Faces: []FaceSpec{
			{
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v1, v0},
				Label:  populatedMask,
			},
		},
	})
}

func TestGraphOverlayReproduceGhostOnGeometryBug(t *testing.T) {
	overlay := createOverlayFromWKTs(t,
		"LINESTRING(0 1,0 0,1 0)",
		"POLYGON((0 0,1 0,1 1,0 1,0 0.5,0 0))",
	)

	/*
	   v3        v2
	   @----------+
	   |          |
	   |          |
	 v4+    f1    |  f0
	   |          |
	   |          |
	   @----------@
	   v0        v1
	*/

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{1, 1}
	v3 := XY{0, 1}
	v4 := XY{0, 0.5}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 3,
		NumEdges: 6,
		NumFaces: 2,
		Vertices: []VertexSpec{
			{
				Vertices: []XY{v0, v1, v3},
				Label:    populatedMask | inputAInSet | inputBInSet,
			},
		},
		Edges: []EdgeSpec{
			{
				Sequence:  []XY{v0, v1},
				EdgeLabel: populatedMask | inSetMask,
				FaceLabel: populatedMask | inputBInSet,
			},
			{
				Sequence:  []XY{v1, v0},
				EdgeLabel: populatedMask | inSetMask,
				FaceLabel: populatedMask,
			},
			{
				Sequence:  []XY{v1, v2, v3},
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated | inputBInSet,
			},
			{
				Sequence:  []XY{v3, v2, v1},
				EdgeLabel: populatedMask | inputBInSet,
				FaceLabel: inputBPopulated,
			},
			{
				Sequence:  []XY{v3, v4, v0},
				EdgeLabel: populatedMask | inSetMask,
				FaceLabel: populatedMask | inputBInSet,
			},
			{
				Sequence:  []XY{v0, v4, v3},
				EdgeLabel: populatedMask | inSetMask,
				FaceLabel: populatedMask,
			},
		},
		Faces: []FaceSpec{
			{
				First:  v1,
				Second: v0,
				Cycle:  []XY{v1, v0, v4, v3, v2, v1},
				Label:  populatedMask,
			},
			{
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v3, v4, v0},
				Label:  populatedMask | inputBInSet,
			},
		},
	})
}
