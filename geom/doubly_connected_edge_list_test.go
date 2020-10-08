package geom

import (
	"fmt"
	"strconv"
	"testing"
)

type DCELSpec struct {
	NumVerts int
	NumEdges int
	NumFaces int
	Faces    []FaceSpec
	Edges    []EdgeLabelSpec
	Vertices []VertexSpec
}

type FaceSpec struct {
	// Origin and destination of an edge that is incident to the face.
	EdgeOrigin      XY
	EdgeDestin      XY
	OuterComponent  []XY
	InnerComponents [][]XY
	Label           uint8
}

type EdgeLabelSpec struct {
	Label uint8
	Edges []XY
}

type VertexSpec struct {
	Label    uint8
	Vertices []XY
}

func CheckDCEL(t *testing.T, dcel *doublyConnectedEdgeList, spec DCELSpec) {
	t.Helper()
	if spec.NumVerts != len(dcel.vertices) {
		t.Fatalf("verticies: want=%d got=%d", spec.NumVerts, len(dcel.vertices))
	}
	if spec.NumEdges != len(dcel.halfEdges) {
		t.Fatalf("edges: want=%d got=%d", spec.NumEdges, len(dcel.halfEdges))
	}
	if spec.NumFaces != len(dcel.faces) {
		t.Fatalf("faces: want=%d got=%d", spec.NumFaces, len(dcel.faces))
	}
	if spec.NumFaces != len(spec.Faces) {
		t.Fatalf("NumFaces doesn't match len(spec.Faces): %d vs %d", spec.NumFaces, len(spec.Faces))
	}

	for xy, vr := range dcel.vertices {
		if xy != vr.coords {
			t.Errorf("xy in vertex map doesn't match record")
		}
		if vr.incident == nil {
			t.Fatalf("vertex record (at %v) incident ptr not set", vr.coords)
		}
		if vr.incident.origin != vr {
			t.Errorf("incident edge of vert (at %v) doesn't have vert as its origin", vr.coords)
		}
	}

	for i, want := range spec.Faces {
		t.Run(fmt.Sprintf("f%d", i), func(t *testing.T) {
			got := findEdge(t, dcel, want.EdgeOrigin, want.EdgeDestin).incident

			if len(want.OuterComponent) == 0 {
				if got.outerComponent != nil {
					t.Fatal("want no outer component but outer component is not nil")
				}
			} else {
				if len(want.OuterComponent) != 0 && got.outerComponent == nil {
					t.Fatal("want outer component but outer component is nil")
				}
				CheckComponent(t, got, got.outerComponent, want.OuterComponent)
			}

			if len(got.innerComponents) != len(want.InnerComponents) {
				t.Fatalf("len want inners not equal to actual inners: %d vs %d",
					len(want.InnerComponents), len(got.innerComponents))
			}
			for i, wantInner := range want.InnerComponents {
				CheckComponent(t, got, got.innerComponents[i], wantInner)
			}
			if want.Label != got.label {
				t.Errorf("label doesn't match: want=%b got=%b", want.Label, got.label)
			}
		})
	}

	t.Run("edge_labels", func(t *testing.T) {
		for _, want := range spec.Edges {
			for i := 0; i+1 < len(want.Edges); i++ {
				u := want.Edges[i]
				v := want.Edges[i+1]
				e := findEdge(t, dcel, u, v)
				if e.label != want.Label {
					t.Errorf("incorrect label for edge %v -> %v: "+
						"want=%b got=%b", u, v, want.Label, e.label)
				}
				if e.twin.label != want.Label {
					t.Errorf("incorrect label for edge %v -> %v: "+
						"want=%b got=%b", v, u, want.Label, e.twin.label)
				}
			}
		}
	})

	t.Run("vertex_labels", func(t *testing.T) {
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
			}
		}
	})
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
		if i == 20 {
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

func TestGraphTriangle(t *testing.T) {
	poly, err := UnmarshalWKT("POLYGON((0 0,0 1,1 0,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	dcel := newDCELFromMultiPolygon(poly.AsPolygon().AsMultiPolygon(), inputAMask)

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
	  V0 *--------* V1

	*/

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{0, 1}

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 3,
		NumEdges: 6,
		NumFaces: 2,
		Faces: []FaceSpec{
			{
				// f0
				EdgeOrigin:      v2,
				EdgeDestin:      v1,
				OuterComponent:  nil,
				InnerComponents: [][]XY{{v2, v1, v0}},
				Label:           inputAPopulated,
			},
			{
				// f1
				EdgeOrigin:      v0,
				EdgeDestin:      v1,
				OuterComponent:  []XY{v0, v1, v2},
				InnerComponents: [][]XY{},
				Label:           inputAMask,
			},
		},
		Edges: []EdgeLabelSpec{{
			Label: inputAPopulated | inputAInSet,
			Edges: []XY{v0, v1, v2},
		}},
		Vertices: []VertexSpec{{
			Label:    inputAPopulated | inputAInSet,
			Vertices: []XY{v0, v1, v2},
		}},
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

	dcel := newDCELFromMultiPolygon(poly.AsPolygon().AsMultiPolygon(), inputBMask)

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
		NumVerts: 12,
		NumEdges: 24,
		NumFaces: 4,
		Faces: []FaceSpec{
			{
				// f0
				EdgeOrigin:      v3,
				EdgeDestin:      v2,
				OuterComponent:  nil,
				InnerComponents: [][]XY{{v3, v2, v1, v0}},
				Label:           inputBPopulated,
			},
			{
				// f1
				EdgeOrigin:      v2,
				EdgeDestin:      v3,
				OuterComponent:  []XY{v2, v3, v0, v1},
				InnerComponents: [][]XY{{v7, v4, v5, v6}, {v11, v8, v9, v10}},
				Label:           inputBPopulated | inputBInSet,
			},
			{
				// f2
				EdgeOrigin:      v4,
				EdgeDestin:      v7,
				OuterComponent:  []XY{v4, v7, v6, v5},
				InnerComponents: nil,
				Label:           inputBPopulated,
			},
			{
				// f3
				EdgeOrigin:      v8,
				EdgeDestin:      v11,
				OuterComponent:  []XY{v8, v11, v10, v9},
				InnerComponents: nil,
				Label:           inputBPopulated,
			},
		},
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

	dcel := newDCELFromMultiPolygon(mp.AsMultiPolygon(), inputBMask)

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{1, 1}
	v3 := XY{0, 1}
	v4 := XY{2, 0}
	v5 := XY{3, 0}
	v6 := XY{3, 1}
	v7 := XY{2, 1}

	CheckDCEL(t, dcel, DCELSpec{
		NumVerts: 8,
		NumEdges: 16,
		NumFaces: 3,
		Faces: []FaceSpec{
			{
				// f0
				EdgeOrigin:      v7,
				EdgeDestin:      v6,
				OuterComponent:  nil,
				InnerComponents: [][]XY{{v3, v2, v1, v0}, {v7, v6, v5, v4}},
				Label:           inputBPopulated,
			},
			{
				// f1
				EdgeOrigin:      v6,
				EdgeDestin:      v7,
				OuterComponent:  []XY{v6, v7, v4, v5},
				InnerComponents: nil,
				Label:           inputBPopulated | inputBInSet,
			},
			{
				// f2
				EdgeOrigin:      v2,
				EdgeDestin:      v3,
				OuterComponent:  []XY{v2, v3, v0, v1},
				InnerComponents: nil,
				Label:           inputBPopulated | inputBInSet,
			},
		},
	})
}

func TestGraphMultiLineString(t *testing.T) {
	mls, err := UnmarshalWKT("MULTILINESTRING((1 0,0 1,1 2),(2 0,3 1,2 2))")
	if err != nil {
		t.Fatal(err)
	}
	dcel := newDCELFromGeometry(mls, inputAMask)

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
		NumVerts: 6,
		NumEdges: 8,
		NumFaces: 1,
		Faces: []FaceSpec{{
			EdgeOrigin:      v0,
			EdgeDestin:      v1,
			OuterComponent:  nil,
			InnerComponents: [][]XY{{v0, v1, v2, v1}, {v5, v4, v3, v4}},
			Label:           inputAPopulated,
		}},
		Edges: []EdgeLabelSpec{
			{
				Label: inputAPopulated | inputAInSet,
				Edges: []XY{v0, v1, v2},
			},
			{
				Label: inputAPopulated | inputAInSet,
				Edges: []XY{v3, v4, v5},
			},
		},
		Vertices: []VertexSpec{{
			Label:    inputAPopulated | inputAInSet,
			Vertices: []XY{v0, v1, v2, v3, v4, v5},
		}},
	})
}

func TestGraphSelfOverlappingLineString(t *testing.T) {
	ls, err := UnmarshalWKT("LINESTRING(0 0,0 1,1 1,1 0,0 1,1 1,2 1)")
	if err != nil {
		t.Fatal(err)
	}
	dcel := newDCELFromGeometry(ls, inputAMask)

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
		NumVerts: 5,
		NumEdges: 10,
		NumFaces: 1, // just the infinite face
		Faces: []FaceSpec{{
			EdgeOrigin:      v0,
			EdgeDestin:      v1,
			OuterComponent:  nil,
			InnerComponents: [][]XY{{v0, v1, v2, v3, v1, v3, v2, v1}, {v2, v4}},
			Label:           inputAPopulated,
		}},
		Edges: []EdgeLabelSpec{
			{
				Label: inputAPopulated | inputAInSet,
				Edges: []XY{v0, v1, v3, v2},
			},
			{
				Label: inputAPopulated | inputAInSet,
				Edges: []XY{v1, v2, v4},
			},
		},
		Vertices: []VertexSpec{{
			Label:    inputAPopulated | inputAInSet,
			Vertices: []XY{v0, v1, v2, v3, v4},
		}},
	})
}

func TestGraphOverlayDisjoint(t *testing.T) {
	polyA, err := UnmarshalWKT("POLYGON((0 0,1 0,1 1,0 1,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	polyB, err := UnmarshalWKT("POLYGON((2 2,2 3,3 3,3 2,2 2))")
	if err != nil {
		t.Fatal(err)
	}

	overlay := createOverlay(polyA, polyB)

	/*
	                v7------v6
	                |        |
	                |   f2   |
	                |        |
	                |        |
	                v4------v5

	   v3------v2
	   |        |
	   |   f1   |       f0
	   |        |
	   |        |
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
		NumVerts: 8,
		NumEdges: 16,
		NumFaces: 3,
		Faces: []FaceSpec{
			{
				// f0
				EdgeOrigin:      v2,
				EdgeDestin:      v1,
				OuterComponent:  nil,
				InnerComponents: [][]XY{{v6, v5, v4, v7}, {v2, v1, v0, v3}},
				Label:           populatedMask,
			},
			{
				// f1
				EdgeOrigin:      v1,
				EdgeDestin:      v2,
				OuterComponent:  []XY{v2, v3, v0, v1},
				InnerComponents: nil,
				Label:           populatedMask | inputAInSet,
			},
			{
				// f2
				EdgeOrigin:      v5,
				EdgeDestin:      v6,
				OuterComponent:  []XY{v5, v6, v7, v4},
				InnerComponents: nil,
				Label:           populatedMask | inputBInSet,
			},
		},
	})
}

func TestGraphOverlayIntersecting(t *testing.T) {
	polyA, err := UnmarshalWKT("POLYGON((0 0,1 2,2 0,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	polyB, err := UnmarshalWKT("POLYGON((0 1,2 1,1 3,0 1))")
	if err != nil {
		t.Fatal(err)
	}

	overlay := createOverlay(polyA, polyB)

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
	       /        \
	      /    f1    \    f0
	     /            \
	    /              \
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
		NumVerts: 8,
		NumEdges: 20,
		NumFaces: 4,
		Faces: []FaceSpec{
			{
				// f0
				EdgeOrigin:      v7,
				EdgeDestin:      v6,
				OuterComponent:  nil,
				InnerComponents: [][]XY{{v7, v6, v2, v1, v0, v4, v5}},
				Label:           populatedMask,
			},
			{
				// f1
				EdgeOrigin:      v0,
				EdgeDestin:      v1,
				OuterComponent:  []XY{v0, v1, v2, v4},
				InnerComponents: nil,
				Label:           populatedMask | inputAInSet,
			},
			{
				// f2
				EdgeOrigin:      v6,
				EdgeDestin:      v7,
				OuterComponent:  []XY{v6, v7, v5, v4, v3, v2},
				InnerComponents: nil,
				Label:           populatedMask | inputBInSet,
			},
			{
				// f3
				EdgeOrigin:      v4,
				EdgeDestin:      v2,
				OuterComponent:  []XY{v4, v2, v3},
				InnerComponents: nil,
				Label:           populatedMask | inputAInSet | inputBInSet,
			},
		},
		Edges: []EdgeLabelSpec{
			{
				Label: populatedMask | inputAInSet,
				Edges: []XY{v4, v0, v1, v2},
			},
			{
				Label: populatedMask | inputBInSet,
				Edges: []XY{v4, v5, v7, v6, v2},
			},
			{
				Label: populatedMask | inSetMask,
				Edges: []XY{v4, v3, v2, v4},
			},
		},
		Vertices: []VertexSpec{
			{
				Label:    populatedMask | inputAInSet,
				Vertices: []XY{v0, v1},
			},
			{
				Label:    populatedMask | inputBInSet,
				Vertices: []XY{v5, v7, v6},
			},
			{
				Label:    populatedMask | inSetMask,
				Vertices: []XY{v2, v3, v4},
			},
		},
	})
}

func TestGraphOverlayInside(t *testing.T) {
	polyA, err := UnmarshalWKT("POLYGON((0 0,3 0,3 3,0 3,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	polyB, err := UnmarshalWKT("POLYGON((1 1,2 1,2 2,1 2,1 1))")
	if err != nil {
		t.Fatal(err)
	}

	overlay := createOverlay(polyA, polyB)

	/*
	  v3-----------------v2
	   |                 |
	   |                 |
	   |    v7-----v6    |
	   |     | f2  |     |
	   |     |     |     |
	   |    v4-----v5    |  f0
	   |                 |
	   |       f1        |
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
		NumVerts: 8,
		NumEdges: 16,
		NumFaces: 3,
		Faces: []FaceSpec{
			{
				// f0
				EdgeOrigin:      v2,
				EdgeDestin:      v1,
				OuterComponent:  nil,
				InnerComponents: [][]XY{{v2, v1, v0, v3}},
				Label:           populatedMask,
			},
			{
				// f1
				EdgeOrigin:      v0,
				EdgeDestin:      v1,
				OuterComponent:  []XY{v0, v1, v2, v3},
				InnerComponents: [][]XY{{v4, v7, v6, v5}},
				Label:           populatedMask | inputAInSet,
			},
			{
				// f2
				EdgeOrigin:      v4,
				EdgeDestin:      v5,
				OuterComponent:  []XY{v4, v5, v6, v7},
				InnerComponents: nil,
				Label:           populatedMask | inputAInSet | inputBInSet,
			},
		},
	})
}

func TestGraphOverlayReproduceHorizontalHoleLinkageBug(t *testing.T) {
	polyA, err := UnmarshalWKT("MULTIPOLYGON(((4 0,4 1,5 1,5 0,4 0)),((1 0,1 2,3 2,3 0,1 0)))")
	if err != nil {
		t.Fatal(err)
	}
	polyB, err := UnmarshalWKT("MULTIPOLYGON(((0 4,0 5,1 5,1 4,0 4)),((0 1,0 3,2 3,2 1,0 1)))")
	if err != nil {
		t.Fatal(err)
	}

	overlay := createOverlay(polyA, polyB)

	/*
	  v16---v15
	   | f2  |
	   |     |
	  v13---v14


	  v12---------v11
	   |  f4       |
	   |           |
	   |    v4----v18----v3
	   |     | f5  |     |    f0
	   |     |     |     |
	  v9----v17---v10    |    v8-----v7
	         |           |     | f1  |
	         |  f3       |     |     |
	   o    v1-----------v2   v5-----v6
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

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 18,
		NumEdges: 40,
		NumFaces: 6,
		Faces: []FaceSpec{
			{
				// f0
				EdgeOrigin:     v12,
				EdgeDestin:     v11,
				OuterComponent: nil,
				InnerComponents: [][]XY{
					{v14, v13, v16, v15},
					{v6, v5, v8, v7},
					{v1, v17, v9, v12, v11, v18, v3, v2},
				},
				Label: inputBPopulated | inputAPopulated,
			},
			{
				// f1
				EdgeOrigin:      v6,
				EdgeDestin:      v7,
				OuterComponent:  []XY{v6, v7, v8, v5},
				InnerComponents: nil,
				Label:           inputBPopulated | inputAPopulated | inputAInSet,
			},
			{
				// f2
				EdgeOrigin:      v13,
				EdgeDestin:      v14,
				OuterComponent:  []XY{v13, v14, v15, v16},
				InnerComponents: nil,
				Label:           inputBPopulated | inputAPopulated | inputBInSet,
			},
			{
				// f3
				EdgeOrigin:      v1,
				EdgeDestin:      v2,
				OuterComponent:  []XY{v1, v2, v3, v18, v10, v17},
				InnerComponents: nil,
				Label:           inputBPopulated | inputAPopulated | inputAInSet,
			},
			{
				// f4
				EdgeOrigin:      v4,
				EdgeDestin:      v18,
				OuterComponent:  []XY{v4, v18, v11, v12, v9, v17},
				InnerComponents: nil,
				Label:           inputBPopulated | inputAPopulated | inputBInSet,
			},
			{
				// f5
				EdgeOrigin:      v17,
				EdgeDestin:      v10,
				OuterComponent:  []XY{v17, v10, v18, v4},
				InnerComponents: nil,
				Label:           inputBPopulated | inputAPopulated | inputBInSet | inputAInSet,
			},
		},
	})
}

func TestGraphOverlayFullyOverlappingEdge(t *testing.T) {
	polyA, err := UnmarshalWKT("POLYGON((0 0,0 1,1 1,1 0,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	polyB, err := UnmarshalWKT("POLYGON((1 0,1 1,2 1,2 0,1 0))")
	if err != nil {
		t.Fatal(err)
	}

	overlay := createOverlay(polyA, polyB)

	/*
	  v5-----v4----v3
	   |  f2 |  f1 |  f0
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
		NumVerts: 6,
		NumEdges: 14,
		NumFaces: 3,
		Faces: []FaceSpec{
			{
				EdgeOrigin:      v1,
				EdgeDestin:      v0,
				OuterComponent:  nil,
				InnerComponents: [][]XY{{v0, v5, v4, v3, v2, v1}},
				Label:           inputAPopulated | inputBPopulated,
			},
			{
				EdgeOrigin:      v0,
				EdgeDestin:      v1,
				OuterComponent:  []XY{v0, v1, v4, v5},
				InnerComponents: nil,
				Label:           inputAPopulated | inputBPopulated | inputAInSet,
			},
			{
				EdgeOrigin:      v1,
				EdgeDestin:      v2,
				OuterComponent:  []XY{v1, v2, v3, v4},
				InnerComponents: nil,
				Label:           inputAPopulated | inputBPopulated | inputBInSet,
			},
		},
		Edges: []EdgeLabelSpec{
			{
				Label: populatedMask | inputAInSet,
				Edges: []XY{v1, v0, v5, v4},
			},
			{
				Label: populatedMask | inputBInSet,
				Edges: []XY{v4, v3, v2, v1},
			},
			{
				Label: populatedMask | inputAInSet | inputBInSet,
				Edges: []XY{v1, v4},
			},
		},
	})
}

func TestGraphOverlayPartiallyOverlappingEdge(t *testing.T) {
	polyA, err := UnmarshalWKT("POLYGON((0 1,0 3,2 3,2 1,0 1))")
	if err != nil {
		t.Fatal(err)
	}
	polyB, err := UnmarshalWKT("POLYGON((2 0,2 2,4 2,4 0,2 0))")
	if err != nil {
		t.Fatal(err)
	}

	overlay := createOverlay(polyA, polyB)

	/*
	  v7-------v6    f0
	   |       |
	   | f2   v5-------v4
	   |       |       |
	  v0------v1   f1  |
	           |       |
	          v2-------v3
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
		NumVerts: 8,
		NumEdges: 18,
		NumFaces: 3,
		Faces: []FaceSpec{
			{
				EdgeOrigin:      v1,
				EdgeDestin:      v0,
				OuterComponent:  nil,
				InnerComponents: [][]XY{{v1, v0, v7, v6, v5, v4, v3, v2}},
				Label:           inputAPopulated | inputBPopulated,
			},
			{
				EdgeOrigin:      v0,
				EdgeDestin:      v1,
				OuterComponent:  []XY{v0, v1, v5, v6, v7},
				InnerComponents: nil,
				Label:           inputAPopulated | inputBPopulated | inputAInSet,
			},
			{
				EdgeOrigin:      v1,
				EdgeDestin:      v2,
				OuterComponent:  []XY{v1, v2, v3, v4, v5},
				InnerComponents: nil,
				Label:           inputAPopulated | inputBPopulated | inputBInSet,
			},
		},
		Edges: []EdgeLabelSpec{
			{
				Label: populatedMask | inputAInSet,
				Edges: []XY{v1, v0, v7, v6, v5},
			},
			{
				Label: populatedMask | inputBInSet,
				Edges: []XY{v5, v4, v3, v2, v1},
			},
			{
				Label: populatedMask | inputAInSet | inputBInSet,
				Edges: []XY{v1, v5},
			},
		},
	})
}

func TestGraphOverlayFullyOverlappingCycle(t *testing.T) {
	polyA, err := UnmarshalWKT("POLYGON((0 0,0 1,1 1,1 0,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	polyB, err := UnmarshalWKT("POLYGON((0 0,0 1,1 1,1 0,0 0))")
	if err != nil {
		t.Fatal(err)
	}

	overlay := createOverlay(polyA, polyB)

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
		NumVerts: 4,
		NumEdges: 8,
		NumFaces: 2,
		Faces: []FaceSpec{
			{
				EdgeOrigin:      v1,
				EdgeDestin:      v0,
				OuterComponent:  nil,
				InnerComponents: [][]XY{{v1, v0, v3, v2}},
				Label:           inputAPopulated | inputBPopulated,
			},
			{
				EdgeOrigin:      v0,
				EdgeDestin:      v1,
				OuterComponent:  []XY{v0, v1, v2, v3},
				InnerComponents: nil,
				Label:           inputAPopulated | inputBPopulated | inputAInSet | inputBInSet,
			},
		},
		Edges: []EdgeLabelSpec{
			{
				Label: populatedMask | inputAInSet | inputBInSet,
				Edges: []XY{v0, v1, v2, v3},
			},
		},
	})
}

func TestGraphOverlayTwoLineStringsIntersectingAtEndpoints(t *testing.T) {
	lsA, err := UnmarshalWKT("LINESTRING(0 0,1 0)")
	if err != nil {
		t.Fatal(err)
	}
	lsB, err := UnmarshalWKT("LINESTRING(0 0,0 1)")
	if err != nil {
		t.Fatal(err)
	}

	overlay := createOverlay(lsA, lsB)

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
		Faces: []FaceSpec{{
			EdgeOrigin:      v0,
			EdgeDestin:      v1,
			OuterComponent:  nil,
			InnerComponents: [][]XY{{v0, v1, v2, v1}},
			Label:           populatedMask,
		}},
		Edges: []EdgeLabelSpec{
			{Edges: []XY{v1, v2}, Label: populatedMask | inputAInSet},
			{Edges: []XY{v0, v1}, Label: populatedMask | inputBInSet},
		},
		Vertices: []VertexSpec{
			{Vertices: []XY{v2}, Label: populatedMask | inputAInSet},
			{Vertices: []XY{v0}, Label: populatedMask | inputBInSet},
			{Vertices: []XY{v1}, Label: populatedMask | inSetMask},
		},
	})
}

func TestGraphOverlayReproduceFaceAllocationBug(t *testing.T) {
	geomA, err := UnmarshalWKT("LINESTRING(0 1,1 0)")
	if err != nil {
		t.Fatal(err)
	}
	geomB, err := UnmarshalWKT("MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 0,2 1,3 1,3 0,2 0)))")
	if err != nil {
		t.Fatal(err)
	}

	overlay := createOverlay(geomA, geomB)

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{1, 1}
	v3 := XY{0, 1}
	v4 := XY{2, 0}
	v5 := XY{3, 0}
	v6 := XY{3, 1}
	v7 := XY{2, 1}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 8,
		NumEdges: 18,
		NumFaces: 4,
		Faces: []FaceSpec{
			{
				EdgeOrigin:     v1,
				EdgeDestin:     v0,
				OuterComponent: nil,
				InnerComponents: [][]XY{
					{v4, v7, v6, v5},
					{v0, v3, v2, v1},
				},
				Label: populatedMask,
			},
			{
				EdgeOrigin:      v0,
				EdgeDestin:      v1,
				OuterComponent:  []XY{v0, v1, v3},
				InnerComponents: nil,
				Label:           populatedMask | inputBInSet,
			},
			{
				EdgeOrigin:      v1,
				EdgeDestin:      v2,
				OuterComponent:  []XY{v1, v2, v3},
				InnerComponents: nil,
				Label:           populatedMask | inputBInSet,
			},
			{
				EdgeOrigin:      v4,
				EdgeDestin:      v5,
				OuterComponent:  []XY{v4, v5, v6, v7},
				InnerComponents: nil,
				Label:           populatedMask | inputBInSet,
			},
		},
		Edges: []EdgeLabelSpec{
			{Edges: []XY{v1, v3}, Label: populatedMask | inputAInSet | inputBInSet},
			{Edges: []XY{v0, v1, v2, v3}, Label: populatedMask | inputBInSet},
			{Edges: []XY{v4, v5, v6, v7}, Label: populatedMask | inputBInSet},
		},
		Vertices: []VertexSpec{
			{Vertices: []XY{v1, v3}, Label: populatedMask | inputAInSet | inputBInSet},
			{Vertices: []XY{v0, v2, v4, v5, v6, v7}, Label: populatedMask | inputBInSet},
		},
	})
}

func TestGraphOverlayReproducePointOnLineStringPrecisionBug(t *testing.T) {
	geomA, err := UnmarshalWKT("LINESTRING(0 0,1 1)")
	if err != nil {
		t.Fatal(err)
	}
	geomB, err := UnmarshalWKT("POINT(0.35355339059327373 0.35355339059327373)")
	if err != nil {
		t.Fatal(err)
	}

	overlay := createOverlay(geomA, geomB)

	v0 := XY{0, 0}
	v1 := XY{0.35355339059327373, 0.35355339059327373}
	v2 := XY{1, 1}

	CheckDCEL(t, overlay, DCELSpec{
		NumVerts: 3,
		NumEdges: 4,
		NumFaces: 1,
		Faces: []FaceSpec{
			{
				EdgeOrigin:      v0,
				EdgeDestin:      v1,
				OuterComponent:  nil,
				InnerComponents: [][]XY{{v0, v1, v2, v1}},
				Label:           populatedMask,
			},
		},
		Edges: []EdgeLabelSpec{
			{Edges: []XY{v0, v1, v2}, Label: populatedMask | inputAInSet},
		},
		Vertices: []VertexSpec{
			{Vertices: []XY{v0, v2}, Label: populatedMask | inputAInSet},
			{Vertices: []XY{v1}, Label: populatedMask | inputAInSet | inputBInSet},
		},
	})
}

func TestRemoveDuplicateEdges(t *testing.T) {
	for i, tt := range []struct {
		input, output string
	}{
		{
			"MULTILINESTRING((0 0,1 1))",
			"MULTILINESTRING((0 0,1 1))",
		},
		{
			"MULTILINESTRING((0 0,1 1),(0 0,1 1))",
			"MULTILINESTRING((0 0,1 1))",
		},
		{
			"MULTILINESTRING((0 0,1 1),(1 1,0 0))",
			"MULTILINESTRING((0 0,1 1))",
		},
		{
			"MULTILINESTRING((0 0,0 1,1 1,2 1,2 0),(0 1,1 1,2 1))",
			"MULTILINESTRING((0 0,0 1,1 1,2 1,2 0))",
		},
		{
			"MULTILINESTRING((0 1,1 1,2 1),(0 0,0 1,1 1,2 1,2 0))",
			"MULTILINESTRING((0 1,1 1,2 1),(0 0,0 1),(2 1,2 0))",
		},
		{
			"MULTILINESTRING((0 0,0 1),(0 0,0 1,1 1))",
			"MULTILINESTRING((0 0,0 1),(0 1,1 1))",
		},
		{
			"MULTILINESTRING((1 1,0 1),(0 0,0 1,1 1))",
			"MULTILINESTRING((1 1,0 1),(0 0,0 1))",
		},
		{
			"MULTILINESTRING((0 0,0 1,1 1,1 0,0 1,1 1,2 1))",
			"MULTILINESTRING((0 0,0 1,1 1,1 0,0 1),(1 1,2 1))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			in, err := UnmarshalWKT(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			out, err := UnmarshalWKT(tt.output)
			if err != nil {
				t.Fatal(err)
			}
			got := removeDuplicateEdges(in.AsMultiLineString())
			if !out.EqualsExact(got.AsGeometry()) {
				t.Errorf(
					"\ninput: %v\nwant:  %v\ngot:   %v\n",
					in.AsText(), out.AsText(), got.AsText())
			}
		})
	}
}
