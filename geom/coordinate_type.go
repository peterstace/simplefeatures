package geom

type CoordinatesType byte

const (
	DimXY   CoordinatesType = 0b00
	DimXYZ  CoordinatesType = 0b01
	DimXYM  CoordinatesType = 0b10
	DimXYZM CoordinatesType = 0b11
)

func (t CoordinatesType) String() string {
	return [4]string{"XY", "XYZ", "XYM", "XYZM"}[t]
}

func (t CoordinatesType) Dimension() int {
	return [4]int{2, 3, 3, 4}[t]
}

func (t CoordinatesType) Is3D() bool {
	return (t & DimXYZ) != 0
}

func (t CoordinatesType) IsMeasured() bool {
	return (t & DimXYM) != 0
}

func (t CoordinatesType) wktModifier() string {
	return [4]string{"", " Z ", " M ", " ZM "}[t]
}
