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
	expectBoolEq(t, z.IsGeometryCollection(), true)
	z.AsGeometryCollection() // Doesn't crash.
	expectStringEq(t, z.AsText(), "GEOMETRYCOLLECTION EMPTY")

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(z)
	expectNoErr(t, err)
	expectStringEq(t, strings.TrimSpace(buf.String()), `{"type":"GeometryCollection","geometries":[]}`)

	z = NewPointF(1, 2).AsGeometry() // Set away from zero value
	expectBoolEq(t, z.IsPoint(), true)
	err = json.NewDecoder(&buf).Decode(&z)
	expectNoErr(t, err)
	expectBoolEq(t, z.IsPoint(), false)
	expectBoolEq(t, z.IsGeometryCollection(), true)
	expectBoolEq(t, z.IsEmpty(), true)
	z = Geometry{}

	z.AsBinary(ioutil.Discard) // Doesn't crash

	_, err = z.Value()
	expectNoErr(t, err)

	expectIntEq(t, z.Dimension(), 0)
}
