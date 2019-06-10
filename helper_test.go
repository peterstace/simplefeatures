package simplefeatures_test

import . "github.com/peterstace/simplefeatures"

func Must(g Geometry, err error) Geometry {
	if err != nil {
		panic(err)
	}
	return g
}

func MustNewPoint(x, y float64) Point {
	return Must(NewPoint(x, y)).(Point)
}

func MustNewLineString(pts []Point) LineString {
	return Must(NewLineString(pts)).(LineString)
}
