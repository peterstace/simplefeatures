package geom

func transformOptional1dCoords(coords []OptionalCoordinates, fn func(XY) XY) {
	for i := range coords {
		if coords[i].Present {
			coords[i].Value.XY = fn(coords[i].Value.XY)
		}
	}
}

func transform1dCoords(coords []Coordinates, fn func(XY) XY) {
	for i := range coords {
		coords[i].XY = fn(coords[i].XY)
	}
}

func transform2dCoords(coords [][]Coordinates, fn func(XY) XY) {
	for i := range coords {
		transform1dCoords(coords[i], fn)
	}
}

func transform3dCoords(coords [][][]Coordinates, fn func(XY) XY) {
	for i := range coords {
		transform2dCoords(coords[i], fn)
	}
}
