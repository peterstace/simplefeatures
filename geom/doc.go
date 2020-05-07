// Package geom implements the OpenGIS Simple Feature Access specification. The
// specification describes an access and storage model for 2-dimensional
// geometries.
//
// The package serves three primary purposes:
//
// 1. Access: It provides a type for each of the different geometry types
// described by the standard: Point, MultiPoint, LineString, MultiLineString,
// Polygon, MultiPolygon, and GeometryCollection. It also contains supporting
// types such as Geometry, Envelope, Sequence, and XY.  Methods on these types
// allow access the internal parts of each geometry. For example, there is a
// method that will obtain the first Point in a LineString.
//
// 2. Analysis: There are methods on the types that perform spatial
// analysis on the geometries. For example, to check if a geometry is simple or
// to calculate its smallest bounding box.
//
// 3. Storage: The types implement various methods that allow conversion to
// and from various storage and encoding formats. WKT (Well Known Text), WKB
// (Well Known Binary), and GeoJSON are supported.
package geom
