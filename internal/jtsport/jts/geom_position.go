package jts

// Geom_Position_On specifies that a location is on a component.
const Geom_Position_On = 0

// Geom_Position_Left specifies that a location is to the left of a component.
const Geom_Position_Left = 1

// Geom_Position_Right specifies that a location is to the right of a component.
const Geom_Position_Right = 2

// Geom_Position_Opposite returns Left if the position is Right, Right if the position is
// Left, or the position otherwise.
func Geom_Position_Opposite(position int) int {
	if position == Geom_Position_Left {
		return Geom_Position_Right
	}
	if position == Geom_Position_Right {
		return Geom_Position_Left
	}
	return position
}
