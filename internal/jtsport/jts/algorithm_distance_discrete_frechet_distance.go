package jts

import (
	"math"
	"sort"
)

// AlgorithmDistance_DiscreteFrechetDistance computes the Discrete Fréchet Distance
// between two geometries.
//
// The Fréchet distance is a measure of similarity between curves. Thus, it can
// be used like the Hausdorff distance.
//
// An analogy for the Fréchet distance taken from
// Computing Discrete Fréchet Distance:
// A man is walking a dog on a leash: the man can move
// on one curve, the dog on the other; both may vary their
// speed, but backtracking is not allowed.
//
// Its metric is better than the Hausdorff distance
// because it takes the directions of the curves into account.
// It is possible that two curves have a small Hausdorff but a large
// Fréchet distance.
//
// This implementation is based on the following optimized Fréchet distance algorithm:
// Thomas Devogele, Maxence Esnault, Laurent Etienne. Distance discrète de Fréchet optimisée. Spatial
// Analysis and Geomatics (SAGEO), Nov 2016, Nice, France. hal-02110055
//
// Several matrix storage implementations are provided.
type AlgorithmDistance_DiscreteFrechetDistance struct {
	g0     *Geom_Geometry
	g1     *Geom_Geometry
	ptDist *AlgorithmDistance_PointPairDistance
}

// AlgorithmDistance_DiscreteFrechetDistance_Distance computes the Discrete Fréchet Distance between two Geometries
// using a Cartesian distance computation function.
func AlgorithmDistance_DiscreteFrechetDistance_Distance(g0, g1 *Geom_Geometry) float64 {
	dist := AlgorithmDistance_NewDiscreteFrechetDistance(g0, g1)
	return dist.distance()
}

// AlgorithmDistance_NewDiscreteFrechetDistance creates an instance of this class using the provided geometries.
func AlgorithmDistance_NewDiscreteFrechetDistance(g0, g1 *Geom_Geometry) *AlgorithmDistance_DiscreteFrechetDistance {
	return &AlgorithmDistance_DiscreteFrechetDistance{
		g0: g0,
		g1: g1,
	}
}

// distance computes the Discrete Fréchet Distance between the input geometries.
func (dfd *AlgorithmDistance_DiscreteFrechetDistance) distance() float64 {
	coords0 := dfd.g0.GetCoordinates()
	coords1 := dfd.g1.GetCoordinates()

	distances := algorithmDistance_createMatrixStorage(len(coords0), len(coords1))
	diagonal := algorithmDistance_bresenhamDiagonal(len(coords0), len(coords1))

	distanceToPair := make(map[float64][]int)
	dfd.computeCoordinateDistances(coords0, coords1, diagonal, distances, distanceToPair)
	dfd.ptDist = algorithmDistance_computeFrechet(coords0, coords1, diagonal, distances, distanceToPair)

	return dfd.ptDist.GetDistance()
}

// algorithmDistance_createMatrixStorage creates a matrix to store the computed distances.
func algorithmDistance_createMatrixStorage(rows, cols int) algorithmDistance_matrixStorage {
	max := rows
	if cols > max {
		max = cols
	}
	// NOTE: these constraints need to be verified
	if max < 1024 {
		return algorithmDistance_newRectMatrix(rows, cols, math.Inf(1))
	}

	return algorithmDistance_newCsrMatrix(rows, cols, math.Inf(1))
}

// GetCoordinates gets the pair of Coordinates at which the distance is obtained.
func (dfd *AlgorithmDistance_DiscreteFrechetDistance) GetCoordinates() []*Geom_Coordinate {
	if dfd.ptDist == nil {
		dfd.distance()
	}

	return dfd.ptDist.GetCoordinates()
}

// algorithmDistance_computeFrechet computes the Fréchet Distance for the given distance matrix.
func algorithmDistance_computeFrechet(coords0, coords1 []*Geom_Coordinate, diagonal []int,
	distances algorithmDistance_matrixStorage, distanceToPair map[float64][]int) *AlgorithmDistance_PointPairDistance {
	for d := 0; d < len(diagonal); d += 2 {
		i0 := diagonal[d]
		j0 := diagonal[d+1]

		for i := i0; i < len(coords0); i++ {
			if distances.isValueSet(i, j0) {
				dist := algorithmDistance_getMinDistanceAtCorner(distances, i, j0)
				if dist > distances.get(i, j0) {
					distances.set(i, j0, dist)
				}
			} else {
				break
			}
		}
		for j := j0 + 1; j < len(coords1); j++ {
			if distances.isValueSet(i0, j) {
				dist := algorithmDistance_getMinDistanceAtCorner(distances, i0, j)
				if dist > distances.get(i0, j) {
					distances.set(i0, j, dist)
				}
			} else {
				break
			}
		}
	}

	result := AlgorithmDistance_NewPointPairDistance()
	distance := distances.get(len(coords0)-1, len(coords1)-1)
	index := distanceToPair[distance]
	if index == nil {
		panic("Pair of points not recorded for computed distance")
	}
	result.InitializeWithCoordinatesAndDistance(coords0[index[0]], coords1[index[1]], distance)
	return result
}

// algorithmDistance_getMinDistanceAtCorner returns the minimum distance at the corner (i, j).
func algorithmDistance_getMinDistanceAtCorner(matrix algorithmDistance_matrixStorage, i, j int) float64 {
	if i > 0 && j > 0 {
		d0 := matrix.get(i-1, j-1)
		d1 := matrix.get(i-1, j)
		d2 := matrix.get(i, j-1)
		return math.Min(math.Min(d0, d1), d2)
	}
	if i == 0 && j == 0 {
		return matrix.get(0, 0)
	}

	if i == 0 {
		return matrix.get(0, j-1)
	}

	// j == 0
	return matrix.get(i-1, 0)
}

// computeCoordinateDistances computes relevant distances between pairs of Coordinates for the
// computation of the Discrete Fréchet Distance.
func (dfd *AlgorithmDistance_DiscreteFrechetDistance) computeCoordinateDistances(coords0, coords1 []*Geom_Coordinate, diagonal []int,
	distances algorithmDistance_matrixStorage, distanceToPair map[float64][]int) {
	numDiag := len(diagonal)
	maxDistOnDiag := 0.0
	imin, jmin := 0, 0
	numCoords0 := len(coords0)
	numCoords1 := len(coords1)

	// First compute all the distances along the diagonal.
	// Record the maximum distance.

	for k := 0; k < numDiag; k += 2 {
		i0 := diagonal[k]
		j0 := diagonal[k+1]
		diagDist := coords0[i0].Distance(coords1[j0])
		if diagDist > maxDistOnDiag {
			maxDistOnDiag = diagDist
		}
		distances.set(i0, j0, diagDist)
		if _, exists := distanceToPair[diagDist]; !exists {
			distanceToPair[diagDist] = []int{i0, j0}
		}
	}

	// Check for distances shorter than maxDistOnDiag along the diagonal
	for k := 0; k < numDiag-2; k += 2 {
		// Decode index
		i0 := diagonal[k]
		j0 := diagonal[k+1]

		// Get reference coordinates for col and row
		coord0 := coords0[i0]
		coord1 := coords1[j0]

		// Check for shorter distances in this row
		i := i0 + 1
		for ; i < numCoords0; i++ {
			if !distances.isValueSet(i, j0) {
				dist := coords0[i].Distance(coord1)
				if dist < maxDistOnDiag || i < imin {
					distances.set(i, j0, dist)
					if _, exists := distanceToPair[dist]; !exists {
						distanceToPair[dist] = []int{i, j0}
					}
				} else {
					break
				}
			} else {
				break
			}
		}
		imin = i

		// Check for shorter distances in this column
		j := j0 + 1
		for ; j < numCoords1; j++ {
			if !distances.isValueSet(i0, j) {
				dist := coord0.Distance(coords1[j])
				if dist < maxDistOnDiag || j < jmin {
					distances.set(i0, j, dist)
					if _, exists := distanceToPair[dist]; !exists {
						distanceToPair[dist] = []int{i0, j}
					}
				} else {
					break
				}
			} else {
				break
			}
		}
		jmin = j
	}

	//System.out.println(distances.toString());
}

// algorithmDistance_bresenhamDiagonal computes the indices for the diagonal of a numCols x numRows grid
// using the Bresenham line algorithm.
func algorithmDistance_bresenhamDiagonal(numCols, numRows int) []int {
	dim := numCols
	if numRows > dim {
		dim = numRows
	}
	diagXY := make([]int, 2*dim)

	dx := numCols - 1
	dy := numRows - 1
	var err int
	i := 0
	if numCols > numRows {
		y := 0
		err = 2*dy - dx
		for x := 0; x < numCols; x++ {
			diagXY[i] = x
			i++
			diagXY[i] = y
			i++
			if err > 0 {
				y += 1
				err -= 2 * dx
			}
			err += 2 * dy
		}
	} else {
		x := 0
		err = 2*dx - dy
		for y := 0; y < numRows; y++ {
			diagXY[i] = x
			i++
			diagXY[i] = y
			i++
			if err > 0 {
				x += 1
				err -= 2 * dy
			}
			err += 2 * dx
		}
	}
	return diagXY
}

// algorithmDistance_matrixStorage is an abstract base class for storing 2d matrix data.
type algorithmDistance_matrixStorage interface {
	get(i, j int) float64
	set(i, j int, value float64)
	isValueSet(i, j int) bool
}

// algorithmDistance_rectMatrix is a straightforward implementation of a rectangular matrix.
type algorithmDistance_rectMatrix struct {
	numRows      int
	numCols      int
	defaultValue float64
	matrix       []float64
}

// algorithmDistance_newRectMatrix creates an instance of this matrix using the given number of rows and columns.
// A default value can be specified.
func algorithmDistance_newRectMatrix(numRows, numCols int, defaultValue float64) *algorithmDistance_rectMatrix {
	matrix := make([]float64, numRows*numCols)
	for i := range matrix {
		matrix[i] = defaultValue
	}
	return &algorithmDistance_rectMatrix{
		numRows:      numRows,
		numCols:      numCols,
		defaultValue: defaultValue,
		matrix:       matrix,
	}
}

func (m *algorithmDistance_rectMatrix) get(i, j int) float64 {
	return m.matrix[i*m.numCols+j]
}

func (m *algorithmDistance_rectMatrix) set(i, j int, value float64) {
	m.matrix[i*m.numCols+j] = value
}

func (m *algorithmDistance_rectMatrix) isValueSet(i, j int) bool {
	return math.Float64bits(m.get(i, j)) != math.Float64bits(m.defaultValue)
}

// algorithmDistance_csrMatrix is a matrix implementation that adheres to the
// Compressed sparse row format.
// Note: Unfortunately not as fast as expected.
type algorithmDistance_csrMatrix struct {
	numRows      int
	numCols      int
	defaultValue float64
	v            []float64
	ri           []int
	ci           []int
}

func algorithmDistance_newCsrMatrix(numRows, numCols int, defaultValue float64) *algorithmDistance_csrMatrix {
	return algorithmDistance_newCsrMatrixWithExpectedValues(numRows, numCols, defaultValue, algorithmDistance_expectedValuesHeuristic(numRows, numCols))
}

func algorithmDistance_newCsrMatrixWithExpectedValues(numRows, numCols int, defaultValue float64, expectedValues int) *algorithmDistance_csrMatrix {
	return &algorithmDistance_csrMatrix{
		numRows:      numRows,
		numCols:      numCols,
		defaultValue: defaultValue,
		v:            make([]float64, expectedValues),
		ci:           make([]int, expectedValues),
		ri:           make([]int, numRows+1),
	}
}

// algorithmDistance_expectedValuesHeuristic computes an initial value for the number of expected values.
func algorithmDistance_expectedValuesHeuristic(numRows, numCols int) int {
	max := numRows
	if numCols > max {
		max = numCols
	}
	return max * max / 10
}

func (m *algorithmDistance_csrMatrix) indexOf(i, j int) int {
	cLow := m.ri[i]
	cHigh := m.ri[i+1]
	if cHigh <= cLow {
		return ^cLow
	}

	idx := sort.SearchInts(m.ci[cLow:cHigh], j)
	if idx < cHigh-cLow && m.ci[cLow+idx] == j {
		return cLow + idx
	}
	return ^(cLow + idx)
}

func (m *algorithmDistance_csrMatrix) get(i, j int) float64 {
	// get the index in the vector
	vi := m.indexOf(i, j)

	// if the vector index is negative, return default value
	if vi < 0 {
		return m.defaultValue
	}

	return m.v[vi]
}

func (m *algorithmDistance_csrMatrix) set(i, j int, value float64) {
	// get the index in the vector
	vi := m.indexOf(i, j)

	// do we already have a value?
	if vi < 0 {
		// no, we don't, we need to ensure space!
		m.ensureCapacity(m.ri[m.numRows] + 1)

		// update row indices
		for ii := i + 1; ii <= m.numRows; ii++ {
			m.ri[ii] += 1
		}

		// move and update column indices, move values
		vi = ^vi
		for ii := m.ri[m.numRows]; ii > vi; ii-- {
			m.ci[ii] = m.ci[ii-1]
			m.v[ii] = m.v[ii-1]
		}

		// insert column index
		m.ci[vi] = j
	}

	// set the new value
	m.v[vi] = value
}

func (m *algorithmDistance_csrMatrix) isValueSet(i, j int) bool {
	return m.indexOf(i, j) >= 0
}

// ensureCapacity ensures that the column index vector (ci) and value vector (v) are sufficiently large.
func (m *algorithmDistance_csrMatrix) ensureCapacity(required int) {
	if required < len(m.v) {
		return
	}

	increment := m.numRows
	if m.numCols > increment {
		increment = m.numCols
	}
	newV := make([]float64, len(m.v)+increment)
	copy(newV, m.v)
	m.v = newV
	newCi := make([]int, len(m.v)+increment)
	copy(newCi, m.ci)
	m.ci = newCi
}

// algorithmDistance_hashMapMatrix is a sparse matrix based on Go's map.
type algorithmDistance_hashMapMatrix struct {
	numRows      int
	numCols      int
	defaultValue float64
	matrix       map[int64]float64
}

// algorithmDistance_newHashMapMatrix creates an instance of this class.
func algorithmDistance_newHashMapMatrix(numRows, numCols int, defaultValue float64) *algorithmDistance_hashMapMatrix {
	return &algorithmDistance_hashMapMatrix{
		numRows:      numRows,
		numCols:      numCols,
		defaultValue: defaultValue,
		matrix:       make(map[int64]float64),
	}
}

func (m *algorithmDistance_hashMapMatrix) get(i, j int) float64 {
	key := int64(i)<<32 | int64(j)
	if v, ok := m.matrix[key]; ok {
		return v
	}
	return m.defaultValue
}

func (m *algorithmDistance_hashMapMatrix) set(i, j int, value float64) {
	key := int64(i)<<32 | int64(j)
	m.matrix[key] = value
}

func (m *algorithmDistance_hashMapMatrix) isValueSet(i, j int) bool {
	key := int64(i)<<32 | int64(j)
	_, ok := m.matrix[key]
	return ok
}
