package oracle

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"

	"github.com/peterstace/simplefeatures/internal/jts"
	"github.com/peterstace/simplefeatures/internal/jts/java"
)

// port is the Go side of the oracle. It mirrors Driver.java: the same
// commands produce replies in the same format.
type port struct {
	factory *jts.Geom_GeometryFactory
	geoms   map[string]*jts.Geom_Geometry
}

func newPort() *port {
	return &port{
		factory: jts.Geom_GeometryFactory_new(),
		geoms:   map[string]*jts.Geom_Geometry{},
	}
}

// setPrecisionModel mirrors the driver's pm command.
func (p *port) setPrecisionModel(fields []string) {
	var pm *jts.Geom_PrecisionModel
	switch fields[1] {
	case "FLOATING":
		pm = jts.Geom_PrecisionModel_new()
	case "FLOATING_SINGLE":
		pm = jts.Geom_PrecisionModel_new_Type(jts.Geom_PrecisionModel_FLOATING_SINGLE)
	case "FIXED":
		pm = jts.Geom_PrecisionModel_new_double(dbl(fields[2]))
	default:
		panic("precision model " + fields[1])
	}
	p.factory = jts.Geom_GeometryFactory_new_PrecisionModel(pm)
}

// load mirrors the driver's load command. It returns a non-empty reply only
// if loading failed.
func (p *port) load(name, hexWKB string) (failure string) {
	return capture(func() string {
		bytes := jts.Io_WKBReader_hexToBytes(java.FromGo(hexWKB))
		p.geoms[name] = jts.Io_WKBReader_new_GeometryFactory(p.factory).Read_byteArr(bytes)
		return ""
	})
}

func (p *port) g(name string) *jts.Geom_Geometry {
	g, ok := p.geoms[name]
	if !ok {
		panic("no geometry " + name)
	}
	return g
}

// op mirrors the driver's op command (without the identity hash prefix).
func (p *port) op(f []string) string {
	return capture(func() string { return "OK " + p.run(f) })
}

func (p *port) run(f []string) string {
	switch op := f[1]; op {
	case "relate":
		return jts.OperationRelateng_RelateNG_relate_Geometry_Geometry(p.g(f[2]), p.g(f[3])).ToString().Go()
	case "relatePredicate":
		return strconv.FormatBool(jts.OperationRelateng_RelateNG_relate_Geometry_Geometry_TopologyPredicate(
			p.g(f[2]), p.g(f[3]), predicate(f[4])))
	case "prepared":
		return strconv.FormatBool(prepared(jts.GeomPrep_PreparedGeometryFactory_prepare(p.g(f[2])), p.g(f[3]), f[4]))
	case "overlay":
		return dump(jts.OperationOverlayng_OverlayNGRobust_overlay(p.g(f[2]), p.g(f[3]), atoi(f[4])))
	case "unaryUnion":
		return dump(jts.OperationOverlayng_OverlayNGRobust_union_Geometry(p.g(f[2])))
	case "buffer":
		params := jts.OperationBuffer_BufferParameters_new_int_int_int_double(atoi(f[4]), atoi(f[5]), atoi(f[6]), dbl(f[7]))
		return dump(jts.OperationBuffer_BufferOp_bufferOp_Geometry_double_BufferParameters(p.g(f[2]), dbl(f[3]), params))
	case "isValid":
		return strconv.FormatBool(jts.OperationValid_IsValidOp_isValid_Geometry(p.g(f[2])))
	case "isSimple":
		return strconv.FormatBool(jts.OperationValid_IsSimpleOp_isSimple(p.g(f[2])))
	case "area":
		return bits(p.g(f[2]).GetArea())
	case "length":
		return bits(p.g(f[2]).GetLength())
	case "centroid":
		return dump(geom(p.g(f[2]).GetCentroid()))
	case "interiorPoint":
		return dump(geom(p.g(f[2]).GetInteriorPoint()))
	case "convexHull":
		return dump(jts.Algorithm_ConvexHull_new_Geometry(p.g(f[2])).GetConvexHull())
	case "boundary":
		return dump(p.g(f[2]).GetBoundary())
	case "envelope":
		return dump(p.g(f[2]).GetEnvelope())
	case "normalize":
		c := p.g(f[2]).Copy()
		c.Normalize()
		return dump(c)
	case "distance":
		return bits(jts.OperationDistance_DistanceOp_distance(p.g(f[2]), p.g(f[3])))
	case "hausdorff":
		return bits(jts.AlgorithmDistance_DiscreteHausdorffDistance_distance_Geometry_Geometry(p.g(f[2]), p.g(f[3])))
	case "simplifyDP":
		return dump(jts.Simplify_DouglasPeuckerSimplifier_simplify(p.g(f[2]), dbl(f[3])))
	case "simplifyTP":
		return dump(jts.Simplify_TopologyPreservingSimplifier_simplify(p.g(f[2]), dbl(f[3])))
	case "dump":
		return dump(p.g(f[2]))
	default:
		panic("unknown op " + op)
	}
}

func predicate(name string) jts.OperationRelateng_TopologyPredicate {
	switch name {
	case "intersects":
		return jts.OperationRelateng_RelatePredicate_intersects()
	case "contains":
		return jts.OperationRelateng_RelatePredicate_contains()
	case "within":
		return jts.OperationRelateng_RelatePredicate_within()
	case "covers":
		return jts.OperationRelateng_RelatePredicate_covers()
	case "coveredBy":
		return jts.OperationRelateng_RelatePredicate_coveredBy()
	case "crosses":
		return jts.OperationRelateng_RelatePredicate_crosses()
	case "disjoint":
		return jts.OperationRelateng_RelatePredicate_disjoint()
	case "equalsTopo":
		return jts.OperationRelateng_RelatePredicate_equalsTopo()
	case "overlaps":
		return jts.OperationRelateng_RelatePredicate_overlaps()
	case "touches":
		return jts.OperationRelateng_RelatePredicate_touches()
	}
	panic("predicate " + name)
}

func prepared(pg jts.GeomPrep_PreparedGeometry, g *jts.Geom_Geometry, name string) bool {
	switch name {
	case "intersects":
		return pg.Intersects_Geometry(g)
	case "contains":
		return pg.Contains_Geometry(g)
	case "containsProperly":
		return pg.ContainsProperly_Geometry(g)
	case "covers":
		return pg.Covers_Geometry(g)
	case "coveredBy":
		return pg.CoveredBy_Geometry(g)
	case "crosses":
		return pg.Crosses_Geometry(g)
	case "disjoint":
		return pg.Disjoint_Geometry(g)
	case "overlaps":
		return pg.Overlaps_Geometry(g)
	case "touches":
		return pg.Touches_Geometry(g)
	case "within":
		return pg.Within_Geometry(g)
	}
	panic("prepared predicate " + name)
}

// capture runs f, turning a panic into the reply the driver gives for a
// thrown exception ("EXC <class>"), or "UNTRANSLATED <reason>" when the port
// reached code the translator could not translate.
func capture(f func() string) (reply string) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if u, ok := r.(java.UnsupportedError); ok {
			reply = "UNTRANSLATED " + u.Reason
			return
		}
		reply = func() (reply string) {
			defer func() {
				if r2 := recover(); r2 != nil {
					reply = fmt.Sprintf("PANIC %v", r)
				}
			}()
			return "EXC " + className(java.Throwable_ofPanic(r))
		}()
	}()
	return f()
}

func className(o any) string {
	return o.(interface{ JClassName() string }).JClassName()
}

// geom upcasts a pointer to a generated geometry class to *Geom_Geometry.
// Every part of a generated object starts at the object's address.
func geom[T any](p *T) *jts.Geom_Geometry {
	return (*jts.Geom_Geometry)(unsafe.Pointer(p))
}

// dump mirrors Driver.dump.
func dump(g *jts.Geom_Geometry) string {
	var sb strings.Builder
	dumpGeom(g, &sb)
	return sb.String()
}

func dumpGeom(g *jts.Geom_Geometry, sb *strings.Builder) {
	if g == nil {
		sb.WriteString("null")
		return
	}
	sb.WriteString(simpleName(className(jts.Geom_Geometry_toAny(g))))
	sb.WriteByte('{')
	switch {
	case jts.Geom_GeometryCollection_isKind(g):
		for i := int32(0); i < g.GetNumGeometries(); i++ {
			if i > 0 {
				sb.WriteByte(',')
			}
			dumpGeom(g.GetGeometryN_int(i), sb)
		}
	case jts.Geom_Polygon_isKind(g):
		p := jts.Geom_Polygon_downcast(g)
		dumpGeom(geom(p.GetExteriorRing()), sb)
		for i := int32(0); i < p.GetNumInteriorRing(); i++ {
			sb.WriteByte(',')
			dumpGeom(geom(p.GetInteriorRingN_int(i)), sb)
		}
	case jts.Geom_LineString_isKind(g):
		dumpSeq(jts.Geom_LineString_downcast(g).GetCoordinateSequence(), sb)
	case jts.Geom_Point_isKind(g):
		dumpSeq(jts.Geom_Point_downcast(g).GetCoordinateSequence(), sb)
	}
	sb.WriteByte('}')
}

func dumpSeq(seq jts.Geom_CoordinateSequence, sb *strings.Builder) {
	fmt.Fprintf(sb, "%s/%d/%d[", simpleName(className(seq)), seq.GetDimension(), seq.GetMeasures())
	for i := int32(0); i < seq.Size(); i++ {
		if i > 0 {
			sb.WriteByte(';')
		}
		c := seq.GetCoordinate_int(i)
		fmt.Fprintf(sb, "%s:%s,%s,%s,%s", simpleName(className(jts.Geom_Coordinate_toAny(c))),
			bits(c.GetX()), bits(c.GetY()), bits(c.GetZ()), bits(c.GetM()))
	}
	sb.WriteByte(']')
}

func simpleName(binary string) string {
	return binary[strings.LastIndexAny(binary, ".$")+1:]
}

func bits(d float64) string {
	return strconv.FormatUint(math.Float64bits(d), 16)
}

func dbl(hexBits string) float64 {
	u, err := strconv.ParseUint(hexBits, 16, 64)
	if err != nil {
		panic(err)
	}
	return math.Float64frombits(u)
}

func atoi(s string) int32 {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return int32(i)
}
