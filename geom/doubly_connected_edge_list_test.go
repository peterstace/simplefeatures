package geom

import "testing"

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

	v0 := dcel.vertices[XY{0, 0}]
	v1 := dcel.vertices[XY{1, 0}]
	v2 := dcel.vertices[XY{0, 1}]

	e0 := dcel.halfEdges[0]
	e1 := dcel.halfEdges[1]
	e2 := dcel.halfEdges[2]
	e3 := dcel.halfEdges[3]
	e4 := dcel.halfEdges[4]
	e5 := dcel.halfEdges[5]

	f0 := dcel.faces[0]
	f1 := dcel.faces[1]

	// half edges should have their incident face populated
	eqFace(t, e0.incident, f1)
	eqFace(t, e1.incident, f0)
	eqFace(t, e2.incident, f1)
	eqFace(t, e3.incident, f0)
	eqFace(t, e4.incident, f1)
	eqFace(t, e5.incident, f0)

	// half edge twins should be populated
	eqEdge(t, e0.twin, e1)
	eqEdge(t, e1.twin, e0)
	eqEdge(t, e2.twin, e3)
	eqEdge(t, e3.twin, e2)
	eqEdge(t, e4.twin, e5)
	eqEdge(t, e5.twin, e4)

	// next edge should be populated
	eqEdge(t, e0.next, e2)
	eqEdge(t, e2.next, e4)
	eqEdge(t, e4.next, e0)
	eqEdge(t, e1.next, e5)
	eqEdge(t, e5.next, e3)
	eqEdge(t, e3.next, e1)

	// prev edge should be populated
	eqEdge(t, e0.prev, e4)
	eqEdge(t, e4.prev, e2)
	eqEdge(t, e2.prev, e0)
	eqEdge(t, e1.prev, e3)
	eqEdge(t, e3.prev, e5)
	eqEdge(t, e5.prev, e1)

	// edge origins should be populated
	eqVertex(t, e0.origin, v0)
	eqVertex(t, e1.origin, v1)
	eqVertex(t, e2.origin, v1)
	eqVertex(t, e3.origin, v2)
	eqVertex(t, e4.origin, v2)
	eqVertex(t, e5.origin, v0)

	// vertex incidents should be populated
	eqEdge(t, v0.incident, e0)
	eqEdge(t, v1.incident, e2)
	eqEdge(t, v2.incident, e4)

	// face components are populated
	eqEdge(t, f0.outerComponent, nil)
	eqInt(t, len(f0.innerComponents), 1)
	eqEdge(t, f0.innerComponents[0], e1)
	eqEdge(t, f1.outerComponent, e0)
	eqInt(t, len(f1.innerComponents), 0)
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

	v0 := dcel.vertices[XY{0, 0}]
	v1 := dcel.vertices[XY{5, 0}]
	v2 := dcel.vertices[XY{5, 5}]
	v3 := dcel.vertices[XY{0, 5}]
	v4 := dcel.vertices[XY{1, 1}]
	v5 := dcel.vertices[XY{1, 2}]
	v6 := dcel.vertices[XY{2, 2}]
	v7 := dcel.vertices[XY{2, 1}]
	v8 := dcel.vertices[XY{3, 3}]
	v9 := dcel.vertices[XY{3, 4}]
	v10 := dcel.vertices[XY{4, 4}]
	v11 := dcel.vertices[XY{4, 3}]

	e0 := dcel.halfEdges[0]
	e1 := dcel.halfEdges[1]
	e2 := dcel.halfEdges[2]
	e3 := dcel.halfEdges[3]
	e4 := dcel.halfEdges[4]
	e5 := dcel.halfEdges[5]
	e6 := dcel.halfEdges[6]
	e7 := dcel.halfEdges[7]
	e8 := dcel.halfEdges[8]
	e9 := dcel.halfEdges[9]
	e10 := dcel.halfEdges[10]
	e11 := dcel.halfEdges[11]
	e12 := dcel.halfEdges[12]
	e13 := dcel.halfEdges[13]
	e14 := dcel.halfEdges[14]
	e15 := dcel.halfEdges[15]
	e16 := dcel.halfEdges[16]
	e17 := dcel.halfEdges[17]
	e18 := dcel.halfEdges[18]
	e19 := dcel.halfEdges[19]
	e20 := dcel.halfEdges[20]
	e21 := dcel.halfEdges[21]
	e22 := dcel.halfEdges[22]
	e23 := dcel.halfEdges[23]

	f0 := dcel.faces[0]
	f1 := dcel.faces[1]
	f2 := dcel.faces[2]
	f3 := dcel.faces[3]

	// half edges should have their incident face populated
	eqFace(t, e0.incident, f1)
	eqFace(t, e1.incident, f0)
	eqFace(t, e2.incident, f1)
	eqFace(t, e3.incident, f0)
	eqFace(t, e4.incident, f1)
	eqFace(t, e5.incident, f0)
	eqFace(t, e6.incident, f1)
	eqFace(t, e7.incident, f0)
	eqFace(t, e8.incident, f1)
	eqFace(t, e9.incident, f2)
	eqFace(t, e10.incident, f1)
	eqFace(t, e11.incident, f2)
	eqFace(t, e12.incident, f1)
	eqFace(t, e13.incident, f2)
	eqFace(t, e14.incident, f1)
	eqFace(t, e15.incident, f2)
	eqFace(t, e16.incident, f1)
	eqFace(t, e17.incident, f3)
	eqFace(t, e18.incident, f1)
	eqFace(t, e19.incident, f3)
	eqFace(t, e20.incident, f1)
	eqFace(t, e21.incident, f3)
	eqFace(t, e22.incident, f1)
	eqFace(t, e23.incident, f3)

	// half edge twins should be populated
	eqEdge(t, e0.twin, e1)
	eqEdge(t, e1.twin, e0)
	eqEdge(t, e2.twin, e3)
	eqEdge(t, e3.twin, e2)
	eqEdge(t, e4.twin, e5)
	eqEdge(t, e5.twin, e4)
	eqEdge(t, e6.twin, e7)
	eqEdge(t, e7.twin, e6)
	eqEdge(t, e8.twin, e9)
	eqEdge(t, e9.twin, e8)
	eqEdge(t, e10.twin, e11)
	eqEdge(t, e11.twin, e10)
	eqEdge(t, e12.twin, e13)
	eqEdge(t, e13.twin, e12)
	eqEdge(t, e14.twin, e15)
	eqEdge(t, e15.twin, e14)
	eqEdge(t, e16.twin, e17)
	eqEdge(t, e17.twin, e16)
	eqEdge(t, e18.twin, e19)
	eqEdge(t, e19.twin, e18)
	eqEdge(t, e20.twin, e21)
	eqEdge(t, e21.twin, e20)
	eqEdge(t, e22.twin, e23)
	eqEdge(t, e23.twin, e22)

	// next edge should be populated
	eqEdge(t, e0.next, e2)
	eqEdge(t, e1.next, e7)
	eqEdge(t, e2.next, e4)
	eqEdge(t, e3.next, e1)
	eqEdge(t, e4.next, e6)
	eqEdge(t, e5.next, e3)
	eqEdge(t, e6.next, e0)
	eqEdge(t, e7.next, e5)
	eqEdge(t, e8.next, e10)
	eqEdge(t, e9.next, e15)
	eqEdge(t, e10.next, e12)
	eqEdge(t, e11.next, e9)
	eqEdge(t, e12.next, e14)
	eqEdge(t, e13.next, e11)
	eqEdge(t, e14.next, e8)
	eqEdge(t, e15.next, e13)
	eqEdge(t, e16.next, e18)
	eqEdge(t, e17.next, e23)
	eqEdge(t, e18.next, e20)
	eqEdge(t, e19.next, e17)
	eqEdge(t, e20.next, e22)
	eqEdge(t, e21.next, e19)
	eqEdge(t, e22.next, e16)
	eqEdge(t, e23.next, e21)

	// prev edge should be populated
	eqEdge(t, e2.prev, e0)
	eqEdge(t, e7.prev, e1)
	eqEdge(t, e4.prev, e2)
	eqEdge(t, e1.prev, e3)
	eqEdge(t, e6.prev, e4)
	eqEdge(t, e3.prev, e5)
	eqEdge(t, e0.prev, e6)
	eqEdge(t, e5.prev, e7)
	eqEdge(t, e10.prev, e8)
	eqEdge(t, e15.prev, e9)
	eqEdge(t, e12.prev, e10)
	eqEdge(t, e9.prev, e11)
	eqEdge(t, e14.prev, e12)
	eqEdge(t, e11.prev, e13)
	eqEdge(t, e8.prev, e14)
	eqEdge(t, e13.prev, e15)
	eqEdge(t, e18.prev, e16)
	eqEdge(t, e23.prev, e17)
	eqEdge(t, e20.prev, e18)
	eqEdge(t, e17.prev, e19)
	eqEdge(t, e22.prev, e20)
	eqEdge(t, e19.prev, e21)
	eqEdge(t, e16.prev, e22)
	eqEdge(t, e21.prev, e23)

	// edge origins should be populated
	eqVertex(t, e0.origin, v0)
	eqVertex(t, e1.origin, v1)
	eqVertex(t, e2.origin, v1)
	eqVertex(t, e3.origin, v2)
	eqVertex(t, e4.origin, v2)
	eqVertex(t, e5.origin, v3)
	eqVertex(t, e6.origin, v3)
	eqVertex(t, e7.origin, v0)
	eqVertex(t, e8.origin, v4)
	eqVertex(t, e9.origin, v5)
	eqVertex(t, e10.origin, v5)
	eqVertex(t, e11.origin, v6)
	eqVertex(t, e12.origin, v6)
	eqVertex(t, e13.origin, v7)
	eqVertex(t, e14.origin, v7)
	eqVertex(t, e15.origin, v4)
	eqVertex(t, e16.origin, v8)
	eqVertex(t, e17.origin, v9)
	eqVertex(t, e18.origin, v9)
	eqVertex(t, e19.origin, v10)
	eqVertex(t, e20.origin, v10)
	eqVertex(t, e21.origin, v11)
	eqVertex(t, e22.origin, v11)
	eqVertex(t, e23.origin, v8)

	// vertex incidents should be populated
	eqEdge(t, v0.incident, e0)
	eqEdge(t, v1.incident, e2)
	eqEdge(t, v2.incident, e4)
	eqEdge(t, v3.incident, e6)
	eqEdge(t, v4.incident, e8)
	eqEdge(t, v5.incident, e10)
	eqEdge(t, v6.incident, e12)
	eqEdge(t, v7.incident, e14)
	eqEdge(t, v8.incident, e16)
	eqEdge(t, v9.incident, e18)
	eqEdge(t, v10.incident, e20)
	eqEdge(t, v11.incident, e22)

	// face components are populated
	eqEdge(t, f0.outerComponent, nil)
	eqInt(t, len(f0.innerComponents), 1)
	eqEdge(t, f0.innerComponents[0], e1)
	eqEdge(t, f1.outerComponent, e0)
	eqInt(t, len(f1.innerComponents), 2)
	eqEdge(t, f1.innerComponents[0], e8)
	eqEdge(t, f1.innerComponents[1], e16)
	eqEdge(t, f2.outerComponent, e9)
	eqInt(t, len(f2.innerComponents), 0)
	eqEdge(t, f3.outerComponent, e17)
	eqInt(t, len(f2.innerComponents), 0)
}

func eqFace(t *testing.T, f1, f2 *faceRecord) {
	t.Helper()
	if f1 != f2 {
		t.Errorf("faces not equal: %p vs %p", f1, f2)
	}
}

func eqEdge(t *testing.T, e1, e2 *halfEdgeRecord) {
	t.Helper()
	if e1 != e2 {
		t.Errorf("edges not equal: %p vs %p", e1, e2)
	}
}

func eqVertex(t *testing.T, v1, v2 *vertexRecord) {
	t.Helper()
	if v1 != v2 {
		t.Errorf("vertices not equal: %p vs %p", v1, v2)
	}
}

func eqInt(t *testing.T, i1, i2 int) {
	t.Helper()
	if i1 != i2 {
		t.Errorf("ints not equal: %d vs %d", i1, i2)
	}
}
