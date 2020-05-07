package rtree

// Package rtree implements an in-memory r-tree data structure. This data
// structure can be used as a spatial index, allowing fast spatial searches
// based on a bounding box.
//
// The implementation is heavily based on ["R-Trees. A Dynamic Index Structure
// For Spatial
// Searching"](http://www-db.deis.unibo.it/courses/SI-LS/papers/Gut84.pdf) by
// Antonin Guttman.
