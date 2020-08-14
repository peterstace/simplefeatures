package geom

import (
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/rtree"
)

func dispatchDistance(g1, g2 Geometry) (float64, bool) {
	if g1.Intersects(g2) {
		return 0, true
	}

	xys1, lns1 := extractXYsAndLines(g1)
	xys2, lns2 := extractXYsAndLines(g2)
	tr := loadTree(xys2, lns2)

	minDist := math.Inf(+1)

	for _, xy := range xys1 {
		xyEnv := NewEnvelope(xy)
		tr.PrioritySearch(xyEnv.box(), func(recordID int) error {
			var recordEnv Envelope
			if recordID > 0 {
				recordEnv = NewEnvelope(xys2[recordID-1])
			} else {
				recordEnv = lns2[-recordID-1].envelope()
			}
			envDist := recordEnv.Distance(xyEnv)
			if envDist > minDist {
				return rtree.Stop
			}

			var dist float64
			if recordID > 0 {
				dist = distBetweenXYs(xy, xys2[recordID-1])
			} else {
				dist = distBetweenXYAndLine(xy, lns2[-recordID-1])
			}
			if dist < minDist {
				minDist = dist
			}
			return nil
		})
	}

	for _, ln := range lns1 {
		lnEnv := ln.envelope()
		tr.PrioritySearch(lnEnv.box(), func(recordID int) error {
			var recordEnv Envelope
			if recordID > 0 {
				recordEnv = NewEnvelope(xys2[recordID-1])
			} else {
				recordEnv = lns2[-recordID-1].envelope()
			}
			envDist := recordEnv.Distance(lnEnv)
			if envDist > minDist {
				return rtree.Stop
			}

			var dist float64
			if recordID > 0 {
				dist = distBetweenXYAndLine(xys2[recordID-1], ln)
			} else {
				dist = distBetweenLineAndLine(lns2[-recordID-1], ln)
			}
			if dist < minDist {
				minDist = dist
			}
			return nil
		})
	}

	if math.IsInf(minDist, +1) {
		return 0, false
	}
	return minDist, true
}

func extractXYsAndLines(g Geometry) ([]XY, []line) {
	switch g.Type() {
	case TypePoint:
		return g.AsPoint().asXYs(), nil
	case TypeLineString:
		return nil, g.AsLineString().asLines()
	case TypePolygon:
		return nil, g.AsPolygon().Boundary().asLines()
	case TypeMultiPoint:
		return g.AsMultiPoint().asXYs(), nil
	case TypeMultiLineString:
		return nil, g.AsMultiLineString().asLines()
	case TypeMultiPolygon:
		return nil, g.AsMultiPolygon().Boundary().asLines()
	case TypeGeometryCollection:
		var allXYs []XY
		var allLines []line
		g.AsGeometryCollection().walk(func(child Geometry) {
			xys, lns := extractXYsAndLines(child)
			allXYs = append(allXYs, xys...)
			allLines = append(allLines, lns...)
		})
		return allXYs, allLines
	default:
		panic(fmt.Sprintf("implementation error: unhandled geometry types %s", g.Type()))
	}
}

// loadTree creates a new RTree that indexes both the XYs and the lines. It
// uses positive record IDs to refer to the XYs, and negative recordIDs to
// refer to the lines. Because +0 and -0 are the same, indexing is 1-based and
// recordID 0 is not used.
func loadTree(xys []XY, lns []line) *rtree.RTree {
	items := make([]rtree.BulkItem, len(xys)+len(lns))
	for i, xy := range xys {
		items[i] = rtree.BulkItem{
			Box:      NewEnvelope(xy).box(),
			RecordID: i + 1,
		}
	}
	for i, ln := range lns {
		items[i+len(xys)] = rtree.BulkItem{
			Box:      ln.envelope().box(),
			RecordID: -(i + 1),
		}
	}
	return rtree.BulkLoad(items)
}

func distBetweenXYs(xy1, xy2 XY) float64 {
	sub := xy1.Sub(xy2)
	return math.Sqrt(sub.Dot(sub))
}

func distBetweenXYAndLine(xy XY, ln line) float64 {
	ab := ln.b.Sub(ln.a)
	abLen := ab.Length()
	proj := xy.Sub(ln.a).Dot(ab) / abLen
	var closest XY
	switch {
	case proj < 0:
		closest = ln.a
	case proj > abLen:
		closest = ln.b
	default:
		scaled := ab.Scale(proj / abLen)
		closest = scaled.Add(ln.a)
	}
	return distBetweenXYs(xy, closest)
}

func distBetweenLineAndLine(ln1, ln2 line) float64 {
	minDist := math.Inf(+1)
	for _, dist := range [4]float64{
		distBetweenXYAndLine(ln1.a, ln2),
		distBetweenXYAndLine(ln1.b, ln2),
		distBetweenXYAndLine(ln2.a, ln1),
		distBetweenXYAndLine(ln2.b, ln1),
	} {
		if dist < minDist {
			minDist = dist
		}
	}
	return minDist
}
