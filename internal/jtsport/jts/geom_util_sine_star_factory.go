package jts

import "math"

func GeomUtil_SineStarFactory_Create(origin *Geom_Coordinate, size float64, nPts int, nArms int, armLengthRatio float64) *Geom_Geometry {
	gsf := GeomUtil_NewSineStarFactory()
	gsf.SetCentre(origin)
	gsf.SetSize(size)
	gsf.SetNumPoints(nPts)
	gsf.SetArmLengthRatio(armLengthRatio)
	gsf.SetNumArms(nArms)
	poly := gsf.CreateSineStar()
	return poly
}

type GeomUtil_SineStarFactory struct {
	*Util_GeometricShapeFactory
	numArms        int
	armLengthRatio float64
}

func GeomUtil_NewSineStarFactory() *GeomUtil_SineStarFactory {
	return GeomUtil_NewSineStarFactoryWithFactory(Geom_NewGeometryFactoryDefault())
}

func GeomUtil_NewSineStarFactoryWithFactory(geomFact *Geom_GeometryFactory) *GeomUtil_SineStarFactory {
	return &GeomUtil_SineStarFactory{
		Util_GeometricShapeFactory: Util_NewGeometricShapeFactoryWithFactory(geomFact),
		numArms:                    8,
		armLengthRatio:             0.5,
	}
}

func (gsf *GeomUtil_SineStarFactory) SetNumArms(numArms int) {
	gsf.numArms = numArms
}

func (gsf *GeomUtil_SineStarFactory) SetArmLengthRatio(armLengthRatio float64) {
	gsf.armLengthRatio = armLengthRatio
}

func (gsf *GeomUtil_SineStarFactory) CreateSineStar() *Geom_Geometry {
	env := gsf.dim.getEnvelope()
	radius := env.GetWidth() / 2.0

	armRatio := gsf.armLengthRatio
	if armRatio < 0.0 {
		armRatio = 0.0
	}
	if armRatio > 1.0 {
		armRatio = 1.0
	}

	armMaxLen := armRatio * radius
	insideRadius := (1 - armRatio) * radius

	centreX := env.GetMinX() + radius
	centreY := env.GetMinY() + radius

	pts := make([]*Geom_Coordinate, gsf.nPts+1)
	iPt := 0
	for i := 0; i < gsf.nPts; i++ {
		// the fraction of the way through the current arm - in [0,1]
		ptArcFrac := (float64(i) / float64(gsf.nPts)) * float64(gsf.numArms)
		armAngFrac := ptArcFrac - math.Floor(ptArcFrac)

		// the angle for the current arm - in [0,2Pi]
		// (each arm is a complete sine wave cycle)
		armAng := 2 * math.Pi * armAngFrac
		// the current length of the arm
		armLenFrac := (math.Cos(armAng) + 1.0) / 2.0

		// the current radius of the curve (core + arm)
		curveRadius := insideRadius + armMaxLen*armLenFrac

		// the current angle of the curve
		ang := float64(i) * (2 * math.Pi / float64(gsf.nPts))
		x := curveRadius*math.Cos(ang) + centreX
		y := curveRadius*math.Sin(ang) + centreY
		pts[iPt] = gsf.coord(x, y)
		iPt++
	}
	pts[iPt] = Geom_NewCoordinateFromCoordinate(pts[0])

	ring := gsf.geomFact.CreateLinearRingFromCoordinates(pts)
	poly := gsf.geomFact.CreatePolygonFromLinearRing(ring)
	return poly.Geom_Geometry
}
