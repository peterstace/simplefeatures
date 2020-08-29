package geom

import (
	"testing"
)

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
	dcel := newDCELFromPolygon(poly.AsPolygon())

	/*

		V2 .
		      ^ \
		   ^|  \ \
		   ||   \ \
		   |e4   \ e3   f0
		   ||    e2 \
		  e5|      \ \
		   ||  f1   \ \
		   |v        \ v
		     ---e0--->
		V0 . <---e1---  . V1

	*/

	eqInt(t, len(dcel.vertices), 3)
	eqInt(t, len(dcel.halfEdges), 6)
	eqInt(t, len(dcel.faces), 2)

	f0 := dcel.faces[0]
	f1 := dcel.faces[1]

	v0 := XY{0, 0}
	v1 := XY{1, 0}
	v2 := XY{0, 1}

	CheckVertexIncidents(t, dcel.vertices)
	CheckFaceComponents(
		t, f0,
		nil,
		[]XY{v2, v1, v0},
	)
	CheckFaceComponents(
		t, f1,
		[]XY{v0, v1, v2},
	)
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

	dcel := newDCELFromPolygon(poly.AsPolygon())

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
}

func TestGraphReNode(t *testing.T) {
	poly, err := UnmarshalWKT("POLYGON((0 0,2 0,1 2,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	dcel := newDCELFromPolygon(poly.AsPolygon())

	other, err := UnmarshalWKT("POLYGON((0 1,2 1,1 3,0 1))")
	if err != nil {
		t.Fatal(err)
	}
	dcel.reNodeGraph(other.AsPolygon())

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
}

func TestGraphReNodeTwoCutsInOneEdge(t *testing.T) {
	poly, err := UnmarshalWKT("POLYGON((0 0,1 2,2 0,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	dcel := newDCELFromPolygon(poly.AsPolygon())

	other, err := UnmarshalWKT("POLYGON((0 -1,1 1,2 -1,0 -1))")
	if err != nil {
		t.Fatal(err)
	}
	dcel.reNodeGraph(other.AsPolygon())

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
}

func TestGraphReNodeOverlappingEdge(t *testing.T) {
	poly, err := UnmarshalWKT("POLYGON((0 0,0 2,2 2,2 0,0 0))")
	if err != nil {
		t.Fatal(err)
	}
	dcel := newDCELFromPolygon(poly.AsPolygon())

	other, err := UnmarshalWKT("POLYGON((1 2,2 2,2 3,1 3,1 2))")
	if err != nil {
		t.Fatal(err)
	}
	dcel.reNodeGraph(other.AsPolygon())

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
}

func eqInt(t *testing.T, i1, i2 int) {
	t.Helper()
	if i1 != i2 {
		t.Errorf("ints not equal: %d vs %d", i1, i2)
	}
}