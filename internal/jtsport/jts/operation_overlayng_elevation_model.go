package jts

import "math"

const operationOverlayng_ElevationModel_DEFAULT_CELL_NUM = 3

// OperationOverlayng_ElevationModel is a simple elevation model used to
// populate missing Z values in overlay results.
//
// The model divides the extent of the input geometry(s) into an NxM grid. The
// default grid size is 3x3. If the input has no extent in the X or Y dimension,
// that dimension is given grid size 1. The elevation of each grid cell is
// computed as the average of the Z values of the input vertices in that cell
// (if any). If a cell has no input vertices within it, it is assigned the
// average elevation over all cells.
//
// If no input vertices have Z values, the model does not assign a Z value.
//
// The elevation of an arbitrary location is determined as the Z value of the
// nearest grid cell.
type OperationOverlayng_ElevationModel struct {
	extent        *Geom_Envelope
	numCellX      int
	numCellY      int
	cellSizeX     float64
	cellSizeY     float64
	cells         [][]*operationOverlayng_ElevationCell
	isInitialized bool
	hasZValue     bool
	averageZ      float64
}

// OperationOverlayng_ElevationModel_Create creates an elevation model from two
// geometries (which may be nil).
func OperationOverlayng_ElevationModel_Create(geom1, geom2 *Geom_Geometry) *OperationOverlayng_ElevationModel {
	extent := geom1.GetEnvelopeInternal().Copy()
	if geom2 != nil {
		extent.ExpandToIncludeEnvelope(geom2.GetEnvelopeInternal())
	}
	model := OperationOverlayng_NewElevationModel(extent, operationOverlayng_ElevationModel_DEFAULT_CELL_NUM, operationOverlayng_ElevationModel_DEFAULT_CELL_NUM)
	if geom1 != nil {
		model.Add(geom1)
	}
	if geom2 != nil {
		model.Add(geom2)
	}
	return model
}

// OperationOverlayng_NewElevationModel creates a new elevation model covering
// an extent by a grid of given dimensions.
func OperationOverlayng_NewElevationModel(extent *Geom_Envelope, numCellX, numCellY int) *OperationOverlayng_ElevationModel {
	em := &OperationOverlayng_ElevationModel{
		extent:    extent,
		numCellX:  numCellX,
		numCellY:  numCellY,
		cellSizeX: extent.GetWidth() / float64(numCellX),
		cellSizeY: extent.GetHeight() / float64(numCellY),
		averageZ:  math.NaN(),
	}

	if em.cellSizeX <= 0.0 {
		em.numCellX = 1
	}
	if em.cellSizeY <= 0.0 {
		em.numCellY = 1
	}
	em.cells = make([][]*operationOverlayng_ElevationCell, em.numCellX)
	for i := range em.cells {
		em.cells[i] = make([]*operationOverlayng_ElevationCell, em.numCellY)
	}
	return em
}

// Add updates the model using the Z values of a given geometry.
func (em *OperationOverlayng_ElevationModel) Add(geom *Geom_Geometry) {
	hasZ := true
	filter := newElevationModelCSFilter(em, &hasZ)
	geom.ApplyCoordinateSequenceFilter(filter)
}

type elevationModelCSFilter struct {
	em   *OperationOverlayng_ElevationModel
	hasZ *bool
}

var _ Geom_CoordinateSequenceFilter = (*elevationModelCSFilter)(nil)

func (f *elevationModelCSFilter) IsGeom_CoordinateSequenceFilter() {}

func newElevationModelCSFilter(em *OperationOverlayng_ElevationModel, hasZ *bool) *elevationModelCSFilter {
	return &elevationModelCSFilter{
		em:   em,
		hasZ: hasZ,
	}
}

func (f *elevationModelCSFilter) Filter(seq Geom_CoordinateSequence, i int) {
	if !seq.HasZ() {
		*f.hasZ = false
		return
	}
	z := seq.GetOrdinate(i, Geom_Coordinate_Z)
	f.em.add(seq.GetOrdinate(i, Geom_Coordinate_X), seq.GetOrdinate(i, Geom_Coordinate_Y), z)
}

func (f *elevationModelCSFilter) IsDone() bool {
	return !*f.hasZ
}

func (f *elevationModelCSFilter) IsGeometryChanged() bool {
	return false
}

func (em *OperationOverlayng_ElevationModel) add(x, y, z float64) {
	if math.IsNaN(z) {
		return
	}
	em.hasZValue = true
	cell := em.getCell(x, y, true)
	cell.Add(z)
}

func (em *OperationOverlayng_ElevationModel) init() {
	em.isInitialized = true
	numCells := 0
	sumZ := 0.0

	for i := 0; i < len(em.cells); i++ {
		for j := 0; j < len(em.cells[0]); j++ {
			cell := em.cells[i][j]
			if cell != nil {
				cell.Compute()
				numCells++
				sumZ += cell.GetZ()
			}
		}
	}
	em.averageZ = math.NaN()
	if numCells > 0 {
		em.averageZ = sumZ / float64(numCells)
	}
}

// GetZ gets the model Z value at a given location.
func (em *OperationOverlayng_ElevationModel) GetZ(x, y float64) float64 {
	if !em.isInitialized {
		em.init()
	}
	cell := em.getCell(x, y, false)
	if cell == nil {
		return em.averageZ
	}
	return cell.GetZ()
}

// PopulateZ computes Z values for any missing Z values in a geometry, using
// the computed model.
func (em *OperationOverlayng_ElevationModel) PopulateZ(geom *Geom_Geometry) {
	// Short-circuit if no Zs are present in model.
	if !em.hasZValue {
		return
	}

	if !em.isInitialized {
		em.init()
	}

	isDone := false
	filter := newElevationModelPopulateFilter(em, &isDone)
	geom.ApplyCoordinateSequenceFilter(filter)
}

type elevationModelPopulateFilter struct {
	em     *OperationOverlayng_ElevationModel
	isDone *bool
}

var _ Geom_CoordinateSequenceFilter = (*elevationModelPopulateFilter)(nil)

func (f *elevationModelPopulateFilter) IsGeom_CoordinateSequenceFilter() {}

func newElevationModelPopulateFilter(em *OperationOverlayng_ElevationModel, isDone *bool) *elevationModelPopulateFilter {
	return &elevationModelPopulateFilter{
		em:     em,
		isDone: isDone,
	}
}

func (f *elevationModelPopulateFilter) Filter(seq Geom_CoordinateSequence, i int) {
	if !seq.HasZ() {
		// If no Z then short-circuit evaluation.
		*f.isDone = true
		return
	}
	// If Z not populated then assign using model.
	if math.IsNaN(seq.GetZ(i)) {
		z := f.em.GetZ(seq.GetOrdinate(i, Geom_Coordinate_X), seq.GetOrdinate(i, Geom_Coordinate_Y))
		seq.SetOrdinate(i, Geom_Coordinate_Z, z)
	}
}

func (f *elevationModelPopulateFilter) IsDone() bool {
	return *f.isDone
}

func (f *elevationModelPopulateFilter) IsGeometryChanged() bool {
	return false
}

func (em *OperationOverlayng_ElevationModel) getCell(x, y float64, isCreateIfMissing bool) *operationOverlayng_ElevationCell {
	ix := 0
	if em.numCellX > 1 {
		ix = int((x - em.extent.GetMinX()) / em.cellSizeX)
		ix = Math_MathUtil_ClampInt(ix, 0, em.numCellX-1)
	}
	iy := 0
	if em.numCellY > 1 {
		iy = int((y - em.extent.GetMinY()) / em.cellSizeY)
		iy = Math_MathUtil_ClampInt(iy, 0, em.numCellY-1)
	}
	cell := em.cells[ix][iy]
	if isCreateIfMissing && cell == nil {
		cell = newElevationCell()
		em.cells[ix][iy] = cell
	}
	return cell
}

// operationOverlayng_ElevationCell is a cell for accumulating Z values.
type operationOverlayng_ElevationCell struct {
	numZ int
	sumZ float64
	avgZ float64
}

func newElevationCell() *operationOverlayng_ElevationCell {
	return &operationOverlayng_ElevationCell{}
}

func (c *operationOverlayng_ElevationCell) Add(z float64) {
	c.numZ++
	c.sumZ += z
}

func (c *operationOverlayng_ElevationCell) Compute() {
	c.avgZ = math.NaN()
	if c.numZ > 0 {
		c.avgZ = c.sumZ / float64(c.numZ)
	}
}

func (c *operationOverlayng_ElevationCell) GetZ() float64 {
	return c.avgZ
}
