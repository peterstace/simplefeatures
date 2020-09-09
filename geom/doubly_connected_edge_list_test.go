package geom

import (
	"fmt"
	"testing"
)

type DCELSpec struct {
	NumVerts int
	NumEdges int
	NumFaces int
	Faces    []FaceSpec
}

type FaceSpec struct {
	// Origin and destination of an edge that is incident to the face.
	EdgeOrigin      XY
	EdgeDestin      XY
	OuterComponent  []XY
	InnerComponents [][]XY
	Label           uint8
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
	for i, face := range spec.Faces {
		t.Run(fmt.Sprintf("f%d", i), func(t *testing.T) {
			f := findEdge(t, dcel, face.EdgeOrigin, face.EdgeDestin).incident

			if len(face.OuterComponent) == 0 {
				if f.outerComponent != nil {
					t.Fatal("want no outer component but outer component is not nil")
				}
			} else {
				if len(face.OuterComponent) != 0 && f.outerComponent == nil {
					t.Fatal("want outer component but outer component is nil")
				}
				CheckComponent(t, f, f.outerComponent, face.OuterComponent)
			}

			if len(f.innerComponents) != len(face.InnerComponents) {
				t.Errorf("len want inners not equal to actual inners: %d vs %d",
					len(face.InnerComponents), len(f.innerComponents))
				return
			}
			for i, wantInner := range face.InnerComponents {
				CheckComponent(t, f, f.innerComponents[i], wantInner)
			}
			if face.Label != f.label {
				t.Errorf("label doesn't match: want=%b got=%b", face.Label, f.label)
			}
		})
	}
}

func CheckVertexIncidents(t *testing.T, verts map[XY]*vertexRecord) {
	t.Helper()
	for xy, vr := range verts {
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

// TODO: This function is a hack/duplication of CheckComponent that can be used
// before we sort out faces in the overlay algorithm.
func CheckHalfEdgeLoop(t *testing.T, start *halfEdgeRecord, want []XY) {
	// Check component matches forward order when following 'next' pointer.
	e := start
	var got []XY
	for {
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

func CheckFaceComponents(
	t *testing.T, f *faceRecord, wantOuter []XY, wantInners ...[]XY,
) {
	t.Helper()

	if len(wantOuter) == 0 {
		if f.outerComponent != nil {
			t.Fatal("want no outer component but outer component is not nil")
		}
	} else {
		if len(wantOuter) != 0 && f.outerComponent == nil {
			t.Fatal("want outer component but outer component is nil")
		}
		CheckComponent(t, f, f.outerComponent, wantOuter)
	}

	if len(f.innerComponents) != len(wantInners) {
		t.Errorf("len want inners not equal to actual inners: %d vs %d",
			len(wantInners), len(f.innerComponents))
		return
	}
	for i, wantInner := range wantInners {
		CheckComponent(t, f, f.innerComponents[i], wantInner)
	}
}

func CheckComponent(t *testing.T, f *faceRecord, start *halfEdgeRecord, want []XY) {
	// Check component matches forward order when following 'next' pointer.
	e := start
	var got []XY
	for {
		if e.incident != f {
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

		if e.incident != f {
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
				Label:           inputAPresent,
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
	})
}

func TestGraphWithHoles(t *testing.T) {
	poly, err := UnmarshalWKT("POLYGON((0 0,5 0,5 5,0 5,0 0),(1 1,2 1,2 2,1 2,1 1),(3 3,4 3,4 4,3 4,3 3))")
	if err != nil {
		t.Fatal(err)
	}

	/*
		                             F0
		V3                                                        V2

		 .  ------------e5------------------------------------->  .
		    <-----------e4--------------------------------------
		 ^ |                                                    ^ |
		 | |                             V9             V10     | |
		 | |                              .  ---e18--->  .      | |
		 | |                                 <--e19----         | |
		 | |                              ^ |          ^ |      | |
		 | |           F1                 | |    F3    | |      | |
		 | |                            e16 |          | e20    | |
		e7 |                              | e17      e21 |      | |
		 | e6                             | v          | v      | |
		 | |                                 ---e23--->         | |
		 | |                              .  <--e22----  .      | |
		 | |                             V8             V11     | |
		 | |                                                    | |
		 | |                                                    | |
		 | |   V5              V6                               | |
		 | |    .  ---e10--->  .                                | |
		 | |       <--e11----                                   | |
		 | |    ^ |          ^ |                                | |
		 | |    | |    F2    | |                                | e3
		 | |   e8 |          | e12                             e2 |
		 | |    | e9       e13 |                                | |
		 | |    | v          | v                                | |
		 | |       ----e15-->                                   | |
		 | |    .  <---e14---  .                                | |
		 | |   V4              V7                               | |
		 | v                                                    | v
		    ----------------------------------e0--------------->
		 .  <---------------------------------e1----------------  .

		V0                                                        V1
	*/

	dcel := newDCELFromMultiPolygon(poly.AsPolygon().AsMultiPolygon(), inputBMask)

	eqInt(t, len(dcel.vertices), 12)
	eqInt(t, len(dcel.halfEdges), 24)
	eqInt(t, len(dcel.faces), 4)

	f0 := dcel.faces[0]
	f1 := dcel.faces[1]
	f2 := dcel.faces[2]
	f3 := dcel.faces[3]

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

	CheckVertexIncidents(t, dcel.vertices)
	CheckFaceComponents(
		t, f0,
		nil,
		[]XY{v3, v2, v1, v0},
	)
	CheckFaceComponents(
		t, f1,
		[]XY{v3, v0, v1, v2},
		[]XY{v5, v6, v7, v4},
		[]XY{v9, v10, v11, v8},
	)
	CheckFaceComponents(
		t, f2,
		[]XY{v4, v7, v6, v5},
	)
	CheckFaceComponents(
		t, f3,
		[]XY{v8, v11, v10, v9},
	)

	eqUint8(t, f0.label, inputBPresent)
	eqUint8(t, f1.label, inputBPresent|inputBValue)
	eqUint8(t, f2.label, inputBPresent)
	eqUint8(t, f3.label, inputBPresent)
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

	eqInt(t, len(dcel.vertices), 8)
	eqInt(t, len(dcel.halfEdges), 16)
	eqInt(t, len(dcel.faces), 3)

	f0 := dcel.faces[0]
	f1 := dcel.faces[1]
	f2 := dcel.faces[2]

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{1, 1}
	v3 := XY{0, 1}
	v4 := XY{2, 0}
	v5 := XY{3, 0}
	v6 := XY{3, 1}
	v7 := XY{2, 1}

	CheckVertexIncidents(t, dcel.vertices)
	CheckFaceComponents(
		t, f0,
		nil,
		[]XY{v3, v2, v1, v0},
		[]XY{v7, v6, v5, v4},
	)
	CheckFaceComponents(
		t, f1,
		[]XY{v0, v1, v2, v3},
	)
	CheckFaceComponents(
		t, f2,
		[]XY{v4, v5, v6, v7},
	)

	eqUint8(t, f0.label, inputBPresent)
	eqUint8(t, f1.label, inputBPresent|inputBValue)
	eqUint8(t, f2.label, inputBPresent|inputBValue)
}

func TestGraphReNode(t *testing.T) {
	poly, err := UnmarshalWKT("POLYGON((0 0,2 0,1 2,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	dcel := newDCELFromMultiPolygon(poly.AsPolygon().AsMultiPolygon(), inputAMask)

	other, err := UnmarshalWKT("POLYGON((0 1,2 1,1 3,0 1))")
	if err != nil {
		t.Fatal(err)
	}
	dcel.reNodeGraph(other.AsPolygon().Boundary().asLines())

	/*

	           v3
	          /  \
	         /    \
	        /      \
	      v4        v2
	      /    f1    \    f0
	     /            \
	    /              \
	   v0--------------v1

	*/

	eqInt(t, len(dcel.vertices), 5)
	eqInt(t, len(dcel.halfEdges), 10)
	eqInt(t, len(dcel.faces), 2)

	f0 := dcel.faces[0]
	f1 := dcel.faces[1]

	v0 := XY{0, 0}
	v1 := XY{2, 0}
	v2 := XY{1.5, 1}
	v3 := XY{1, 2}
	v4 := XY{0.5, 1}

	CheckVertexIncidents(t, dcel.vertices)
	CheckFaceComponents(
		t, f0,
		nil,
		[]XY{v1, v0, v4, v3, v2},
	)
	CheckFaceComponents(
		t, f1,
		[]XY{v0, v1, v2, v3, v4},
	)

	eqUint8(t, f0.label, inputAPresent)
	eqUint8(t, f1.label, inputAPresent|inputAValue)
}

func TestGraphReNodeTwoCutsInOneEdge(t *testing.T) {
	poly, err := UnmarshalWKT("POLYGON((0 0,1 2,2 0,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	dcel := newDCELFromMultiPolygon(poly.AsPolygon().AsMultiPolygon(), inputBMask)

	other, err := UnmarshalWKT("POLYGON((0 -1,1 1,2 -1,0 -1))")
	if err != nil {
		t.Fatal(err)
	}
	dcel.reNodeGraph(other.AsPolygon().Boundary().asLines())

	/*

	           v4
	          /  \
	         /    \
	        /      \
	       /        \
	      /    f1    \    f0
	     /            \
	    /              \
	   v0---v1----v2---v3

	*/

	eqInt(t, len(dcel.vertices), 5)
	eqInt(t, len(dcel.halfEdges), 10)
	eqInt(t, len(dcel.faces), 2)

	f0 := dcel.faces[0]
	f1 := dcel.faces[1]

	v0 := XY{0, 0}
	v1 := XY{0.5, 0}
	v2 := XY{1.5, 0}
	v3 := XY{2, 0}
	v4 := XY{1, 2}

	CheckVertexIncidents(t, dcel.vertices)
	CheckFaceComponents(
		t, f0,
		nil,
		[]XY{v0, v4, v3, v2, v1},
	)
	CheckFaceComponents(
		t, f1,
		[]XY{v0, v1, v2, v3, v4},
	)

	eqUint8(t, f0.label, inputBPresent)
	eqUint8(t, f1.label, inputBPresent|inputBValue)
}

func TestGraphReNodeOverlappingEdge(t *testing.T) {
	poly, err := UnmarshalWKT("POLYGON((0 0,0 2,2 2,2 0,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	dcel := newDCELFromMultiPolygon(poly.AsPolygon().AsMultiPolygon(), inputAMask)

	other, err := UnmarshalWKT("POLYGON((1 2,2 2,2 3,1 3,1 2))")
	if err != nil {
		t.Fatal(err)
	}
	dcel.reNodeGraph(other.AsPolygon().Boundary().asLines())

	/*

	  V0---V1---V2
	  |          |
	  |          |
	  |    f1    |   f0
	  |          |
	  |          |
	  V4--------V3

	*/

	eqInt(t, len(dcel.vertices), 5)
	eqInt(t, len(dcel.halfEdges), 10)
	eqInt(t, len(dcel.faces), 2)

	f0 := dcel.faces[0]
	f1 := dcel.faces[1]

	v0 := XY{0, 2}
	v1 := XY{1, 2}
	v2 := XY{2, 2}
	v3 := XY{2, 0}
	v4 := XY{0, 0}

	CheckVertexIncidents(t, dcel.vertices)
	CheckFaceComponents(
		t, f0,
		nil,
		[]XY{v0, v1, v2, v3, v4},
	)
	CheckFaceComponents(
		t, f1,
		[]XY{v4, v3, v2, v1, v0},
	)

	eqUint8(t, f0.label, inputAPresent)
	eqUint8(t, f1.label, inputAPresent|inputAValue)
}

func TestGraphOverlayDisjoint(t *testing.T) {
	polyA, err := UnmarshalWKT("POLYGON((0 0,1 0,1 1,0 1,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	dcelA := newDCELFromMultiPolygon(polyA.AsPolygon().AsMultiPolygon(), inputAMask)

	polyB, err := UnmarshalWKT("POLYGON((2 2,2 3,3 3,3 2,2 2))")
	if err != nil {
		t.Fatal(err)
	}
	dcelB := newDCELFromMultiPolygon(polyB.AsPolygon().AsMultiPolygon(), inputBMask)

	dcelA.reNodeGraph(polyB.AsPolygon().Boundary().asLines())
	dcelB.reNodeGraph(polyA.AsPolygon().Boundary().asLines())

	dcelA.overlay(dcelB)

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

	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v0, v1), []XY{v0, v1, v2, v3})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v1, v0), []XY{v3, v2, v1, v0})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v4, v5), []XY{v4, v5, v6, v7})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v5, v4), []XY{v7, v6, v5, v4})

	eqInt(t, len(dcelA.vertices), 8)
	eqInt(t, len(dcelA.halfEdges), 16)

	eqInt(t, len(dcelA.faces), 3)
	// TODO: trial and error was used to find the right permutation of face
	// labels. It relies on the permutation being stable. There is probably a
	// better way to test this.
	f0 := dcelA.faces[2]
	f1 := dcelA.faces[1]
	f2 := dcelA.faces[0]
	CheckFaceComponents(t, f0, nil, []XY{v4, v7, v6, v5}, []XY{v0, v3, v2, v1})
	CheckFaceComponents(t, f1, []XY{v0, v1, v2, v3})
	CheckFaceComponents(t, f2, []XY{v4, v5, v6, v7})

	eqUint8(t, f0.label, inputAPresent|inputBPresent)
	eqUint8(t, f1.label, inputAPresent|inputBPresent|inputAValue)
	eqUint8(t, f2.label, inputAPresent|inputBPresent|inputBValue)
}

func TestGraphOverlayIntersecting(t *testing.T) {
	polyA, err := UnmarshalWKT("POLYGON((0 0,1 2,2 0,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	dcelA := newDCELFromMultiPolygon(polyA.AsPolygon().AsMultiPolygon(), inputAMask)

	polyB, err := UnmarshalWKT("POLYGON((0 1,2 1,1 3,0 1))")
	if err != nil {
		t.Fatal(err)
	}
	dcelB := newDCELFromMultiPolygon(polyB.AsPolygon().AsMultiPolygon(), inputBMask)

	dcelA.reNodeGraph(polyB.AsPolygon().Boundary().asLines())
	dcelB.reNodeGraph(polyA.AsPolygon().Boundary().asLines())

	dcelA.overlay(dcelB)

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

	eqInt(t, len(dcelA.vertices), 8)
	eqInt(t, len(dcelA.halfEdges), 20)

	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v0, v1), []XY{v0, v1, v2, v4})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v6, v7), []XY{v6, v7, v5, v4, v3, v2})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v4, v2), []XY{v4, v2, v3})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v1, v0), []XY{v1, v0, v4, v5, v7, v6, v2})

	eqInt(t, len(dcelA.faces), 4)
	// TODO: trial and error was used to find the right permutation of face
	// labels. It relies on the permutation being stable. There is probably a
	// better way to test this.
	f0 := dcelA.faces[3]
	f1 := dcelA.faces[1]
	f2 := dcelA.faces[0]
	f3 := dcelA.faces[2]
	CheckFaceComponents(t, f0, nil, []XY{v0, v4, v5, v7, v6, v2, v1})
	CheckFaceComponents(t, f1, []XY{v0, v1, v2, v4})
	CheckFaceComponents(t, f2, []XY{v5, v4, v3, v2, v6, v7})
	CheckFaceComponents(t, f3, []XY{v4, v2, v3})

	eqUint8(t, f0.label, presenceMask)
	eqUint8(t, f1.label, presenceMask|inputAValue)
	eqUint8(t, f2.label, presenceMask|inputBValue)
	eqUint8(t, f3.label, presenceMask|inputAValue|inputBValue)
}

func TestGraphOverlayInside(t *testing.T) {
	polyA, err := UnmarshalWKT("POLYGON((0 0,3 0,3 3,0 3,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	dcelA := newDCELFromMultiPolygon(polyA.AsPolygon().AsMultiPolygon(), inputAMask)

	polyB, err := UnmarshalWKT("POLYGON((1 1,2 1,2 2,1 2,1 1))")
	if err != nil {
		t.Fatal(err)
	}
	dcelB := newDCELFromMultiPolygon(polyB.AsPolygon().AsMultiPolygon(), inputBMask)

	dcelA.reNodeGraph(polyB.AsPolygon().Boundary().asLines())
	dcelB.reNodeGraph(polyA.AsPolygon().Boundary().asLines())

	dcelA.overlay(dcelB)

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

	eqInt(t, len(dcelA.vertices), 8)
	eqInt(t, len(dcelA.halfEdges), 16)

	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v0, v1), []XY{v0, v1, v2, v3})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v1, v0), []XY{v0, v3, v2, v1})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v4, v5), []XY{v4, v5, v6, v7})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v5, v4), []XY{v5, v4, v7, v6})

	eqInt(t, len(dcelA.faces), 3)
	// TODO: trial and error was used to find the right permutation of face
	// labels. It relies on the permutation being stable. There is probably a
	// better way to test this.
	f0 := dcelA.faces[1]
	f1 := dcelA.faces[2]
	f2 := dcelA.faces[0]
	CheckFaceComponents(t, f0, nil, []XY{v1, v0, v3, v2})
	CheckFaceComponents(t, f1, []XY{v0, v1, v2, v3}, []XY{v7, v6, v5, v4})
	CheckFaceComponents(t, f2, []XY{v4, v5, v6, v7})

	eqUint8(t, f0.label, inputAPresent|inputBPresent)
	eqUint8(t, f1.label, inputAPresent|inputBPresent|inputAValue)
	eqUint8(t, f2.label, inputAPresent|inputBPresent|inputAValue|inputBValue)
}

func TestGraphOverlayReproduceHorizontalHoleLinkageBug(t *testing.T) {
	polyA, err := UnmarshalWKT("MULTIPOLYGON(((4 0,4 1,5 1,5 0,4 0)),((1 0,1 2,3 2,3 0,1 0)))")
	if err != nil {
		t.Fatal(err)
	}
	dcelA := newDCELFromGeometry(polyA, inputAMask)

	polyB, err := UnmarshalWKT("MULTIPOLYGON(((0 4,0 5,1 5,1 4,0 4)),((0 1,0 3,2 3,2 1,0 1)))")
	if err != nil {
		t.Fatal(err)
	}
	dcelB := newDCELFromGeometry(polyB, inputBMask)

	dcelA.reNodeGraph(polyB.AsMultiPolygon().Boundary().asLines())
	dcelB.reNodeGraph(polyA.AsMultiPolygon().Boundary().asLines())

	dcelA.overlay(dcelB)

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

	eqInt(t, len(dcelA.vertices), 18)
	eqInt(t, len(dcelA.halfEdges), 40)

	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v5, v6), []XY{v5, v6, v7, v8})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v1, v2), []XY{v1, v2, v3, v18, v10, v17})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v17, v10), []XY{v17, v10, v18, v4})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v9, v17), []XY{v9, v17, v4, v18, v11, v12})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v13, v14), []XY{v13, v14, v15, v16})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v6, v5), []XY{v6, v5, v8, v7})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v14, v13), []XY{v14, v13, v16, v15})
	CheckHalfEdgeLoop(t, findEdge(t, dcelA, v2, v1), []XY{v1, v17, v9, v12, v11, v18, v3, v2})

	eqInt(t, len(dcelA.faces), 6)
	// TODO: trial and error was used to find the right permutation of face
	// labels. It relies on the permutation being stable. There is probably a
	// better way to test this.
	f0 := dcelA.faces[5]
	f1 := dcelA.faces[3]
	f2 := dcelA.faces[1]
	f3 := dcelA.faces[4]
	f4 := dcelA.faces[0]
	f5 := dcelA.faces[2]
	CheckFaceComponents(t, f0, nil,
		[]XY{v14, v13, v16, v15},
		[]XY{v6, v5, v8, v7},
		[]XY{v1, v17, v9, v12, v11, v18, v3, v2},
	)
	CheckFaceComponents(t, f1, []XY{v5, v6, v7, v8})
	CheckFaceComponents(t, f2, []XY{v13, v14, v15, v16})
	CheckFaceComponents(t, f3, []XY{v1, v2, v3, v18, v10, v17})
	CheckFaceComponents(t, f4, []XY{v12, v9, v17, v4, v18, v11})
	CheckFaceComponents(t, f5, []XY{v17, v10, v18, v4})

	eqUint8(t, f0.label, inputBPresent|inputAPresent)
	eqUint8(t, f1.label, inputBPresent|inputAPresent|inputAValue)
	eqUint8(t, f2.label, inputBPresent|inputAPresent|inputBValue)
	eqUint8(t, f3.label, inputBPresent|inputAPresent|inputAValue)
	eqUint8(t, f4.label, inputBPresent|inputAPresent|inputBValue)
	eqUint8(t, f5.label, inputBPresent|inputAPresent|inputBValue|inputAValue)
}

func eqInt(t *testing.T, i1, i2 int) {
	t.Helper()
	if i1 != i2 {
		t.Errorf("ints not equal: %d vs %d", i1, i2)
	}
}

func eqUint8(t *testing.T, u1, u2 uint8) {
	t.Helper()
	if u1 != u2 {
		t.Errorf("uint8s not equal: %d vs %d", u1, u2)
	}
}
