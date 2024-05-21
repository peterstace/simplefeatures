package geom

import "database/sql/driver"

// NullGeometry represents a Geometry that may be NULL. It implements the
// database/sql.Scanner and database/sql/driver.Valuer interfaces, so may be
// used as a scan destination or query argument in SQL queries.
type NullGeometry struct {
	Geometry Geometry
	Valid    bool // Valid is true iff Geometry is not NULL
}

// Scan implements the database/sql.Scanner interface.
func (ng *NullGeometry) Scan(value interface{}) error {
	if value == nil {
		ng.Geometry = Geometry{}
		ng.Valid = false
		return nil
	}
	ng.Valid = true
	return ng.Geometry.Scan(value)
}

// Value implements the database/sql/driver.Valuer interface.
func (ng NullGeometry) Value() (driver.Value, error) {
	if !ng.Valid {
		return nil, nil //nolint:nilnil
	}
	return ng.Geometry.Value()
}
