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
	InSet  [2]bool
}

type EdgeSpec struct {
	SrcEdge  [2]bool
	SrcFace  [2]bool
	InSet    [2]bool
	Sequence []XY
}

type VertexSpec struct {
	Src      [2]bool
	InSet    [2]bool
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
				if want.Src != vert.src {
					t.Errorf("%v: src mismatch, want:%v got:%v",
						wantXY, want.Src, vert.src)
				}
				if want.InSet != vert.inSet {
					t.Errorf("%v: inSet mismatch, want:%v got:%v",
						wantXY, want.InSet, vert.inSet)
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
			if !reflect.DeepEqual(bruteForceEdgeSet, vr.incidents) {
				t.Fatalf("vertex record at %v doesn't have correct incidents: "+
					"bruteForceEdgeSet=%v incidentsSet=%v", vr.coords, bruteForceEdgeSet, vr.incidents)
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
				if !xysEqual(sequenceToXYs(e.seq), seq) {
					continue
				}
				want = spec.Edges[i]
				found = true
			}
			if !found {
				t.Fatalf("could not find edge spec matching edge: %v", e)
			}

			if want.SrcEdge != e.srcEdge {
				t.Errorf("%v: srcEdge mismatch, want:%v got:%v",
					want.Sequence, want.SrcEdge, e.srcEdge)
			}
			if want.SrcFace != e.srcFace {
				t.Errorf("%v: srcFace mismatch, want:%v got:%v",
					want.Sequence, want.SrcFace, e.srcFace)
			}
			if want.InSet != e.inSet {
				t.Errorf("%v: inSet mismatch, want:%v got:%v",
					want.Sequence, want.InSet, e.inSet)
			}
		}
	})

	for i, want := range spec.Faces {
		t.Run(fmt.Sprintf("face_%d", i), func(t *testing.T) {
			var got *faceRecord
			if len(spec.Faces) == 1 {
				got = dcel.faces[0]
			} else {
				got = findEdge(t, dcel, want.First, want.Second).incident
			}
			CheckCycle(t, got, got.cycle, want.Cycle)
			if want.InSet != got.inSet {
				t.Errorf("%v: inSet mismatch, want:%v got:%v",
					want.Cycle, want.InSet, got.inSet)
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
		if e.seq.GetXY(0) == first && e.seq.GetXY(1) == second {
			return e
		}
	}
	t.Fatalf("could not find edge with first %v and second %v", first, second)
	return nil
}

func CheckCycle(t *testing.T, f *faceRecord, start *halfEdgeRecord, want []XY) {
	if start == nil {
		if len(want) != 0 {
			t.Errorf("start is nil but want non-empty cycle: %v", want)
		}
		return
	}

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
		got = append(got, sequenceToXYs(e.seq)[:e.seq.Length()-1]...)
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

		got = append(got, sequenceToXYs(e.seq.Reverse())[:e.seq.Length()-1]...)

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

func newDCELFromWKTs(t *testing.T, wktA, wktB string) *doublyConnectedEdgeList {
	gA, err := UnmarshalWKT(wktA)
	if err != nil {
		t.Fatal(err)
	}
	gB, err := UnmarshalWKT(wktB)
	if err != nil {
		t.Fatal(err)
	}
	return newDCELFromGeometries(gA, gB)
}

func newDCELFromWKT(t *testing.T, wkt string) *doublyConnectedEdgeList {
	return newDCELFromWKTs(t, wkt, "GEOMETRYCOLLECTION EMPTY")
}

func TestDCELTriangle(t *testing.T) {
	dcel := newDCELFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))")

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
		NumFaces: 2,
		Vertices: []VertexSpec{{
			Src:      [2]bool{true},
			InSet:    [2]bool{true},
			Vertices: []XY{v0},
		}},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v0, v1, v2, v0},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v0, v2, v1, v0},
			},
		},
		Faces: []FaceSpec{
			{
				First:  v0,
				Second: v2,
				Cycle:  []XY{v0, v2, v1, v0},
				InSet:  [2]bool{false},
			},
			{
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v0},
				InSet:  [2]bool{true},
			},
		},
	})
}

func TestDCELWithHoles(t *testing.T) {
	dcel := newDCELFromWKT(t, "POLYGON((0 0,5 0,5 5,0 5,0 0),(1 1,2 1,2 2,1 2,1 1),(3 3,4 3,4 4,3 4,3 3))")

	/*
	         f0

	  v3-------------------v2
	   |                    |
	   |           v9---v10 |
	   |    f1      |f3 |   |
	   |           v8---v11 |
	   |          /         |
	   |  v5----v6          |
	   |   |  ,` |          |
	   |   |,`   |          |
	   |  v4----v7          |
	   |,`                  |
	  v0-------------------v1

	*/

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
		NumVerts: 4,
		NumEdges: 14,
		NumFaces: 5,
		Vertices: []VertexSpec{{
			Src:      [2]bool{true},
			InSet:    [2]bool{true},
			Vertices: []XY{v0, v4, v6, v8},
		}},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v0, v1, v2, v3, v0},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v0, v3, v2, v1, v0},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v4, v5, v6},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v6, v5, v4},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v6, v7, v4},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v4, v7, v6},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v8, v9, v10, v11, v8},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v8, v11, v10, v9, v8},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v6, v8},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v8, v6},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{false},
				Sequence: []XY{v4, v6},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{false},
				Sequence: []XY{v6, v4},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v0, v4},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v4, v0},
			},
		},
		Faces: []FaceSpec{
			{
				First:  v0,
				Second: v3,
				Cycle:  []XY{v0, v3, v2, v1, v0},
				InSet:  [2]bool{false},
			},
			{
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v3, v0, v4, v5, v6, v8, v9, v10, v11, v8, v6, v7, v4, v0},
				InSet:  [2]bool{true},
			},
			{
				First:  v4,
				Second: v7,
				Cycle:  []XY{v4, v7, v6, v4},
				InSet:  [2]bool{false},
			},
			{
				First:  v6,
				Second: v5,
				Cycle:  []XY{v6, v5, v4, v6},
				InSet:  [2]bool{false},
			},
			{
				First:  v8,
				Second: v11,
				Cycle:  []XY{v8, v11, v10, v9, v8},
				InSet:  [2]bool{false},
			},
		},
	})
}

func TestDCELWithMultiPolygon(t *testing.T) {
	dcel := newDCELFromWKT(t, "MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))")

	/*
	            f0
	  v3-----v2   v7-----v6
	   | f1  |     | f2  |
	   |     |     |     |
	  v0-----v1---v4-----v5
	*/

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{1, 1}
	v3 := XY{0, 1}
	v4 := XY{2, 0}
	v5 := XY{3, 0}
	v6 := XY{3, 1}
	v7 := XY{2, 1}

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 3,
		NumEdges: 8,
		NumFaces: 3,
		Vertices: []VertexSpec{{
			Src:      [2]bool{true},
			InSet:    [2]bool{true},
			Vertices: []XY{v0, v1, v4},
		}},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v0, v1},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v1, v0},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v1, v2, v3, v0},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v0, v3, v2, v1},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v4, v5, v6, v7, v4},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v4, v7, v6, v5, v4},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{false},
				Sequence: []XY{v1, v4},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{false},
				Sequence: []XY{v4, v1},
			},
		},
		Faces: []FaceSpec{
			{
				First:  v1,
				Second: v4,
				Cycle:  []XY{v3, v2, v1, v4, v7, v6, v5, v4, v1, v0, v3},
				InSet:  [2]bool{false},
			},
			{
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v3, v0},
				InSet:  [2]bool{true},
			},
			{
				First:  v4,
				Second: v5,
				Cycle:  []XY{v4, v5, v6, v7, v4},
				InSet:  [2]bool{true},
			},
		},
	})
}

func TestDCELMultiLineString(t *testing.T) {
	dcel := newDCELFromWKT(t, "MULTILINESTRING((1 0,0 1,1 2),(2 0,3 1,2 2))")

	/*
	        v2    v3
	       /        \
	      /          \
	     /            \
	   v1              v4
	     \            /
	      \          /
	       \        /
	        v0....v5
	*/

	v0 := XY{1, 0}
	v1 := XY{0, 1}
	v2 := XY{1, 2}
	v3 := XY{2, 2}
	v4 := XY{3, 1}
	v5 := XY{2, 0}

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 4,
		NumEdges: 6,
		NumFaces: 1,
		Vertices: []VertexSpec{{
			Src:      [2]bool{true},
			InSet:    [2]bool{true},
			Vertices: []XY{v0, v2, v3, v5},
		}},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v0, v1, v2},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v2, v1, v0},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v3, v4, v5},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v5, v4, v3},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{false},
				Sequence: []XY{v0, v5},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{false},
				Sequence: []XY{v5, v0},
			},
		},
		Faces: []FaceSpec{
			{
				First:  v5,
				Second: v4,
				Cycle:  []XY{v5, v4, v3, v4, v5, v0, v1, v2, v1, v0, v5},
				InSet:  [2]bool{false},
			},
		},
	})
}

func TestDCELSelfOverlappingLineString(t *testing.T) {
	dcel := newDCELFromWKT(t, "LINESTRING(0 0,0 1,1 1,1 0,0 1,1 1,2 1)")

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
		NumFaces: 2,
		Vertices: []VertexSpec{{
			Src:      [2]bool{true},
			InSet:    [2]bool{true},
			Vertices: []XY{v0, v1, v2, v4},
		}},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v0, v1},
			},
			{
				SrcEdge:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v1, v0},
			},
			{
				SrcEdge:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v1, v2},
			},
			{
				SrcEdge:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v2, v1},
			},
			{
				SrcEdge:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v1, v3, v2},
			},
			{
				SrcEdge:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v2, v3, v1},
			},
			{
				SrcEdge:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v2, v4},
			},
			{
				SrcEdge:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v4, v2},
			},
		},
		Faces: []FaceSpec{
			{
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v4, v2, v3, v1, v0},
			},
			{
				First:  v1,
				Second: v3,
				Cycle:  []XY{v3, v2, v1, v3},
			},
		},
	})
}

func TestDCELDisjoint(t *testing.T) {
	dcel := newDCELFromWKTs(t,
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

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 3,
		NumEdges: 10,
		NumFaces: 4,
		Vertices: []VertexSpec{
			{
				Src:      [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Vertices: []XY{v0, v2},
			},
			{
				Src:      [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Vertices: []XY{v4},
			},
		},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v0, v1, v2},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v2, v1, v0},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v2, v3, v0},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v0, v3, v2},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v4, v5, v6, v7, v4},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v4, v7, v6, v5, v4},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v0, v2},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v2, v0},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
				Sequence: []XY{v2, v4},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
				Sequence: []XY{v4, v2},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v2,
				Second: v1,
				Cycle:  []XY{v2, v1, v0, v3, v2, v4, v7, v6, v5, v4, v2},
				InSet:  [2]bool{false, false},
			},
			{
				// f1
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v0},
				InSet:  [2]bool{true, false},
			},
			{
				// f2
				First:  v2,
				Second: v3,
				Cycle:  []XY{v2, v3, v0, v2},
				InSet:  [2]bool{true, false},
			},
			{
				// f3
				First:  v4,
				Second: v5,
				Cycle:  []XY{v4, v5, v6, v7, v4},
				InSet:  [2]bool{false, true},
			},
		},
	})
}

func TestDCELIntersecting(t *testing.T) {
	dcel := newDCELFromWKTs(t,
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

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 4,
		NumEdges: 14,
		NumFaces: 5,
		Vertices: []VertexSpec{
			{
				Src:      [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Vertices: []XY{v0},
			},
			{
				Src:      [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Vertices: []XY{v5},
			},
			{
				Src:      [2]bool{true, true},
				InSet:    [2]bool{true, true},
				Vertices: []XY{v2, v4},
			},
		},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v4, v0},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v0, v1, v2},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v2, v1, v0},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v0, v4},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v2, v6, v7, v5},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v5, v4},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v4, v5},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v5, v7, v6, v2},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v2, v4},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v4, v2},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v4, v3, v2},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v2, v3, v4},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
				Sequence: []XY{v5, v0},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
				Sequence: []XY{v0, v5},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v2,
				Second: v1,
				Cycle:  []XY{v2, v1, v0, v5, v7, v6, v2},
				InSet:  [2]bool{false, false},
			},
			{
				// f1
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v4, v0},
				InSet:  [2]bool{true, false},
			},
			{
				// f2
				First:  v2,
				Second: v6,
				Cycle:  []XY{v2, v6, v7, v5, v4, v3, v2},
				InSet:  [2]bool{false, true},
			},
			{
				// f3
				First:  v4,
				Second: v2,
				Cycle:  []XY{v4, v2, v3, v4},
				InSet:  [2]bool{true, true},
			},
			{
				// f4
				First:  v0,
				Second: v4,
				Cycle:  []XY{v0, v4, v5, v0},
				InSet:  [2]bool{false, false},
			},
		},
	})
}

func TestDCELInside(t *testing.T) {
	dcel := newDCELFromWKTs(t,
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

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 2,
		NumEdges: 6,
		NumFaces: 3,
		Vertices: []VertexSpec{
			{
				Src:      [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Vertices: []XY{v0},
			},
			{
				Src:      [2]bool{false, true},
				InSet:    [2]bool{true, true},
				Vertices: []XY{v4},
			},
		},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v0, v3, v2, v1, v0},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v0, v1, v2, v3, v0},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v4, v7, v6, v5, v4},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v4, v5, v6, v7, v4},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v0, v4},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v4, v0},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v0,
				Second: v3,
				Cycle:  []XY{v0, v3, v2, v1, v0},
				InSet:  [2]bool{false, false},
			},
			{
				// f1
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v3, v0, v4, v7, v6, v5, v4, v0},
				InSet:  [2]bool{true, false},
			},
			{
				// f2
				First:  v4,
				Second: v5,
				Cycle:  []XY{v4, v5, v6, v7, v4},
				InSet:  [2]bool{true, true},
			},
		},
	})
}

func TestDCELReproduceHorizontalHoleLinkageBug(t *testing.T) {
	dcel := newDCELFromWKTs(t,
		"MULTIPOLYGON(((4 0,4 1,5 1,5 0,4 0)),((1 0,1 2,3 2,3 0,1 0)))",
		"MULTIPOLYGON(((0 4,0 5,1 5,1 4,0 4)),((0 1,0 3,2 3,2 1,0 1)))",
	)

	/*
	  v16---v15
	   | f2  |
	   |     |
	  v13---v14
	   |
	   |
	  v12---------v11    f0
	   |  f4       |
	   |           |
	   |    v4----v18----v3
	   |     | f5  |     |
	   |     |     |     |
	  v9----v17---v10    |    v8-----v7
	   `, f6 |           |     | f1  |
	     `,  |  f3       |     |     |
	   o   `v1-----------v2---v5-----v6
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

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 8,
		NumEdges: 26,
		NumFaces: 7,
		Vertices: []VertexSpec{
			{
				Src:      [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Vertices: []XY{v1, v2, v5},
			},
			{
				Src:      [2]bool{true, true},
				InSet:    [2]bool{true, true},
				Vertices: []XY{v17, v18},
			},
			{
				Src:      [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Vertices: []XY{v9, v12, v13},
			},
		},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
				Sequence: []XY{v5, v2},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
				Sequence: []XY{v2, v5},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
				Sequence: []XY{v12, v13},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
				Sequence: []XY{v13, v12},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v5, v6, v7, v8, v5},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v5, v8, v7, v6, v5},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v13, v14, v15, v16, v13},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v13, v16, v15, v14, v13},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v2, v1},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v1, v17},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v17, v9},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v9, v12},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v17, v1},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v1, v2},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v12, v9},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v9, v17},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v18, v10, v17},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v17, v4, v18},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v17, v10, v18},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v18, v4, v17},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
				Sequence: []XY{v1, v9},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
				Sequence: []XY{v9, v1},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v18, v3, v2},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v2, v3, v18},
			},

			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v18, v11, v12},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v12, v11, v18},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v12,
				Second: v11,
				Cycle: []XY{
					v12, v11, v18, v3, v2, v5, v8, v7,
					v6, v5, v2, v1, v9, v12,
					v13, v16, v15, v14, v13, v12,
				},
				InSet: [2]bool{false, false},
			},
			{
				// f1
				First:  v5,
				Second: v6,
				Cycle:  []XY{v5, v6, v7, v8, v5},
				InSet:  [2]bool{true, false},
			},
			{
				// f2
				First:  v13,
				Second: v14,
				Cycle:  []XY{v13, v14, v15, v16, v13},
				InSet:  [2]bool{false, true},
			},
			{
				// f3
				First:  v1,
				Second: v2,
				Cycle:  []XY{v1, v2, v3, v18, v10, v17, v1},
				InSet:  [2]bool{true, false},
			},
			{
				// f4
				First:  v17,
				Second: v4,
				Cycle:  []XY{v17, v4, v18, v11, v12, v9, v17},
				InSet:  [2]bool{false, true},
			},
			{
				// f5
				First:  v17,
				Second: v10,
				Cycle:  []XY{v17, v10, v18, v4, v17},
				InSet:  [2]bool{true, true},
			},
			{
				// f6
				First:  v1,
				Second: v17,
				Cycle:  []XY{v1, v17, v9, v1},
				InSet:  [2]bool{false, false},
			},
		},
	})
}

func TestDCELFullyOverlappingEdge(t *testing.T) {
	dcel := newDCELFromWKTs(t,
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

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 3,
		NumEdges: 8,
		NumFaces: 3,
		Vertices: []VertexSpec{
			{
				Vertices: []XY{v0},
				Src:      [2]bool{true, false},
				InSet:    [2]bool{true, false},
			},
			{
				Vertices: []XY{v1, v4},
				Src:      [2]bool{true, true},
				InSet:    [2]bool{true, true},
			},
		},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v1, v0},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v0, v5, v4},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v4, v5, v0},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v0, v1},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v4, v3, v2, v1},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v1, v2, v3, v4},
			},
			{
				SrcEdge:  [2]bool{true, true},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v1, v4},
			},
			{
				SrcEdge:  [2]bool{true, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v4, v1},
			},
		},
		Faces: []FaceSpec{
			{
				First:  v1,
				Second: v0,
				Cycle:  []XY{v0, v5, v4, v3, v2, v1, v0},
				InSet:  [2]bool{false, false},
			},
			{
				First:  v1,
				Second: v4,
				Cycle:  []XY{v1, v4, v5, v0, v1},
				InSet:  [2]bool{true, false},
			},
			{
				First:  v1,
				Second: v2,
				Cycle:  []XY{v1, v2, v3, v4, v1},
				InSet:  [2]bool{false, true},
			},
		},
	})
}

func TestDCELPartiallyOverlappingEdge(t *testing.T) {
	dcel := newDCELFromWKTs(t,
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

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 4,
		NumEdges: 12,
		NumFaces: 4,
		Vertices: []VertexSpec{
			{
				Vertices: []XY{v0},
				Src:      [2]bool{true, false},
				InSet:    [2]bool{true, false},
			},
			{
				Vertices: []XY{v2},
				Src:      [2]bool{false, true},
				InSet:    [2]bool{false, true},
			},
			{
				Vertices: []XY{v1, v5},
				Src:      [2]bool{true, true},
				InSet:    [2]bool{true, true},
			},
		},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v1, v0},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v0, v7, v6, v5},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v5, v6, v7, v0},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v0, v1},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v5, v4, v3, v2},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v2, v1},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v1, v2},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v2, v3, v4, v5},
			},
			{
				SrcEdge:  [2]bool{true, true},
				SrcFace:  [2]bool{true, false},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v1, v5},
			},
			{
				SrcEdge:  [2]bool{true, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v5, v1},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
				Sequence: []XY{v2, v0},
			},
			{
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
				Sequence: []XY{v0, v2},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v0,
				Second: v7,
				Cycle:  []XY{v0, v7, v6, v5, v4, v3, v2, v0},
				InSet:  [2]bool{false, false},
			},
			{
				// f1
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v5, v6, v7, v0},
				InSet:  [2]bool{true, false},
			},
			{
				// f2
				First:  v1,
				Second: v2,
				Cycle:  []XY{v1, v2, v3, v4, v5, v1},
				InSet:  [2]bool{false, true},
			},
			{
				// f3
				First:  v2,
				Second: v1,
				Cycle:  []XY{v2, v1, v0, v2},
				InSet:  [2]bool{false, false},
			},
		},
	})
}

func TestDCELFullyOverlappingCycle(t *testing.T) {
	dcel := newDCELFromWKTs(t,
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

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 1,
		NumEdges: 2,
		NumFaces: 2,
		Vertices: []VertexSpec{{
			Src:      [2]bool{true, true},
			InSet:    [2]bool{true, true},
			Vertices: []XY{v0},
		}},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{true, true},
				SrcFace:  [2]bool{true, true},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v0, v1, v2, v3, v0},
			},
			{
				SrcEdge:  [2]bool{true, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, true},
				Sequence: []XY{v0, v3, v2, v1, v0},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v0,
				Second: v3,
				Cycle:  []XY{v0, v3, v2, v1, v0},
				InSet:  [2]bool{false, false},
			},
			{
				// f1
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v3, v0},
				InSet:  [2]bool{true, true},
			},
		},
	})
}

func TestDCELTwoLineStringsIntersectingAtEndpoints(t *testing.T) {
	dcel := newDCELFromWKTs(t,
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

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 3,
		NumEdges: 4,
		NumFaces: 1,
		Vertices: []VertexSpec{
			{
				Vertices: []XY{v2},
				Src:      [2]bool{true, false},
				InSet:    [2]bool{true, false},
			},
			{
				Vertices: []XY{v0},
				Src:      [2]bool{false, true},
				InSet:    [2]bool{false, true},
			},
			{
				Vertices: []XY{v1},
				Src:      [2]bool{true, true},
				InSet:    [2]bool{true, true},
			},
		},
		Edges: []EdgeSpec{
			{
				Sequence: []XY{v1, v2},
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
			},
			{
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
				Sequence: []XY{v2, v1},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v0, v1},
			},
			{
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
				Sequence: []XY{v1, v0},
			},
		},
		Faces: []FaceSpec{{
			First:  v0,
			Second: v1,
			Cycle:  []XY{v0, v1, v2, v1, v0},
			InSet:  [2]bool{false, false},
		}},
	})
}

func TestDCELReproduceFaceAllocationBug(t *testing.T) {
	dcel := newDCELFromWKTs(t,
		"LINESTRING(0 1,1 0)",
		"MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))",
	)

	/*
	  v3------v2    v7------v6
	   |`, f2 |      |      |
	   |\ `,  |  f0  |  f4  |
	   | \  `,|      |      |
	   |  \f1 v8     |      |
	   |   \  | `,   |      |
	   | f3 \ |   `, |      |
	   |     \|     `|      |
	  v0------v1    v4------v5
	*/

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{1, 1}
	v3 := XY{0, 1}
	v4 := XY{2, 0}
	v5 := XY{3, 0}
	v6 := XY{3, 1}
	v7 := XY{2, 1}
	v8 := XY{1, 0.5}

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 5,
		NumEdges: 16,
		NumFaces: 5,
		Vertices: []VertexSpec{
			{
				Vertices: []XY{v1, v3},
				Src:      [2]bool{true, true},
				InSet:    [2]bool{true, true},
			},
			{
				Vertices: []XY{v0, v4, v8},
				Src:      [2]bool{false, true},
				InSet:    [2]bool{false, true},
			},
		},
		Edges: []EdgeSpec{
			{
				Sequence: []XY{v1, v3},
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, true},
			},
			{
				Sequence: []XY{v3, v1},
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, true},
			},
			{
				Sequence: []XY{v0, v1},
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
			},
			{
				Sequence: []XY{v1, v0},
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
			},
			{
				Sequence: []XY{v3, v0},
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
			},
			{
				Sequence: []XY{v0, v3},
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
			},
			{
				Sequence: []XY{v4, v5, v6, v7, v4},
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
			},
			{
				Sequence: []XY{v4, v7, v6, v5, v4},
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
			},

			{
				Sequence: []XY{v1, v8},
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
			},
			{
				Sequence: []XY{v8, v1},
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
			},

			{
				Sequence: []XY{v4, v8},
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
			},
			{
				Sequence: []XY{v8, v4},
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, false},
			},

			{
				Sequence: []XY{v3, v8},
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
			},
			{
				Sequence: []XY{v8, v3},
				SrcEdge:  [2]bool{false, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
			},
			{
				Sequence: []XY{v8, v2, v3},
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
			},
			{
				Sequence: []XY{v3, v2, v8},
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
			},
		},
		Faces: []FaceSpec{
			{
				// f0
				First:  v1,
				Second: v0,
				Cycle:  []XY{v1, v0, v3, v2, v8, v4, v7, v6, v5, v4, v8, v1},
				InSet:  [2]bool{false, false},
			},
			{
				// f1
				First:  v1,
				Second: v8,
				Cycle:  []XY{v1, v8, v3, v1},
				InSet:  [2]bool{false, true},
			},
			{
				// f2
				First:  v8,
				Second: v2,
				Cycle:  []XY{v8, v2, v3, v8},
				InSet:  [2]bool{false, true},
			},
			{
				// f3
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v3, v0},
				InSet:  [2]bool{false, true},
			},
			{
				// f4
				First:  v4,
				Second: v5,
				Cycle:  []XY{v4, v5, v6, v7, v4},
				InSet:  [2]bool{false, true},
			},
		},
	})
}

func TestDCELReproducePointOnLineStringPrecisionBug(t *testing.T) {
	dcel := newDCELFromWKTs(t,
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

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 3,
		NumEdges: 4,
		NumFaces: 1,
		Vertices: []VertexSpec{
			{
				Vertices: []XY{v0, v2},
				Src:      [2]bool{true, false},
				InSet:    [2]bool{true, false},
			},
			{
				Vertices: []XY{v1},
				Src:      [2]bool{true, true},
				InSet:    [2]bool{true, true},
			},
		},
		Edges: []EdgeSpec{
			{
				Sequence: []XY{v0, v1},
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
			},
			{
				Sequence: []XY{v1, v2},
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
			},
			{
				Sequence: []XY{v2, v1},
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
			},
			{
				Sequence: []XY{v1, v0},
				SrcEdge:  [2]bool{true, false},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, false},
			},
		},
		Faces: []FaceSpec{
			{
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v1, v0},
				InSet:  [2]bool{false, false},
			},
		},
	})
}

func TestDCELReproduceGhostOnGeometryBug(t *testing.T) {
	dcel := newDCELFromWKTs(t,
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

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 3,
		NumEdges: 6,
		NumFaces: 2,
		Vertices: []VertexSpec{
			{
				Vertices: []XY{v0, v1, v3},
				Src:      [2]bool{true, true},
				InSet:    [2]bool{true, true},
			},
		},
		Edges: []EdgeSpec{
			{
				Sequence: []XY{v0, v1},
				SrcEdge:  [2]bool{true, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{true, true},
			},
			{
				Sequence: []XY{v1, v0},
				SrcEdge:  [2]bool{true, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, true},
			},
			{
				Sequence: []XY{v1, v2, v3},
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{false, true},
			},
			{
				Sequence: []XY{v3, v2, v1},
				SrcEdge:  [2]bool{false, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{false, true},
			},
			{
				Sequence: []XY{v3, v4, v0},
				SrcEdge:  [2]bool{true, true},
				SrcFace:  [2]bool{false, true},
				InSet:    [2]bool{true, true},
			},
			{
				Sequence: []XY{v0, v4, v3},
				SrcEdge:  [2]bool{true, true},
				SrcFace:  [2]bool{false, false},
				InSet:    [2]bool{true, true},
			},
		},
		Faces: []FaceSpec{
			{
				First:  v1,
				Second: v0,
				Cycle:  []XY{v1, v0, v4, v3, v2, v1},
				InSet:  [2]bool{false, false},
			},
			{
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v3, v4, v0},
				InSet:  [2]bool{false, true},
			},
		},
	})
}

func TestDECLWithEmptyGeometryCollection(t *testing.T) {
	dcel := newDCELFromWKT(t, "GEOMETRYCOLLECTION EMPTY")
	CheckDCEL(t, dcel, DCELSpec{
		NumFaces: 1,
		Faces: []FaceSpec{{
			Cycle: nil, // No cycle
			InSet: [2]bool{false},
		}},
	})
}

func TestDCELWithGeometryCollection(t *testing.T) {
	dcel := newDCELFromWKT(t, `GEOMETRYCOLLECTION(
 		POINT(0 0),
 		LINESTRING(0 1,1 1),
 		POLYGON((2 0,3 0,3 1,2 1,2 0))
 	)`)

	/*
	  v1---v2   v6----v5
	  | `-,     |      |
	  |    `-,  |      |
	  v0      `-v3----v4
	*/

	v0 := XY{0, 0}
	v1 := XY{0, 1}
	v2 := XY{1, 1}
	v3 := XY{2, 0}
	v4 := XY{3, 0}
	v5 := XY{3, 1}
	v6 := XY{2, 1}

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 4,
		NumEdges: 8,
		NumFaces: 2,
		Vertices: []VertexSpec{
			{
				Src:      [2]bool{true},
				InSet:    [2]bool{true},
				Vertices: []XY{v0, v1, v2, v3},
			},
		},
		Edges: []EdgeSpec{
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v1, v2},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v2, v1},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{true},
				InSet:    [2]bool{true},
				Sequence: []XY{v3, v4, v5, v6, v3},
			},
			{
				SrcEdge:  [2]bool{true},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{true},
				Sequence: []XY{v3, v6, v5, v4, v3},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{false},
				Sequence: []XY{v0, v1},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{false},
				Sequence: []XY{v1, v0},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{false},
				Sequence: []XY{v1, v3},
			},
			{
				SrcEdge:  [2]bool{false},
				SrcFace:  [2]bool{false},
				InSet:    [2]bool{false},
				Sequence: []XY{v3, v1},
			},
		},
		Faces: []FaceSpec{
			{
				First:  v0,
				Second: v1,
				Cycle:  []XY{v0, v1, v2, v1, v3, v6, v5, v4, v3, v1, v0},
				InSet:  [2]bool{false},
			},
			{
				First:  v3,
				Second: v4,
				Cycle:  []XY{v3, v4, v5, v6, v3},
				InSet:  [2]bool{true},
			},
		},
	})
}
