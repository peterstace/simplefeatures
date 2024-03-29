// Package geos provides a cgo wrapper around the GEOS (Geometry Engine, Open
// Source)  library. See https://www.osgeo.org/projects/geos/ for more details.
//
// Its purpose is to provide functionality that has been implemented in GEOS,
// but is not yet available natively in the simplefeatures library.
//
// Results from functions in this package are returned from GEOS unvalidated
// and as-is. Users may call the Validate method on results if they wish to
// check result validity.
//
// The operations in this package ignore Z and M values if they are present.
//
// To use this package, you will need to install the GEOS library.
package geos
