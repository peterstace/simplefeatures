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
}

func newNamedDCEL(d *doublyConnectedEdgeList) *namedDCEL {
	return &namedDCEL{
		doublyConnectedEdgeList: d,
		vertexNames:             buildVertexNames(d.vertices),
		edgeNames:               buildEdgeNames(d.halfEdges),
		faceNames:               buildFaceNames(d.faces),
	}
}

func buildVertexNames(vertices map[XY]*vertexRecord) map[*vertexRecord]string {
	var vertexList []*vertexRecord
	for _, v := range vertices {
		vertexList = append(vertexList, v)
	}
	sort.Slice(vertexList, func(i, j int) bool {
		vi := vertexList[i]
		vj := vertexList[j]
		return vi.less(vj)
	})
	vertexNames := make(map[*vertexRecord]string)
	for i, v := range vertexList {
		vertexNames[v] = fmt.Sprintf("v%d", i)
	}
	return vertexNames
}

func buildEdgeNames(edges edgeSet) map[*halfEdgeRecord]string {
	var edgeList []*halfEdgeRecord
	for _, e := range edges {
		edgeList = append(edgeList, e)
	}
	sort.Slice(edgeList, func(i, j int) bool {
		ei := edgeList[i]
		ej := edgeList[j]
		return ei.less(ej)
	})
	edgeNames := make(map[*halfEdgeRecord]string)
	for i, e := range edgeList {
		edgeNames[e] = fmt.Sprintf("e%02d", i)
	}
	return edgeNames
}

func buildFaceNames(faces []*faceRecord) map[*faceRecord]string {
	sort.Slice(faces, func(i, j int) bool {
		fi := faces[i]
		fj := faces[j]
		return fi.less(fj)
	})
	faceNames := make(map[*faceRecord]string)
	for i, f := range faces {
		faceNames[f] = fmt.Sprintf("f%d", i)
	}
	return faceNames
}

func (n *namedDCEL) show() {
	log.Printf("faces: %d", len(n.faces))
	for _, f := range n.faces {
		log.Printf("\t%s: %s", n.faceNames[f], n.faceRepr(f))
	}
	log.Printf("halfEdges: %d", len(n.halfEdges))
	for _, e := range n.halfEdges {
		log.Printf("\t%s: %s", n.edgeNames[e], n.edgeRepr(e))
	}
	log.Printf("vertices: %d", len(n.vertices))
	for _, v := range n.vertices {
		log.Printf("\t%s: %s", n.vertexNames[v], n.vertexRepr(v))
	}
}

func (n *namedDCEL) faceRepr(f *faceRecord) string {
	if f == nil {
		return "nil"
	}
	return fmt.Sprintf("cycle:%s inSet:%s", n.edgeRepr(f.cycle), bstoa(f.inSet))
}

func (n *namedDCEL) edgeRepr(e *halfEdgeRecord) string {
	if e == nil {
		return "nil"
	}
	return fmt.Sprintf(
		"origin:%s twin:%s incident:%s next:%s prev:%s edgeInSet:%s faceInSet:%s xys:%v",
		n.vertexNames[e.origin], n.edgeNames[e.twin], n.faceNames[e.incident], n.edgeNames[e.next], n.edgeNames[e.prev], bstoa(e.edgeInSet), bstoa(e.faceInSet), e.xys())
}

func (n *namedDCEL) vertexRepr(v *vertexRecord) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprintf(
		"inSet:%s coords:%v incidents:%v",
		bstoa(v.inSet), v.coords, v.incidents,
	)
}

func bstoa(b [2]bool) string {
	var s string
	for i := 0; i < 2; i++ {
		if b[i] {
			s += "1"
		} else {
			s += "0"
		}
	}
	return s
}
