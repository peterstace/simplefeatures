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

// TODO: Remove this type.

// AnyGeometry is a concrete type that holds any GeometryX value. It exists as a
// helper to make SQL and JSON interactions easier.
type AnyGeometry struct {
	// Geom is the destination into which scanned geometries are stored.
	Geom GeometryX

	// Options control the way that geometries are constructed.
	Options []ConstructorOption
}

// SetOptions is a helper to set provided variadic list of options into
// the Options field.
func (a *AnyGeometry) SetOptions(opts ...ConstructorOption) {
	a.Options = opts
}

// Value implements the "sql/driver".Valuer interface by emitting WKB. Gives
// an error if the Geom element is nil.
func (a AnyGeometry) Value() (driver.Value, error) {
	if a.Geom == nil {
		return nil, errors.New("no geometry set")
	}
	var buf bytes.Buffer
	err := ToGeometry(a.Geom).AsBinary(&buf)
	return buf.Bytes(), err
}

// Scan implements the "sql".Scanner interface by parsing the src value as WKB.
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
	a.Geom, err = UnmarshalWKB(r, a.Options...)
	return err
}

// UnmarshalJSON implements the "encoding/json".Unmarshaller interface by
// parsing the JSON stream as GeoJSON.
func (a *AnyGeometry) UnmarshalJSON(p []byte) error {
	geom, err := UnmarshalGeoJSON(p, a.Options...)
	if err != nil {
		return err
	}
	a.Geom = geom
	return nil
}

func (a AnyGeometry) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Geom)
}
