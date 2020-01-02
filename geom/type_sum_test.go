package geom_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestZeroGeometry(t *testing.T) {
	var z Geometry
	expectDeepEqual(t, z.IsGeometryCollection(), true)
	z.AsGeometryCollection() // Doesn't crash.
	expectDeepEqual(t, z.AsText(), "GEOMETRYCOLLECTION EMPTY")

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(z)
	expectNoErr(t, err)
	expectDeepEqual(t, strings.TrimSpace(buf.String()), `{"type":"GeometryCollection","geometries":[]}`)

	z = NewPointF(1, 2).AsGeometry() // Set away from zero value
	expectDeepEqual(t, z.IsPoint(), true)
	err = json.NewDecoder(&buf).Decode(&z)
	expectNoErr(t, err)
	expectDeepEqual(t, z.IsPoint(), false)
	expectDeepEqual(t, z.IsGeometryCollection(), true)
	expectDeepEqual(t, z.IsEmpty(), true)
	z = Geometry{}

	z.AsBinary(ioutil.Discard) // Doesn't crash
	expectDeepEqual(t, z.Dimension(), 0)

	// TODO: continue further tests
}
