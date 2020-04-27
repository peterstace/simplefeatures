package geom

func newNearestPoint(target Point) nearestPoint {
	return nearestPoint{target: target}
}

type nearestPoint struct {
	target Point
	point  Point
	dist   float64
}

func (n *nearestPoint) add(candidate Point) {
	targetXY, ok := n.target.XY()
	if !ok {
		return
	}
	candidateXY, ok := candidate.XY()
	if !ok {
		return
	}

	delta := targetXY.Sub(candidateXY)
	candidateDist := delta.Dot(delta)
	if n.point.IsEmpty() || candidateDist < n.dist {
		n.dist = candidateDist
		n.point = candidate
	}
}
