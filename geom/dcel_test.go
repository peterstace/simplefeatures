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
	EdgeOrigin XY
	EdgeDestin XY
	Cycle      []XY
	Label      uint8
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
				t.Fatalf("could not find edge spec matching sequence: %v", want.Sequence)
			}

			if e.edgeLabel != want.EdgeLabel {
				t.Errorf("incorrect edge label for edge with seq %v: "+
					"want=%b got=%b", want.Sequence, want.EdgeLabel, e.edgeLabel)
			}
			if e.faceLabel != want.FaceLabel {
				t.Errorf("incorrect face label for edge with seq %v: "+
					"want=%b got=%b", want.Sequence, want.FaceLabel, e.faceLabel)
			}
		}
	})

	for i, want := range spec.Faces {
		t.Run(fmt.Sprintf("face_%d", i), func(t *testing.T) {
			got := findEdge(t, dcel, want.EdgeOrigin, want.EdgeDestin).incident
			CheckComponent(t, got, got.cycle, want.Cycle)
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

func findEdge(t *testing.T, dcel *doublyConnectedEdgeList, origin, dest XY) *halfEdgeRecord {
	for _, e := range dcel.halfEdges {
		if e.origin.coords == origin && e.twin.origin.coords == dest {
			return e
		}
	}
	t.Fatalf("could not find edge with origin %v and dest %v", origin, dest)
	return nil
}

// TODO: rename to CheckCycle
func CheckComponent(t *testing.T, f *faceRecord, start *halfEdgeRecord, want []XY) {
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
		e = e.next
		if e == start {
			break
		}
	}
	CheckXYs(t, got, want)

	// Check component matches reverse order when following 'prev' pointer.
	var i int
	e = start
	got = nil
	for {
		i++
		if i == 100 {
			t.Fatal("inf loop")
		}

		if e.incident == nil {
			t.Errorf("half edge has no incident face set")
		} else if e.incident != f {
			t.Errorf("half edge has incorrect incident face")
		}
		if e.origin == nil {
			t.Errorf("edge origin not set")
		}
		got = append(got, e.origin.coords)
		e = e.prev
		if e == start {
			break
		}
	}
	for i := 0; i < len(got)/2; i++ {
		j := len(got) - i - 1
		got[i], got[j] = got[j], got[i]
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
		t.Errorf("XY sequences don't match: got=%v want=%v", got, want)
		return
	}
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
	t.Errorf("XY sequences don't match: got=%v want=%v", got, want)
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

//func TestGraphMultiLineString(t *testing.T) {
//	mls, err := UnmarshalWKT("MULTILINESTRING((1 0,0 1,1 2),(2 0,3 1,2 2))")
//	if err != nil {
//		t.Fatal(err)
//	}
//	dcel := newDCELFromGeometry(mls, MultiLineString{}, inputAMask, findInteractionPoints([]Geometry{mls}))
//
//	/*
//	        v2    v3
//	       /        \
//	      /          \
//	     /            \
//	   v1              v4
//	     \            /
//	      \          /
//	       \        /
//	        v0    v5
//	*/
//
//	v0 := XY{1, 0}
//	v1 := XY{0, 1}
//	v2 := XY{1, 2}
//	v3 := XY{2, 2}
//	v4 := XY{3, 1}
//	v5 := XY{2, 0}
//
//	CheckDCEL(t, dcel, DCELSpec{
//		NumVerts: 6,
//		NumEdges: 8,
//		NumFaces: 0,
//		Vertices: []VertexSpec{{
//			Label:    inputAPopulated | inputAInSet,
//			Vertices: []XY{v0, v1, v2, v3, v4, v5},
//		}},
//		Edges: []EdgeLabelSpec{
//			{
//				EdgeLabel: inputAPopulated | inputAInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v0, v1, v2, v1, v0},
//			},
//			{
//				EdgeLabel: inputAPopulated | inputAInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v3, v4, v5, v4, v3},
//			},
//		},
//		Faces: nil,
//	})
//}
//
//func TestGraphSelfOverlappingLineString(t *testing.T) {
//	ls, err := UnmarshalWKT("LINESTRING(0 0,0 1,1 1,1 0,0 1,1 1,2 1)")
//	if err != nil {
//		t.Fatal(err)
//	}
//	dcel := newDCELFromGeometry(ls, MultiLineString{}, inputAMask, findInteractionPoints([]Geometry{ls}))
//
//	/*
//	   v1----v2----v4
//	    |\   |
//	    | \  |
//	    |  \ |
//	    |   \|
//	   v0    v3
//	*/
//
//	v0 := XY{0, 0}
//	v1 := XY{0, 1}
//	v2 := XY{1, 1}
//	v3 := XY{1, 0}
//	v4 := XY{2, 1}
//
//	CheckDCEL(t, dcel, DCELSpec{
//		NumVerts: 5,
//		NumEdges: 10,
//		NumFaces: 0,
//		Vertices: []VertexSpec{{
//			Label:    inputAPopulated | inputAInSet,
//			Vertices: []XY{v0, v1, v2, v3, v4},
//		}},
//		Edges: []EdgeLabelSpec{
//			{
//				EdgeLabel: inputAPopulated | inputAInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v0, v1, v3, v2, v3, v1, v0},
//			},
//			{
//				EdgeLabel: inputAPopulated | inputAInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v1, v2, v4, v2, v1},
//			},
//		},
//		Faces: nil,
//	})
//}
//
//func TestGraphGhostDeduplication(t *testing.T) {
//	ls, err := UnmarshalWKT("LINESTRING(0 0,1 0)")
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// Ghost contains duplicated lines. This could happen in the scenario where
//	// one of the inputs is a multipolygon and the ghost joins the rings, but
//	// then the line joining the two input geometries is the same line segment.
//	ghost, err := UnmarshalWKT("MULTILINESTRING((0 0,0 1,0 0))")
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	dcel := newDCELFromGeometry(ls, ghost.AsMultiLineString(), inputAMask, findInteractionPoints([]Geometry{ls, ghost}))
//
//	v0 := XY{0, 0}
//	v1 := XY{1, 0}
//	v2 := XY{0, 1}
//
//	CheckDCEL(t, dcel, DCELSpec{
//		NumVerts: 3,
//		NumEdges: 4,
//		NumFaces: 0,
//		Vertices: []VertexSpec{
//			{Label: inputAPopulated | inputAInSet, Vertices: []XY{v0, v1}},
//			{Label: 0, Vertices: []XY{v2}},
//		},
//		Edges: []EdgeLabelSpec{
//			{
//				EdgeLabel: inputAPopulated | inputAInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v0, v1, v0},
//			},
//			{
//				EdgeLabel: inputAPopulated,
//				FaceLabel: 0,
//				Edges:     []XY{v0, v2, v0},
//			},
//		},
//	})
//}
//
//func TestGraphOverlayDisjoint(t *testing.T) {
//	overlay := createOverlayFromWKTs(t,
//		"POLYGON((0 0,1 0,1 1,0 1,0 0))",
//		"POLYGON((2 2,2 3,3 3,3 2,2 2))",
//	)
//
//	/*
//	                v7------v6
//	                |        |
//	                |   f3   |
//	                |        |
//	                |        |
//	               ,v4------v5
//	             ,`
//	   v3------v2
//	   |     ,` |
//	   |f2  `   |       f0
//	   |  ,` f1 |
//	   | ,      |
//	   v0------v1
//
//	*/
//
//	v0 := XY{0, 0}
//	v1 := XY{1, 0}
//	v2 := XY{1, 1}
//	v3 := XY{0, 1}
//	v4 := XY{2, 2}
//	v5 := XY{3, 2}
//	v6 := XY{3, 3}
//	v7 := XY{2, 3}
//
//	CheckDCEL(t, overlay, DCELSpec{
//		NumVerts: 8,
//		NumEdges: 20,
//		NumFaces: 4,
//		Vertices: []VertexSpec{
//			{
//				Label:    populatedMask | inputAInSet,
//				Vertices: []XY{v0, v1, v2, v3},
//			},
//			{
//				Label:    populatedMask | inputBInSet,
//				Vertices: []XY{v4, v5, v6, v7},
//			},
//		},
//		Edges: []EdgeLabelSpec{
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated | inputAInSet,
//				Edges:     []XY{v0, v1, v2, v3, v0},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated | inputBInSet,
//				Edges:     []XY{v4, v5, v6, v7, v4},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v0, v3, v2, v1, v0},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated,
//				Edges:     []XY{v4, v7, v6, v5, v4},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: 0,
//				Edges:     []XY{v0, v2, v0},
//			},
//			{
//				EdgeLabel: populatedMask,
//				FaceLabel: 0,
//				Edges:     []XY{v2, v4, v2},
//			},
//		},
//		Faces: []FaceSpec{
//			{
//				// f0
//				EdgeOrigin: v2,
//				EdgeDestin: v1,
//				Cycle:      []XY{v2, v1, v0, v3, v2, v4, v7, v6, v5, v4},
//				Label:      populatedMask,
//			},
//			{
//				// f1
//				EdgeOrigin: v1,
//				EdgeDestin: v2,
//				Cycle:      []XY{v2, v0, v1},
//				Label:      populatedMask | inputAInSet,
//			},
//			{
//				// f2
//				EdgeOrigin: v2,
//				EdgeDestin: v3,
//				Cycle:      []XY{v2, v3, v0},
//				Label:      populatedMask | inputAInSet,
//			},
//			{
//				// f3
//				EdgeOrigin: v5,
//				EdgeDestin: v6,
//				Cycle:      []XY{v5, v6, v7, v4},
//				Label:      populatedMask | inputBInSet,
//			},
//		},
//	})
//}
//
//func TestGraphOverlayIntersecting(t *testing.T) {
//	overlay := createOverlayFromWKTs(t,
//		"POLYGON((0 0,1 2,2 0,0 0))",
//		"POLYGON((0 1,2 1,1 3,0 1))",
//	)
//
//	/*
//	           v7
//	          /  \
//	         /    \
//	        /  f2  \
//	       /        \
//	      /    v3    \
//	     /    /  \    \
//	    /    / f3 \    \
//	   v5--v4------v2--v6
//	   |   /        \
//	   |f4/    f1    \    f0
//	   | /            \
//	   |/              \
//	   v0--------------v1
//
//	*/
//
//	v0 := XY{0, 0}
//	v1 := XY{2, 0}
//	v2 := XY{1.5, 1}
//	v3 := XY{1, 2}
//	v4 := XY{0.5, 1}
//	v5 := XY{0, 1}
//	v6 := XY{2, 1}
//	v7 := XY{1, 3}
//
//	CheckDCEL(t, overlay, DCELSpec{
//		NumVerts: 8,
//		NumEdges: 22,
//		NumFaces: 5,
//		Vertices: []VertexSpec{
//			{
//				Label:    populatedMask | inputAInSet,
//				Vertices: []XY{v0, v1},
//			},
//			{
//				Label:    populatedMask | inputBInSet,
//				Vertices: []XY{v5, v7, v6},
//			},
//			{
//				Label:    populatedMask | inSetMask,
//				Vertices: []XY{v2, v3, v4},
//			},
//		},
//		Edges: []EdgeLabelSpec{
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated | inputAInSet,
//				Edges:     []XY{v4, v0, v1, v2},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v2, v1, v0, v4},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated | inputBInSet,
//				Edges:     []XY{v2, v6, v7, v5, v4},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated,
//				Edges:     []XY{v4, v5, v7, v6, v2},
//			},
//			{
//				EdgeLabel: populatedMask | inSetMask,
//				FaceLabel: inputBPopulated,
//				Edges:     []XY{v2, v4},
//			},
//			{
//				EdgeLabel: populatedMask | inSetMask,
//				FaceLabel: inputBPopulated | inputBInSet,
//				Edges:     []XY{v4, v2},
//			},
//			{
//				EdgeLabel: populatedMask | inSetMask,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v4, v3, v2},
//			},
//			{
//				EdgeLabel: populatedMask | inSetMask,
//				FaceLabel: inputAPopulated | inputAInSet,
//				Edges:     []XY{v2, v3, v4},
//			},
//			{
//				EdgeLabel: populatedMask,
//				FaceLabel: 0,
//				Edges:     []XY{v0, v5, v0},
//			},
//		},
//		Faces: []FaceSpec{
//			{
//				// f0
//				EdgeOrigin: v7,
//				EdgeDestin: v6,
//				Cycle:      []XY{v7, v6, v2, v1, v0, v5},
//				Label:      populatedMask,
//			},
//			{
//				// f1
//				EdgeOrigin: v0,
//				EdgeDestin: v1,
//				Cycle:      []XY{v0, v1, v2, v4},
//				Label:      populatedMask | inputAInSet,
//			},
//			{
//				// f2
//				EdgeOrigin: v6,
//				EdgeDestin: v7,
//				Cycle:      []XY{v6, v7, v5, v4, v3, v2},
//				Label:      populatedMask | inputBInSet,
//			},
//			{
//				// f3
//				EdgeOrigin: v4,
//				EdgeDestin: v2,
//				Cycle:      []XY{v4, v2, v3},
//				Label:      populatedMask | inputAInSet | inputBInSet,
//			},
//			{
//				// f4
//				EdgeOrigin: v0,
//				EdgeDestin: v4,
//				Cycle:      []XY{v0, v4, v5},
//				Label:      populatedMask,
//			},
//		},
//	})
//}
//
//func TestGraphOverlayInside(t *testing.T) {
//	overlay := createOverlayFromWKTs(t,
//		"POLYGON((0 0,3 0,3 3,0 3,0 0))",
//		"POLYGON((1 1,2 1,2 2,1 2,1 1))",
//	)
//
//	/*
//	  v3-----------------v2
//	   |                 |
//	   |                 |
//	   |    v7-----v6    |
//	   |     | f2  |     |
//	   |     |     |     |
//	   |    v4-----v5    |  f0
//	   |  ,`             |
//	   |,`     f1        |
//	  v0-----------------v1
//
//	*/
//
//	v0 := XY{0, 0}
//	v1 := XY{3, 0}
//	v2 := XY{3, 3}
//	v3 := XY{0, 3}
//	v4 := XY{1, 1}
//	v5 := XY{2, 1}
//	v6 := XY{2, 2}
//	v7 := XY{1, 2}
//
//	CheckDCEL(t, overlay, DCELSpec{
//		NumVerts: 8,
//		NumEdges: 18,
//		NumFaces: 3,
//		Vertices: []VertexSpec{
//			{
//				Label:    populatedMask | inputAInSet,
//				Vertices: []XY{v0, v1, v2, v3},
//			},
//			{
//				Label:    populatedMask | inSetMask,
//				Vertices: []XY{v4, v5, v6, v7},
//			},
//		},
//		Edges: []EdgeLabelSpec{
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v0, v3, v2, v1, v0},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated | inputAInSet,
//				Edges:     []XY{v0, v1, v2, v3, v0},
//			},
//			{
//				EdgeLabel: populatedMask | inSetMask,
//				FaceLabel: inputBPopulated,
//				Edges:     []XY{v4, v7, v6, v5, v4},
//			},
//			{
//				EdgeLabel: populatedMask | inSetMask,
//				FaceLabel: inputBPopulated | inputBInSet,
//				Edges:     []XY{v4, v5, v6, v7, v4},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: 0,
//				Edges:     []XY{v0, v4, v0},
//			},
//		},
//		Faces: []FaceSpec{
//			{
//				// f0
//				EdgeOrigin: v2,
//				EdgeDestin: v1,
//				Cycle:      []XY{v2, v1, v0, v3},
//				Label:      populatedMask,
//			},
//			{
//				// f1
//				EdgeOrigin: v0,
//				EdgeDestin: v1,
//				Cycle:      []XY{v0, v1, v2, v3, v0, v4, v7, v6, v5, v4},
//				Label:      populatedMask | inputAInSet,
//			},
//			{
//				// f2
//				EdgeOrigin: v4,
//				EdgeDestin: v5,
//				Cycle:      []XY{v4, v5, v6, v7},
//				Label:      populatedMask | inputAInSet | inputBInSet,
//			},
//		},
//	})
//}
//
//func TestGraphOverlayReproduceHorizontalHoleLinkageBug(t *testing.T) {
//	overlay := createOverlayFromWKTs(t,
//		"MULTIPOLYGON(((4 0,4 1,5 1,5 0,4 0)),((1 0,1 2,3 2,3 0,1 0)))",
//		"MULTIPOLYGON(((0 4,0 5,1 5,1 4,0 4)),((0 1,0 3,2 3,2 1,0 1)))",
//	)
//
//	/*
//	  v16---v15
//	   | f2  |
//	   |     |
//	  v13---v14
//	   | `,
//	   |f9 `,
//	  v12---v19---v11
//	   |  f4   `,f8|
//	   |         `,|
//	   |    v4----v18----v3
//	   |     | f5  | `,f7|    f0
//	   |     |     |   `,|
//	  v9----v17---v10   v20   v8-----v7
//	         |           | `,  | f1  |
//	         |  f3       |f6 `,|     |
//	   o    v1-----------v2---v5-----v6
//	*/
//
//	v1 := XY{1, 0}
//	v2 := XY{3, 0}
//	v3 := XY{3, 2}
//	v4 := XY{1, 2}
//	v5 := XY{4, 0}
//	v6 := XY{5, 0}
//	v7 := XY{5, 1}
//	v8 := XY{4, 1}
//	v9 := XY{0, 1}
//	v10 := XY{2, 1}
//	v11 := XY{2, 3}
//	v12 := XY{0, 3}
//	v13 := XY{0, 4}
//	v14 := XY{1, 4}
//	v15 := XY{1, 5}
//	v16 := XY{0, 5}
//	v17 := XY{1, 1}
//	v18 := XY{2, 2}
//	v19 := XY{1, 3}
//	v20 := XY{3, 1}
//
//	CheckDCEL(t, overlay, DCELSpec{
//		NumVerts: 20,
//		NumEdges: 56,
//		NumFaces: 10,
//		Vertices: []VertexSpec{
//			{
//				Label:    populatedMask | inputAInSet,
//				Vertices: []XY{v1, v3, v20, v2, v8, v5, v7, v6},
//			},
//			{
//				Label:    populatedMask | inputBInSet,
//				Vertices: []XY{v9, v12, v19, v11, v13, v14, v16, v15},
//			},
//			{
//				Label:    populatedMask | inSetMask,
//				Vertices: []XY{v4, v17, v10, v18},
//			},
//		},
//		Edges: []EdgeLabelSpec{
//			{
//				EdgeLabel: populatedMask,
//				FaceLabel: 0,
//				Edges:     []XY{v20, v5, v2, v5, v20},
//			},
//			{
//				EdgeLabel: populatedMask,
//				FaceLabel: 0,
//				Edges:     []XY{v12, v13, v19, v13, v12},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated | inputAInSet,
//				Edges:     []XY{v5, v6, v7, v8, v5},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v5, v8, v7, v6, v5},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated | inputBInSet,
//				Edges:     []XY{v13, v14, v15, v16, v13},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated,
//				Edges:     []XY{v13, v16, v15, v14, v13},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v18, v3, v20, v2, v1, v17},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated,
//				Edges:     []XY{v17, v9, v12, v19, v11, v18},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated | inputAInSet,
//				Edges:     []XY{v17, v1, v2, v20, v3, v18},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated | inputBInSet,
//				Edges:     []XY{v18, v11, v19, v12, v9, v17},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
//				FaceLabel: inputBPopulated,
//				Edges:     []XY{v18, v10, v17},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v17, v4, v18},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
//				FaceLabel: inputBPopulated | inputBInSet,
//				Edges:     []XY{v17, v10, v18},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
//				FaceLabel: inputAPopulated | inputAInSet,
//				Edges:     []XY{v18, v4, v17},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: 0,
//				Edges:     []XY{v20, v18, v20},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: 0,
//				Edges:     []XY{v19, v18, v19},
//			},
//		},
//		Faces: []FaceSpec{
//			{
//				// f0
//				EdgeOrigin: v19,
//				EdgeDestin: v11,
//				Cycle: []XY{
//					v19, v11, v18, v3, v20, v5, v8, v7,
//					v6, v5, v2, v1, v17, v9, v12,
//					v13, v16, v15, v14, v13,
//				},
//				Label: inputBPopulated | inputAPopulated,
//			},
//			{
//				// f1
//				EdgeOrigin: v6,
//				EdgeDestin: v7,
//				Cycle:      []XY{v6, v7, v8, v5},
//				Label:      inputBPopulated | inputAPopulated | inputAInSet,
//			},
//			{
//				// f2
//				EdgeOrigin: v13,
//				EdgeDestin: v14,
//				Cycle:      []XY{v13, v14, v15, v16},
//				Label:      inputBPopulated | inputAPopulated | inputBInSet,
//			},
//			{
//				// f3
//				EdgeOrigin: v1,
//				EdgeDestin: v2,
//				Cycle:      []XY{v1, v2, v20, v18, v10, v17},
//				Label:      inputBPopulated | inputAPopulated | inputAInSet,
//			},
//			{
//				// f4
//				EdgeOrigin: v4,
//				EdgeDestin: v18,
//				Cycle:      []XY{v4, v18, v19, v12, v9, v17},
//				Label:      inputBPopulated | inputAPopulated | inputBInSet,
//			},
//			{
//				// f5
//				EdgeOrigin: v17,
//				EdgeDestin: v10,
//				Cycle:      []XY{v17, v10, v18, v4},
//				Label:      inputBPopulated | inputAPopulated | inputBInSet | inputAInSet,
//			},
//			{
//				// f6
//				EdgeOrigin: v2,
//				EdgeDestin: v5,
//				Cycle:      []XY{v2, v5, v20},
//				Label:      populatedMask,
//			},
//			{
//				// f7
//				EdgeOrigin: v20,
//				EdgeDestin: v3,
//				Cycle:      []XY{v20, v3, v18},
//				Label:      inputBPopulated | inputAPopulated | inputAInSet,
//			},
//			{
//				// f8
//				EdgeOrigin: v18,
//				EdgeDestin: v11,
//				Cycle:      []XY{v18, v11, v19},
//				Label:      inputBPopulated | inputAPopulated | inputBInSet,
//			},
//			{
//				// f9
//				EdgeOrigin: v12,
//				EdgeDestin: v19,
//				Cycle:      []XY{v12, v19, v13},
//				Label:      populatedMask,
//			},
//		},
//	})
//}
//
//func TestGraphOverlayFullyOverlappingEdge(t *testing.T) {
//	overlay := createOverlayFromWKTs(t,
//		"POLYGON((0 0,0 1,1 1,1 0,0 0))",
//		"POLYGON((1 0,1 1,2 1,2 0,1 0))",
//	)
//
//	/*
//	  v5-----v4----v3
//	   |  f2 |  f1 |  f0
//	   |     |     |
//	  v0----v1-----v2
//	*/
//
//	v0 := XY{0, 0}
//	v1 := XY{1, 0}
//	v2 := XY{2, 0}
//	v3 := XY{2, 1}
//	v4 := XY{1, 1}
//	v5 := XY{0, 1}
//
//	CheckDCEL(t, overlay, DCELSpec{
//		NumVerts: 6,
//		NumEdges: 14,
//		NumFaces: 3,
//		Vertices: []VertexSpec{
//			{Vertices: []XY{v0, v5}, Label: populatedMask | inputAInSet},
//			{Vertices: []XY{v1, v4}, Label: populatedMask | inputAInSet | inputBInSet},
//			{Vertices: []XY{v3, v2}, Label: populatedMask | inputBInSet},
//		},
//		Edges: []EdgeLabelSpec{
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v1, v0, v5, v4},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated | inputAInSet,
//				Edges:     []XY{v4, v5, v0, v1},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated,
//				Edges:     []XY{v4, v3, v2, v1},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated | inputBInSet,
//				Edges:     []XY{v1, v2, v3, v4},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
//				FaceLabel: populatedMask | inputAInSet,
//				Edges:     []XY{v1, v4},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
//				FaceLabel: populatedMask | inputBInSet,
//				Edges:     []XY{v4, v1},
//			},
//		},
//		Faces: []FaceSpec{
//			{
//				EdgeOrigin: v1,
//				EdgeDestin: v0,
//				Cycle:      []XY{v0, v5, v4, v3, v2, v1},
//				Label:      inputAPopulated | inputBPopulated,
//			},
//			{
//				EdgeOrigin: v0,
//				EdgeDestin: v1,
//				Cycle:      []XY{v0, v1, v4, v5},
//				Label:      inputAPopulated | inputBPopulated | inputAInSet,
//			},
//			{
//				EdgeOrigin: v1,
//				EdgeDestin: v2,
//				Cycle:      []XY{v1, v2, v3, v4},
//				Label:      inputAPopulated | inputBPopulated | inputBInSet,
//			},
//		},
//	})
//}
//
//func TestGraphOverlayPartiallyOverlappingEdge(t *testing.T) {
//	overlay := createOverlayFromWKTs(t,
//		"POLYGON((0 1,0 3,2 3,2 1,0 1))",
//		"POLYGON((2 0,2 2,4 2,4 0,2 0))",
//	)
//
//	/*
//	  v7-------v6    f0
//	   |       |
//	   | f2   v5-------v4
//	   |       |       |
//	  v0------v1   f1  |
//	    `-, f3 |       |
//	       `-,v2-------v3
//	*/
//
//	v0 := XY{0, 1}
//	v1 := XY{2, 1}
//	v2 := XY{2, 0}
//	v3 := XY{4, 0}
//	v4 := XY{4, 2}
//	v5 := XY{2, 2}
//	v6 := XY{2, 3}
//	v7 := XY{0, 3}
//
//	CheckDCEL(t, overlay, DCELSpec{
//		NumVerts: 8,
//		NumEdges: 20,
//		NumFaces: 4,
//		Vertices: []VertexSpec{
//			{Vertices: []XY{v0, v7, v6}, Label: populatedMask | inputAInSet},
//			{Vertices: []XY{v2, v3, v4}, Label: populatedMask | inputBInSet},
//			{Vertices: []XY{v1, v5}, Label: populatedMask | inputAInSet | inputBInSet},
//		},
//		Edges: []EdgeLabelSpec{
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated,
//				Edges:     []XY{v1, v0, v7, v6, v5},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated | inputAInSet,
//				Edges:     []XY{v5, v6, v7, v0, v1},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated,
//				Edges:     []XY{v5, v4, v3, v2, v1},
//			},
//			{
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated | inputBInSet,
//				Edges:     []XY{v1, v2, v3, v4, v5},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
//				FaceLabel: populatedMask | inputAInSet,
//				Edges:     []XY{v1, v5},
//			},
//			{
//				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
//				FaceLabel: populatedMask | inputBInSet,
//				Edges:     []XY{v5, v1},
//			},
//			{
//				EdgeLabel: populatedMask,
//				FaceLabel: 0,
//				Edges:     []XY{v2, v0, v2},
//			},
//		},
//		Faces: []FaceSpec{
//			{
//				// f0
//				EdgeOrigin: v0,
//				EdgeDestin: v7,
//				Cycle:      []XY{v0, v7, v6, v5, v4, v3, v2},
//				Label:      populatedMask,
//			},
//			{
//				// f1
//				EdgeOrigin: v0,
//				EdgeDestin: v1,
//				Cycle:      []XY{v0, v1, v5, v6, v7},
//				Label:      populatedMask | inputAInSet,
//			},
//			{
//				// f2
//				EdgeOrigin: v1,
//				EdgeDestin: v2,
//				Cycle:      []XY{v1, v2, v3, v4, v5},
//				Label:      populatedMask | inputBInSet,
//			},
//			{
//				// f3
//				EdgeOrigin: v2,
//				EdgeDestin: v1,
//				Cycle:      []XY{v2, v1, v0},
//				Label:      populatedMask,
//			},
//		},
//	})
//}
//
//func TestGraphOverlayFullyOverlappingCycle(t *testing.T) {
//	overlay := createOverlayFromWKTs(t,
//		"POLYGON((0 0,0 1,1 1,1 0,0 0))",
//		"POLYGON((0 0,0 1,1 1,1 0,0 0))",
//	)
//
//	/*
//	  v3-------v2
//	   |       |
//	   |  f1   |  f0
//	   |       |
//	  v0-------v1
//	*/
//
//	v0 := XY{0, 0}
//	v1 := XY{1, 0}
//	v2 := XY{1, 1}
//	v3 := XY{0, 1}
//
//	CheckDCEL(t, overlay, DCELSpec{
//		NumVerts: 4,
//		NumEdges: 8,
//		NumFaces: 2,
//		Vertices: []VertexSpec{{
//			Label:    populatedMask | inSetMask,
//			Vertices: []XY{v0, v1, v2, v3},
//		}},
//		Edges: []EdgeLabelSpec{
//			{
//				EdgeLabel: populatedMask | inSetMask,
//				FaceLabel: populatedMask | inSetMask,
//				Edges:     []XY{v0, v1, v2, v3, v0},
//			},
//			{
//				EdgeLabel: populatedMask | inSetMask,
//				FaceLabel: populatedMask,
//				Edges:     []XY{v0, v3, v2, v1, v0},
//			},
//		},
//		Faces: []FaceSpec{
//			{
//				// f0
//				EdgeOrigin: v1,
//				EdgeDestin: v0,
//				Cycle:      []XY{v1, v0, v3, v2},
//				Label:      inputAPopulated | inputBPopulated,
//			},
//			{
//				// f1
//				EdgeOrigin: v0,
//				EdgeDestin: v1,
//				Cycle:      []XY{v0, v1, v2, v3},
//				Label:      inputAPopulated | inputBPopulated | inputAInSet | inputBInSet,
//			},
//		},
//	})
//}
//
//func TestGraphOverlayTwoLineStringsIntersectingAtEndpoints(t *testing.T) {
//	overlay := createOverlayFromWKTs(t,
//		"LINESTRING(0 0,1 0)",
//		"LINESTRING(0 0,0 1)",
//	)
//
//	/*
//	  v0 B
//	   |
//	   |
//	  v1----v2 A
//	*/
//
//	v0 := XY{0, 1}
//	v1 := XY{0, 0}
//	v2 := XY{1, 0}
//
//	CheckDCEL(t, overlay, DCELSpec{
//		NumVerts: 3,
//		NumEdges: 4,
//		NumFaces: 1,
//		Vertices: []VertexSpec{
//			{Vertices: []XY{v2}, Label: populatedMask | inputAInSet},
//			{Vertices: []XY{v0}, Label: populatedMask | inputBInSet},
//			{Vertices: []XY{v1}, Label: populatedMask | inSetMask},
//		},
//		Edges: []EdgeLabelSpec{
//			{
//				Edges:     []XY{v1, v2, v1},
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated,
//			},
//			{
//				Edges:     []XY{v0, v1, v0},
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated,
//			},
//		},
//		Faces: []FaceSpec{{
//			EdgeOrigin: v0,
//			EdgeDestin: v1,
//			Cycle:      []XY{v0, v1, v2, v1},
//			Label:      populatedMask,
//		}},
//	})
//}
//
//func TestGraphOverlayReproduceFaceAllocationBug(t *testing.T) {
//	overlay := createOverlayFromWKTs(t,
//		"LINESTRING(0 1,1 0)",
//		"MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))",
//	)
//
//	/*
//	  v3------v2    v7------v6
//	   |`, f2 |      |      |
//	   |  `,  |  f0  |  f3  |
//	   | f1 `,|      |      |
//	  v0------v1----v4------v5
//	*/
//
//	v0 := XY{0, 0}
//	v1 := XY{1, 0}
//	v2 := XY{1, 1}
//	v3 := XY{0, 1}
//	v4 := XY{2, 0}
//	v5 := XY{3, 0}
//	v6 := XY{3, 1}
//	v7 := XY{2, 1}
//
//	CheckDCEL(t, overlay, DCELSpec{
//		NumVerts: 8,
//		NumEdges: 20,
//		NumFaces: 4,
//		Vertices: []VertexSpec{
//			{Vertices: []XY{v1, v3}, Label: populatedMask | inputAInSet | inputBInSet},
//			{Vertices: []XY{v0, v2, v4, v5, v6, v7}, Label: populatedMask | inputBInSet},
//		},
//		Edges: []EdgeLabelSpec{
//			{
//				Edges:     []XY{v1, v3, v1},
//				EdgeLabel: populatedMask | inputAInSet | inputBInSet,
//				FaceLabel: inputAPopulated,
//			},
//			{
//				Edges:     []XY{v0, v1, v2, v3, v0},
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated | inputBInSet,
//			},
//			{
//				Edges:     []XY{v0, v3, v2, v1, v0},
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated,
//			},
//			{
//				Edges:     []XY{v4, v5, v6, v7, v4},
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated | inputBInSet,
//			},
//			{
//				Edges:     []XY{v4, v7, v6, v5, v4},
//				EdgeLabel: populatedMask | inputBInSet,
//				FaceLabel: inputBPopulated,
//			},
//			{
//				Edges:     []XY{v1, v4, v1},
//				EdgeLabel: populatedMask,
//				FaceLabel: 0,
//			},
//		},
//		Faces: []FaceSpec{
//			{
//				// f0
//				EdgeOrigin: v1,
//				EdgeDestin: v0,
//				Cycle:      []XY{v1, v0, v3, v2, v1, v4, v7, v6, v5, v4},
//				Label:      populatedMask,
//			},
//			{
//				// f1
//				EdgeOrigin: v0,
//				EdgeDestin: v1,
//				Cycle:      []XY{v0, v1, v3},
//				Label:      populatedMask | inputBInSet,
//			},
//			{
//				// f2
//				EdgeOrigin: v1,
//				EdgeDestin: v2,
//				Cycle:      []XY{v1, v2, v3},
//				Label:      populatedMask | inputBInSet,
//			},
//			{
//				// f3
//				EdgeOrigin: v4,
//				EdgeDestin: v5,
//				Cycle:      []XY{v4, v5, v6, v7},
//				Label:      populatedMask | inputBInSet,
//			},
//		},
//	})
//}
//
//func TestGraphOverlayReproducePointOnLineStringPrecisionBug(t *testing.T) {
//	overlay := createOverlayFromWKTs(t,
//		"LINESTRING(0 0,1 1)",
//		"POINT(0.35355339059327373 0.35355339059327373)",
//	)
//
//	/*
//	      v2
//	      /
//	    v1
//	    /
//	  v0
//	*/
//
//	v0 := XY{0, 0}
//	v1 := XY{0.35355339059327373, 0.35355339059327373}
//	v2 := XY{1, 1}
//
//	CheckDCEL(t, overlay, DCELSpec{
//		NumVerts: 3,
//		NumEdges: 4,
//		NumFaces: 1,
//		Vertices: []VertexSpec{
//			{Vertices: []XY{v0, v2}, Label: populatedMask | inputAInSet},
//			{Vertices: []XY{v1}, Label: populatedMask | inputAInSet | inputBInSet},
//		},
//		Edges: []EdgeLabelSpec{
//			{
//				Edges:     []XY{v0, v1, v2, v1, v0},
//				EdgeLabel: populatedMask | inputAInSet,
//				FaceLabel: inputAPopulated,
//			},
//		},
//		Faces: []FaceSpec{
//			{
//				EdgeOrigin: v0,
//				EdgeDestin: v1,
//				Cycle:      []XY{v0, v1, v2, v1},
//				Label:      populatedMask,
//			},
//		},
//	})
//}
