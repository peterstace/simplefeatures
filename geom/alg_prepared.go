package geom

import "github.com/peterstace/simplefeatures/internal/jtsport/jts"

// PreparedGeometry is a geometry that has been preprocessed to efficiently
// enable evaluation of spatial predicates against many other geometries.
//
// It is created by calling [Prepare] with the geometry to be prepared. The
// prepared geometry caches spatial indices and other structures so that
// repeated predicate evaluations (e.g. [PreparedGeometry.Intersects],
// [PreparedGeometry.Contains]) against different test geometries are fast.
type PreparedGeometry struct {
	prep jts.GeomPrep_PreparedGeometry
}

// Prepare preprocesses a geometry for efficient repeated predicate evaluation.
func Prepare(g Geometry) (PreparedGeometry, error) {
	return catch(func() (PreparedGeometry, error) {
		jtsG, err := toJTS(g)
		if err != nil {
			return PreparedGeometry{}, err
		}
		return PreparedGeometry{
			prep: jts.GeomPrep_PreparedGeometryFactory_Prepare(jtsG),
		}, nil
	})
}

func toJTS(g Geometry) (*jts.Geom_Geometry, error) {
	jtsG, err := jts.Io_NewWKBReader().ReadBytes(g.AsBinary())
	if err != nil {
		return nil, wrap(err, "converting geometry to JTS")
	}
	return jtsG, nil
}

func (p PreparedGeometry) eval(g Geometry, pred func(*jts.Geom_Geometry) bool) (bool, error) {
	return catch(func() (bool, error) {
		jtsG, err := toJTS(g)
		if err != nil {
			return false, err
		}
		return pred(jtsG), nil
	})
}

// Intersects reports whether the prepared geometry intersects g.
func (p PreparedGeometry) Intersects(g Geometry) (bool, error) {
	return p.eval(g, p.prep.Intersects)
}

// Contains reports whether the prepared geometry contains g.
func (p PreparedGeometry) Contains(g Geometry) (bool, error) {
	return p.eval(g, p.prep.Contains)
}

// ContainsProperly reports whether the prepared geometry properly contains g.
// A geometry properly contains another if it contains it and the other
// geometry has no points on the boundary of the prepared geometry.
func (p PreparedGeometry) ContainsProperly(g Geometry) (bool, error) {
	return p.eval(g, p.prep.ContainsProperly)
}

// CoveredBy reports whether the prepared geometry is covered by g.
func (p PreparedGeometry) CoveredBy(g Geometry) (bool, error) {
	return p.eval(g, p.prep.CoveredBy)
}

// Covers reports whether the prepared geometry covers g.
func (p PreparedGeometry) Covers(g Geometry) (bool, error) {
	return p.eval(g, p.prep.Covers)
}

// Disjoint reports whether the prepared geometry is disjoint from g.
func (p PreparedGeometry) Disjoint(g Geometry) (bool, error) {
	return p.eval(g, p.prep.Disjoint)
}

// Overlaps reports whether the prepared geometry overlaps g.
func (p PreparedGeometry) Overlaps(g Geometry) (bool, error) {
	return p.eval(g, p.prep.Overlaps)
}

// Touches reports whether the prepared geometry touches g.
func (p PreparedGeometry) Touches(g Geometry) (bool, error) {
	return p.eval(g, p.prep.Touches)
}

// Within reports whether the prepared geometry is within g.
func (p PreparedGeometry) Within(g Geometry) (bool, error) {
	return p.eval(g, p.prep.Within)
}
