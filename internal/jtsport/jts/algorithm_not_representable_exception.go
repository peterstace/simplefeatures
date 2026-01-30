package jts

// Algorithm_NotRepresentableException indicates that a HCoordinate has been
// computed which is not representable on the Cartesian plane.
type Algorithm_NotRepresentableException struct {
	message string
}

// Algorithm_NewNotRepresentableException creates a new
// NotRepresentableException.
func Algorithm_NewNotRepresentableException() *Algorithm_NotRepresentableException {
	return &Algorithm_NotRepresentableException{
		message: "Projective point not representable on the Cartesian plane.",
	}
}

func (e *Algorithm_NotRepresentableException) Error() string {
	return e.message
}
