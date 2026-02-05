package jts

// OperationValid_PolygonRing_IsShell tests if a polygon ring represents a shell.
func OperationValid_PolygonRing_IsShell(polyRing *OperationValid_PolygonRing) bool {
	if polyRing == nil {
		return true
	}
	return polyRing.IsShell()
}

// OperationValid_PolygonRing_AddTouch records a touch location between two rings,
// and checks if the rings already touch in a different location.
func OperationValid_PolygonRing_AddTouch(ring0, ring1 *OperationValid_PolygonRing, pt *Geom_Coordinate) bool {
	//--- skip if either polygon does not have holes
	if ring0 == nil || ring1 == nil {
		return false
	}

	//--- only record touches within a polygon
	if !ring0.IsSamePolygon(ring1) {
		return false
	}

	if !ring0.isOnlyTouch(ring1, pt) {
		return true
	}
	if !ring1.isOnlyTouch(ring0, pt) {
		return true
	}

	ring0.addTouch(ring1, pt)
	ring1.addTouch(ring0, pt)
	return false
}

// OperationValid_PolygonRing_FindHoleCycleLocation finds a location (if any) where a chain of holes forms a cycle
// in the ring touch graph.
// The shell may form part of the chain as well.
// This indicates that a set of holes disconnects the interior of a polygon.
func OperationValid_PolygonRing_FindHoleCycleLocation(polyRings []*OperationValid_PolygonRing) *Geom_Coordinate {
	for _, polyRing := range polyRings {
		if !polyRing.isInTouchSet() {
			holeCycleLoc := polyRing.findHoleCycleLocation()
			if holeCycleLoc != nil {
				return holeCycleLoc
			}
		}
	}
	return nil
}

// OperationValid_PolygonRing_FindInteriorSelfNode finds a location of an interior self-touch in a list of rings,
// if one exists.
// This indicates that a self-touch disconnects the interior of a polygon,
// which is invalid.
func OperationValid_PolygonRing_FindInteriorSelfNode(polyRings []*OperationValid_PolygonRing) *Geom_Coordinate {
	for _, polyRing := range polyRings {
		interiorSelfNode := polyRing.FindInteriorSelfNode()
		if interiorSelfNode != nil {
			return interiorSelfNode
		}
	}
	return nil
}

// OperationValid_PolygonRing is a ring of a polygon being analyzed for topological validity.
// The shell and hole rings of valid polygons touch only at discrete points.
// The "touch" relationship induces a graph over the set of rings.
// The interior of a valid polygon must be connected.
// This is the case if there is no "chain" of touching rings
// (which would partition off part of the interior).
// This is equivalent to the touch graph having no cycles.
// Thus the touch graph of a valid polygon is a forest - a set of disjoint trees.
//
// Also, in a valid polygon two rings can touch only at a single location,
// since otherwise they disconnect a portion of the interior between them.
// This is checked as the touches relation is built
// (so the touch relation representation for a polygon ring does not need to support
// more than one touch location for each adjacent ring).
//
// The cycle detection algorithm works for polygon rings which also contain self-touches
// (inverted shells and exverted holes).
//
// Polygons with no holes do not need to be checked for
// a connected interior, unless self-touches are allowed.
// The class also records the topology at self-touch nodes,
// to support checking if an invalid self-touch disconnects the polygon.
type OperationValid_PolygonRing struct {
	id    int
	shell *OperationValid_PolygonRing
	ring  *Geom_LinearRing

	// The root of the touch graph tree containing this ring.
	// Serves as the id for the graph partition induced by the touch relation.
	touchSetRoot *OperationValid_PolygonRing

	// lazily created
	// The set of PolygonRingTouch links for this ring.
	// The set of all touches in the rings of a polygon forms the polygon touch graph.
	// This supports detecting touch cycles, which reveal the condition of a disconnected interior.
	// Only a single touch is recorded between any two rings,
	// since more than one touch between two rings indicates interior disconnection as well.
	touches map[int]*operationValid_PolygonRingTouch

	// The set of self-nodes in this ring.
	// This supports checking valid ring self-touch topology.
	selfNodes []*operationValid_PolygonRingSelfNode
}

// OperationValid_NewPolygonRing creates a ring for a polygon shell.
func OperationValid_NewPolygonRing(ring *Geom_LinearRing) *OperationValid_PolygonRing {
	pr := &OperationValid_PolygonRing{
		ring: ring,
		id:   -1,
	}
	pr.shell = pr
	return pr
}

// OperationValid_NewPolygonRingWithIndexAndShell creates a ring for a polygon hole.
func OperationValid_NewPolygonRingWithIndexAndShell(ring *Geom_LinearRing, index int, shell *OperationValid_PolygonRing) *OperationValid_PolygonRing {
	return &OperationValid_PolygonRing{
		ring:  ring,
		id:    index,
		shell: shell,
	}
}

func (pr *OperationValid_PolygonRing) IsSamePolygon(ring *OperationValid_PolygonRing) bool {
	return pr.shell == ring.shell
}

func (pr *OperationValid_PolygonRing) IsShell() bool {
	return pr.shell == pr
}

func (pr *OperationValid_PolygonRing) isInTouchSet() bool {
	return pr.touchSetRoot != nil
}

func (pr *OperationValid_PolygonRing) setTouchSetRoot(ring *OperationValid_PolygonRing) {
	pr.touchSetRoot = ring
}

func (pr *OperationValid_PolygonRing) getTouchSetRoot() *OperationValid_PolygonRing {
	return pr.touchSetRoot
}

func (pr *OperationValid_PolygonRing) hasTouches() bool {
	return pr.touches != nil && len(pr.touches) > 0
}

func (pr *OperationValid_PolygonRing) getTouches() []*operationValid_PolygonRingTouch {
	result := make([]*operationValid_PolygonRingTouch, 0, len(pr.touches))
	for _, touch := range pr.touches {
		result = append(result, touch)
	}
	return result
}

func (pr *OperationValid_PolygonRing) addTouch(ring *OperationValid_PolygonRing, pt *Geom_Coordinate) {
	if pr.touches == nil {
		pr.touches = make(map[int]*operationValid_PolygonRingTouch)
	}
	_, exists := pr.touches[ring.id]
	if !exists {
		pr.touches[ring.id] = operationValid_newPolygonRingTouch(ring, pt)
	}
}

func (pr *OperationValid_PolygonRing) AddSelfTouch(origin, e00, e01, e10, e11 *Geom_Coordinate) {
	if pr.selfNodes == nil {
		pr.selfNodes = make([]*operationValid_PolygonRingSelfNode, 0)
	}
	pr.selfNodes = append(pr.selfNodes, operationValid_newPolygonRingSelfNode(origin, e00, e01, e10, e11))
}

// isOnlyTouch tests if this ring touches a given ring at
// the single point specified.
func (pr *OperationValid_PolygonRing) isOnlyTouch(ring *OperationValid_PolygonRing, pt *Geom_Coordinate) bool {
	//--- no touches for this ring
	if pr.touches == nil {
		return true
	}
	//--- no touches for other ring
	touch, exists := pr.touches[ring.id]
	if !exists {
		return true
	}
	//--- the rings touch - check if point is the same
	return touch.isAtLocation(pt)
}

// findHoleCycleLocation detects whether the subgraph of holes linked by touch to this ring
// contains a hole cycle.
// If no cycles are detected, the set of touching rings is a tree.
// The set is marked using this ring as the root.
func (pr *OperationValid_PolygonRing) findHoleCycleLocation() *Geom_Coordinate {
	//--- the touch set including this ring is already processed
	if pr.isInTouchSet() {
		return nil
	}

	//--- scan the touch set tree rooted at this ring
	// Assert: this.touchSetRoot is null
	root := pr
	root.setTouchSetRoot(root)

	if !pr.hasTouches() {
		return nil
	}

	touchStack := make([]*operationValid_PolygonRingTouch, 0)
	touchStack = operationValid_polygonRing_init(root, touchStack)

	for len(touchStack) > 0 {
		// pop
		touch := touchStack[len(touchStack)-1]
		touchStack = touchStack[:len(touchStack)-1]

		holeCyclePt := pr.scanForHoleCycle(touch, root, &touchStack)
		if holeCyclePt != nil {
			return holeCyclePt
		}
	}
	return nil
}

func operationValid_polygonRing_init(root *OperationValid_PolygonRing, touchStack []*operationValid_PolygonRingTouch) []*operationValid_PolygonRingTouch {
	for _, touch := range root.getTouches() {
		touch.getRing().setTouchSetRoot(root)
		touchStack = append(touchStack, touch)
	}
	return touchStack
}

// scanForHoleCycle scans for a hole cycle starting at a given touch.
func (pr *OperationValid_PolygonRing) scanForHoleCycle(currentTouch *operationValid_PolygonRingTouch, root *OperationValid_PolygonRing, touchStack *[]*operationValid_PolygonRingTouch) *Geom_Coordinate {
	ring := currentTouch.getRing()
	currentPt := currentTouch.getCoordinate()

	// Scan the touched rings
	// Either they form a hole cycle, or they are added to the touch set
	// and pushed on the stack for scanning
	for _, touch := range ring.getTouches() {
		// Don't check touches at the entry point
		// to avoid trivial cycles.
		// They will already be processed or on the stack
		// from the previous ring (which touched
		// all the rings at that point as well)
		if currentPt.Equals2D(touch.getCoordinate()) {
			continue
		}

		// Test if the touched ring has already been
		// reached via a different touch path.
		// This is indicated by it already being marked as
		// part of the touch set.
		// This indicates a hole cycle has been found.
		touchRing := touch.getRing()
		if touchRing.getTouchSetRoot() == root {
			return touch.getCoordinate()
		}

		touchRing.setTouchSetRoot(root)

		*touchStack = append(*touchStack, touch)
	}
	return nil
}

// FindInteriorSelfNode finds the location of an invalid interior self-touch in this ring,
// if one exists.
func (pr *OperationValid_PolygonRing) FindInteriorSelfNode() *Geom_Coordinate {
	if pr.selfNodes == nil {
		return nil
	}

	// Determine if the ring interior is on the Right.
	// This is the case if the ring is a shell and is CW,
	// or is a hole and is CCW.
	isCCW := Algorithm_Orientation_IsCCW(pr.ring.GetCoordinates())
	isInteriorOnRight := pr.IsShell() != isCCW

	for _, selfNode := range pr.selfNodes {
		if !selfNode.isExterior(isInteriorOnRight) {
			return selfNode.getCoordinate()
		}
	}
	return nil
}

func (pr *OperationValid_PolygonRing) String() string {
	return pr.ring.String()
}

// operationValid_PolygonRingTouch records a point where a PolygonRing touches another one.
// This forms an edge in the induced ring touch graph.
type operationValid_PolygonRingTouch struct {
	ring    *OperationValid_PolygonRing
	touchPt *Geom_Coordinate
}

func operationValid_newPolygonRingTouch(ring *OperationValid_PolygonRing, pt *Geom_Coordinate) *operationValid_PolygonRingTouch {
	return &operationValid_PolygonRingTouch{
		ring:    ring,
		touchPt: pt,
	}
}

func (prt *operationValid_PolygonRingTouch) getCoordinate() *Geom_Coordinate {
	return prt.touchPt
}

func (prt *operationValid_PolygonRingTouch) getRing() *OperationValid_PolygonRing {
	return prt.ring
}

func (prt *operationValid_PolygonRingTouch) isAtLocation(pt *Geom_Coordinate) bool {
	return prt.touchPt.Equals2D(pt)
}

// operationValid_PolygonRingSelfNode represents a ring self-touch node, recording the node (intersection point)
// and the endpoints of the four adjacent segments.
// This is used to evaluate validity of self-touching nodes, when they are allowed.
type operationValid_PolygonRingSelfNode struct {
	nodePt *Geom_Coordinate
	e00    *Geom_Coordinate
	e01    *Geom_Coordinate
	e10    *Geom_Coordinate
	//e11    *Geom_Coordinate
}

func operationValid_newPolygonRingSelfNode(nodePt, e00, e01, e10, e11 *Geom_Coordinate) *operationValid_PolygonRingSelfNode {
	return &operationValid_PolygonRingSelfNode{
		nodePt: nodePt,
		e00:    e00,
		e01:    e01,
		e10:    e10,
		//e11:    e11,
	}
}

// getCoordinate returns the node point.
func (sn *operationValid_PolygonRingSelfNode) getCoordinate() *Geom_Coordinate {
	return sn.nodePt
}

// isExterior tests if a self-touch has the segments of each half of the touch
// lying in the exterior of a polygon.
// This is a valid self-touch.
// It applies to both shells and holes.
// Only one of the four possible cases needs to be tested,
// since the situation has full symmetry.
func (sn *operationValid_PolygonRingSelfNode) isExterior(isInteriorOnRight bool) bool {
	// Note that either corner and either of the other edges could be used to test.
	// The situation is fully symmetrical.
	isInteriorSeg := Algorithm_PolygonNodeTopology_IsInteriorSegment(sn.nodePt, sn.e00, sn.e01, sn.e10)
	isExterior := isInteriorOnRight != isInteriorSeg
	return isExterior
}
