package oracle_test

import (
	"sync"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jts"
)

// TestPreparedGeometryConcurrent checks that a prepared geometry can be used
// from many goroutines at once, as JTS's synchronized lazy index building
// allows. Run with -race.
func TestPreparedGeometryConcurrent(t *testing.T) {
	poly := portPolygon(wavy(0, 0, 500))
	f := jts.Geom_GeometryFactory_new()
	var pts []*jts.Geom_Geometry
	for i := 0; i < 200; i++ {
		x, y := float64(i%20)/10-1, float64(i/20)/5-1
		pts = append(pts, geom(f.CreatePoint_Coordinate(jts.Geom_Coordinate_new_double_double(x, y))))
	}
	want := make([]bool, len(pts))
	for i, p := range pts {
		want[i] = jts.GeomPrep_PreparedGeometryFactory_prepare(poly).Contains_Geometry(p)
	}
	for round := 0; round < 20; round++ {
		pg := jts.GeomPrep_PreparedGeometryFactory_prepare(poly)
		var wg sync.WaitGroup
		for g := 0; g < 20; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i, p := range pts {
					if got := pg.Contains_Geometry(p); got != want[i] {
						t.Errorf("point %d: got %v want %v", i, got, want[i])
					}
					pg.Intersects_Geometry(p)
				}
			}()
		}
		wg.Wait()
	}
}
