# Simple Features

[![Documentation](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/peterstace/simplefeatures/geom?tab=doc)
[![Build Status](https://github.com/peterstace/simplefeatures/workflows/build/badge.svg)](https://github.com/peterstace/simplefeatures/actions)
[![Go Report
Card](https://goreportcard.com/badge/github.com/peterstace/simplefeatures)](https://goreportcard.com/report/github.com/peterstace/simplefeatures)

Simple Features is a 2D geometry library. It provides Go types that model
geometries, as well as algorithms that operate on them.

Simple Features is a pure Go Implementation of the OpenGIS Consortium's Simple
Feature Access Specification (which can be found
[here](http://www.opengeospatial.org/standards/sfa)). This is the same
specification that [GEOS](https://trac.osgeo.org/geos),
[JTS](https://locationtech.github.io/jts/), and [PostGIS](https://postgis.net/)
implement, so the Simple Features API will be familiar to developers who have
used those libraries before.

### Geometry Types

<table>

<tr>
<td>Point</td>
<td><img src="./.hidden/assets/point.svg"></td>
<td>TODO</td>
</tr>

<tr>
<td>MultiPoint</td>
<td><img src="./.hidden/assets/multipoint.svg"></td>
<td>TODO</td>
</tr>

<tr>
<td>LineString</td>
<td><img src="./.hidden/assets/linestring.svg"></td>
<td>TODO</td>
</tr>

<tr>
<td>MultiLineString</td>
<td><img src="./.hidden/assets/multilinestring.svg"></td>
<td>TODO</td>
</tr>

<tr>
<td>Polygon</td>
<td><a href="https://upload.wikimedia.org/wikipedia/commons/5/55/SFA_Polygon_with_hole.svg"><img src="https://upload.wikimedia.org/wikipedia/commons/5/55/SFA_Polygon_with_hole.svg"></a></td>
<td>TODO</td>
</tr>

<tr>
<td>MultiPolygon</td>
<td><img src="./.hidden/assets/multipolygon.svg"></td>
<td>TODO</td>
</tr>

<tr>
<td>GeometryCollection</td>
<td><img src="./.hidden/assets/geometrycollection.svg"></td>
<td>TODO</td>
</tr>

<tr>
<td>Geometry</td>
<td><img src="./.hidden/assets/geometry.svg"></td>
<td>TODO</td>
</tr>

<tr>
<td>Envelope</td>
<td><img src="./.hidden/assets/envelope.svg"></td>
<td>TODO</td>
</tr>

</table>

### Marshalling and Unmarshalling

TODO

### Geometry Algorithms

TODO

### GEOS Wrapper

A [GEOS](https://www.osgeo.org/projects/geos/) CGO wrapper is also provided,
giving access to functionality not yet implemented natively in Go. The [wrapper
is implemented in a separate
package](https://pkg.go.dev/github.com/peterstace/simplefeatures/geos?tab=doc),
meaning that library users who don't need this additional functionality don't
need to expose themselves to CGO.

### FAQs

**Q:** Why create Simple Features when the GEOS library already exists?

**A:** TODO
