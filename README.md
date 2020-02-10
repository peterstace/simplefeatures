# Simple Features

[![Build Status](https://travis-ci.com/peterstace/simplefeatures.svg?token=ueRpGt4cSSnk321nW8xG&branch=master)](https://travis-ci.com/peterstace/simplefeatures)
[![Documentation](https://godoc.org/github.com/peterstace/simplefeatures?status.svg)](http://godoc.org/github.com/peterstace/simplefeatures/geom)

Simple Features is a pure Go Implementation of the OpenGIS Simple Feature Access
Specification (which can be found
[here](http://www.opengeospatial.org/standards/sfa)).

The specification describes a common access and storage model for 2-dimensional
geometries. This is the same access and storage model used by libraries such as
[GEOS](https://trac.osgeo.org/geos),
[JTS](https://locationtech.github.io/jts/), and
[PostGIS](https://postgis.net/).

#### Changelog

The changelog can be found [here](CHANGELOG.md).

#### Supported Features

- Marshalling/unmarshalling:
	- WKT (well known text)
	- WKB (well known binary)
	- GeoJSON

- Geometry attribute calculations:
	- Geometry validity checks
	- Dimensionality check
	- Bounding box calculation
	- Emptiness check
	- Boundary calculation

- Spatial analysis:
	- Convex Hull calculation
	- Intersects check
	- Length calculation
	- Closed geometry calculation
	- Ring property calculation
	- Area calculation
	- Centroid calculation

#### In the works

- Spatial analysis:
	- Intersection calculation
	- Spatially equality calculation
	- Point on surface calculation

#### Features Not Planned Yet

- SRIDs
- 3D/Measure coordinates.

- Spatial analysis:
	- Geometry buffering
	- Disjoint check
	- Touches check
	- Crosses check
	- Within check
	- Contains check
	- Overlaps check
	- Relates check

### Tests

Some of the tests have a dependency on a [Postgis](https://postgis.net/)
database being available.

While the tests can be run in the usual Go way if you have Postgis set up
locally, it's easier to run the tests using docker-compose:

```
docker-compose up --abort-on-container-exit
```

There are two additional suite of test suits utilising an automatically
generated test corpus. The test suite test every function against every input
combination exhaustively, and compare the result against a reference
implementation. They take much longer to run, and are designed to be used as a
final double check for correctness. They can be run using the following
commands:

```
docker-compose -f docker-compose-full.yml up --abort-on-container-exit
```

```
docker-compose -f docker-compose-cmprefimpl.yml up
```
