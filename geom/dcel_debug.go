package geom

import (
	"fmt"
	"log"
	"sort"
)

type namedDCEL struct {
	*doublyConnectedEdgeList

	vertexNames map[*vertexRecord]string
	edgeNames   map[*halfEdgeRecord]string
	faceNames   map[*faceRecord]string

	vertexList []*vertexRecord
	edgeList   []*halfEdgeRecord
}

func newNamedDCEL(d *doublyConnectedEdgeList) *namedDCEL {
	var vertexList []*vertexRecord
	for _, v := range d.vertices {
		vertexList = append(vertexList, v)
	}
	sort.Slice(vertexList, func(i, j int) bool {
		return ptrLess(vertexList[i], vertexList[j])
	})
	vertexNames := make(map[*vertexRecord]string)
	for i, v := range vertexList {
		vertexNames[v] = fmt.Sprintf("v%0*d", intLog(10, len(vertexList)), i)
	}

	var edgeList []*halfEdgeRecord
	for _, e := range d.halfEdges {
		edgeList = append(edgeList, e)
	}
	sort.Slice(edgeList, func(i, j int) bool {
		return ptrLess(edgeList[i], edgeList[j])
	})
	edgeNames := make(map[*halfEdgeRecord]string)
	for i, e := range edgeList {
		edgeNames[e] = fmt.Sprintf("e%0*d", intLog(10, len(edgeList)), i)
	}

	sort.Slice(d.faces, func(i, j int) bool {
		return ptrLess(d.faces[i], d.faces[j])
	})
	faceNames := make(map[*faceRecord]string)
	for i, f := range d.faces {
		faceNames[f] = fmt.Sprintf("f%0*d", intLog(10, len(d.faces)), i)
	}

	return &namedDCEL{
		doublyConnectedEdgeList: d,
		vertexNames:             vertexNames,
		edgeNames:               edgeNames,
		faceNames:               faceNames,
		vertexList:              vertexList,
		edgeList:                edgeList,
	}
}

func (n *namedDCEL) show() {
	log.Printf("vertices: %d", len(n.vertices))
	for _, v := range n.vertexList {
		log.Printf("\t%s: %s", n.vertexNames[v], n.vertexRepr(v))
	}

	log.Printf("halfEdges: %d", len(n.halfEdges))
	for _, e := range n.edgeList {
		log.Printf("\t%s: %s", n.edgeNames[e], n.edgeRepr(e))
	}

	log.Printf("faces: %d", len(n.faces))
	for _, f := range n.faces {
		log.Printf("\t%s: %s", n.faceNames[f], n.faceRepr(f))
	}
}

func (n *namedDCEL) faceRepr(f *faceRecord) string {
	if f == nil {
		return "nil"
	}
	return fmt.Sprintf("cycle:%s inSet:%s", n.edgeNames[f.cycle], bstoa(f.inSet))
}

func (n *namedDCEL) edgeRepr(e *halfEdgeRecord) string {
	if e == nil {
		return "nil"
	}
	return fmt.Sprintf(
		"origin:%s twin:%s incident:%s next:%s prev:%s srcEdge:%s srcFace:%s inSet:%s xys:%v",
		n.vertexNames[e.origin], n.edgeNames[e.twin], n.faceNames[e.incident], n.edgeNames[e.next],
		n.edgeNames[e.prev], bstoa(e.srcEdge), bstoa(e.srcFace), bstoa(e.inSet), e.xys())
}

func (n *namedDCEL) vertexRepr(v *vertexRecord) string {
	if v == nil {
		return "nil"
	}
	var incidents []string
	for _, inc := range v.incidents {
		incidents = append(incidents, n.edgeNames[inc])
	}
	return fmt.Sprintf(
		"src:%s inSet:%s coords:%v incidents:%v",
		bstoa(v.src), bstoa(v.inSet), v.coords, incidents,
	)
}

func btoa(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func bstoa(b [2]bool) string {
	return btoa(b[0]) + btoa(b[1])
}

func ptoa(ptr interface{}) string {
	return fmt.Sprintf("%p", ptr)
}

func ptrLess(ptr1, ptr2 interface{}) bool {
	return ptoa(ptr1) < ptoa(ptr2)
}

// intLog finds the smallest exponent such that base^exponent >= power.
func intLog(base, power int) int {
	exponent := 0
	product := 1
	for product < power {
		product *= base
		exponent++
	}
	return exponent
}
