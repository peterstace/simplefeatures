package geom

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// AnyGeometry is a concrete type that holds any Geometry value. It exists to
// make SQL interactions easier (specifically to allow geometries to be scanned
// in from an SQL DB).
type AnyGeometry struct {
	Geom Geometry
}

// Value implements the "sql/driver".Valuer interface by emitting WKT. Gives
// an error if the Geom element is nil.
func (a *AnyGeometry) Value() (driver.Value, error) {
	if a.Geom == nil {
		return nil, errors.New("no geometry set")
	}
	return a.Geom.AsText(), nil
}

// Scan implements the "sql".Scanner interface by parsing the src value as WKT.
func (a *AnyGeometry) Scan(src interface{}) error {
	var r io.Reader
	switch src := src.(type) {
	case []byte:
		r = bytes.NewReader(src)
	case string:
		r = strings.NewReader(src)
	default:
		// nil is specifically not supported. It _could_ map to an empty
		// geometry, however then the caller wouldn't be able to differentiate
		// between a real empty geometry and a NULL. Instead, we should
		// additionally provide a NullableAnyGeometry type with an IsValid flag.
		return fmt.Errorf("unsupported src type in Scan: %T", src)
	}

	var err error
	a.Geom, err = UnmarshalWKT(r)
	return err
}

// UnmarshalJSON implements the "encoding/json".Unmarshaller interface by
// parsing the JSON stream as GeoJSON.
func (a *AnyGeometry) UnmarshalJSON(p []byte) error {
	geom, err := UnmarshalGeoJSON(p)
	if err != nil {
		return err
	}
	a.Geom = geom
	return nil
}

func (a AnyGeometry) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Geom)
}
