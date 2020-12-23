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
<td>TODO</td>
<td>TODO</td>
</tr>

<tr>
<td>MultiPoint</td>
<td>
<svg xmlns="http://www.w3.org/2000/svg"
 width="467" height="462">
  <rect x="80" y="60" width="250" height="250" rx="20"
      style="fill:#00ff00; stroke:#000000;stroke-width:2px;" />
  
  <rect x="140" y="120" width="250" height="250" rx="40"
      style="fill:#0000ff; stroke:#000000; stroke-width:2px;
      fill-opacity:0.7;" />
</svg>
</td>
<td>TODO</td>
</tr>

<tr>
<td>LineString</td>
<td>TODO</td>
<td>TODO</td>
</tr>

<tr>
<td>MultiLineString</td>
<td>TODO</td>
<td>TODO</td>
</tr>

<tr>
<td>Polygon</td>
<td>TODO</td>
<td>TODO</td>
</tr>

<tr>
<td>MultiPolygon</td>
<td>TODO</td>
<td>TODO</td>
</tr>

<tr>
<td>GeometryCollection</td>
<td>TODO</td>
<td>TODO</td>
</tr>

<tr>
<td>Geometry</td>
<td>TODO</td>
<td>TODO</td>
</tr>

<tr>
<td>Envelope</td>
<td>TODO</td>
<td>TODO</td>
</tr>

</table>

![TEST](./.hidden/assets/test.svg)
<img src="./.hidden/assets/test.svg">
<img src="./.hidden/assets/test.svg?sanitize=true">
<img src="https://raw.githubusercontent.com/peterstace/simplefeatures/revamp_readme/.hidden/assets/test.svg"/>

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
