package simplefeatures

// MultiLineString is a multicurve whose elements are LineStrings.
type MultiLineString struct {
	lines []LineString
}

func NewMultiLineString(lines []LineString) (MultiLineString, error) {
	// TODO: validation
	return MultiLineString{lines}, nil
}

func NewMultiLineStringFromCoordinates(coords [][]Coordinates) (MultiLineString, error) {
	lines := make([]LineString, len(coords))
	for i, c := range coords {
		line, err := NewLineStringFromCoords(c)
		if err != nil {
			return MultiLineString{}, err
		}
		lines[i] = line
	}
	return MultiLineString{lines}, nil
}
