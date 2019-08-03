package main

import (
	"database/sql"
	"sort"
	"strings"

	"github.com/peterstace/simplefeatures/geom"
)

type corpus struct {
	engines    [2]GeometryEngine
	candidates []string
	geometries []geom.Geometry
}

func newCorpus(db *sql.DB, candidates []string) *corpus {
	strSet := map[string]struct{}{}
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		strSet[c] = struct{}{}
	}
	corp := &corpus{engines: [2]GeometryEngine{SimpleFeaturesEngine{}, &PostgisEngine{db}}}
	for c := range strSet {
		corp.candidates = append(corp.candidates, c)
	}
	sort.Strings(corp.candidates)
	return corp
}

//func parseGeometry(candidate string) geom.Geometry {
//g, err := geom.UnmarshalWKT(strings.NewReader(candidate))
//if err == nil {
//return g
//}
//hexBytes, err := hexStringToBytes(candidate)
//if err == nil {
//g, err = geom.UnmarshalWKB(bytes.NewReader(hexBytes))
//if err == nil {
//return g
//}
//}
//g, err = geom.UnmarshalGeoJSON([]byte(candidate))
//if err == nil {
//return g
//}
//// We've already checked that we can parse the geometry by this point. So
//// it's a programming mistake if this panic occurs.
//panic("could not parse geometry")
//}

//func (c *corpus) loadGeometries(t *testing.T) {
//type mismatch struct {
//candidate string
//checkName string
//errs      [2]error
//}
//var mismatches []mismatch
//for _, check := range []struct {
//method     func(GeometryEngine, string) error
//name       string
//exceptions map[string]struct{}
//}{
//{
//method: func(e GeometryEngine, s string) error { return e.ValidateWKT(s) },
//name:   "validate wkt",
//},
//{
//method: func(e GeometryEngine, s string) error { return e.ValidateWKB(s) },
//name:   "validate wkb",
//},
//{
//method: func(e GeometryEngine, s string) error { return e.ValidateGeoJSON(s) },
//name:   "validate geojson",
//exceptions: map[string]struct{}{
//// From https://tools.ietf.org/html/rfc7946#section-3.1:
////
//// > GeoJSON processors MAY interpret Geometry objects with
//// > empty "coordinates" arrays as null objects.
////
//// Simplefeatures chooses to accept this as an empty point, but
//// Postgres rejects it.
//`{"type":"Point","coordinates":[]}`: struct{}{},
//},
//},
//} {
//for _, candidate := range c.candidates {
//if _, isException := check.exceptions[candidate]; isException {
//continue
//}
//err0 := check.method(c.engines[0], candidate)
//err1 := check.method(c.engines[1], candidate)
//if (err0 == nil) == (err1 == nil) {
//if err0 == nil {
//c.geometries = append(c.geometries, parseGeometry(candidate))
//}
//continue // error state matches
//}
//mismatches = append(mismatches, mismatch{candidate, check.name, [2]error{err0, err1}})
//}
//}

//time.Sleep(time.Second) // Wait for Postgis output to stop
//for _, m := range mismatches {
//log.Println("=== MISMATCH ===")
//log.Println("candidate: ", m.candidate)
//log.Println("check name: ", m.checkName)
//log.Println("err0: ", m.errs[0])
//log.Println("err1: ", m.errs[1])
//}
//log.Println("===")
//log.Println("mismatches:", len(mismatches))
//log.Println("valid geometries:", len(c.geometries))
//}

//func (c *corpus) checkProperties() {
//for _, check := range []struct {
//name   string
//method func(geom.Geometry) error
//}{
//{
//"WKT", func(g geom.Geometry) error {
//if _, ok := g.(geom.MultiPoint); ok {
//// Skip Multipoints. This is because Postgis doesn't follow
//// the SFA spec by not including parenthesis around each
//// individual point. The simplefeatures library follows the
//// spec correctly.
//return nil
//}
//wkt0, err0 := c.engines[0].AsText(g)
//wkt1, err1 := c.engines[1].AsText(g)
//if err0 != nil {
//return err0
//}
//if err1 != nil {
//return err1
//}
//if wkt0 != wkt1 {
//return fmt.Errorf("mismatch: %v vs %v", wkt0, wkt1)
//}
//return nil
//},
//},
//} {
//for _, g := range c.geometries {
//if err := check.method(g); err != nil {
//log.Printf("%s check failed: %v", check.name, err)
//}
//}
//}
//}
