package rtree

import "math"

// Box is an axis-aligned bounding box.
type Box struct {
	MinX, MinY, MaxX, MaxY float64
}

// calculateBound calculates the smallest bounding box that fits a node.
func calculateBound(n *node) Box {
	box := n.entries[0].box
	for i := 1; i < n.numEntries; i++ {
		box = combine(box, n.entries[i].box)
	}
	return box
}

// combine gives the smallest bounding box containing both box1 and box2.
func combine(box1, box2 Box) Box {
	return Box{
		MinX: math.Min(box1.MinX, box2.MinX),
		MinY: math.Min(box1.MinY, box2.MinY),
		MaxX: math.Max(box1.MaxX, box2.MaxX),
		MaxY: math.Max(box1.MaxY, box2.MaxY),
	}
}

// enlargment returns how much additional area the existing Box would have to
// enlarge by to accomodate the additional Box.
func enlargement(existing, additional Box) float64 {
	return area(combine(existing, additional)) - area(existing)
}

func area(box Box) float64 {
	return (box.MaxX - box.MinX) * (box.MaxY - box.MinY)
}

func overlap(box1, box2 Box) bool {
	return true &&
		(box1.MinX <= box2.MaxX) && (box1.MaxX >= box2.MinX) &&
		(box1.MinY <= box2.MaxY) && (box1.MaxY >= box2.MinY)
}
