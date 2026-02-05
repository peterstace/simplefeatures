package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

type Util_GeometricShapeFactory struct {
	geomFact      *Geom_GeometryFactory
	precModel     *Geom_PrecisionModel
	dim           *util_GeometricShapeFactory_Dimensions
	nPts          int
	rotationAngle float64
}

func Util_NewGeometricShapeFactory() *Util_GeometricShapeFactory {
	return Util_NewGeometricShapeFactoryWithFactory(Geom_NewGeometryFactoryDefault())
}

func Util_NewGeometricShapeFactoryWithFactory(geomFact *Geom_GeometryFactory) *Util_GeometricShapeFactory {
	return &Util_GeometricShapeFactory{
		geomFact:      geomFact,
		precModel:     geomFact.GetPrecisionModel(),
		dim:           util_newGeometricShapeFactory_Dimensions(),
		nPts:          100,
		rotationAngle: 0.0,
	}
}

func (gsf *Util_GeometricShapeFactory) SetEnvelope(env *Geom_Envelope) {
	gsf.dim.setEnvelope(env)
}

func (gsf *Util_GeometricShapeFactory) SetBase(base *Geom_Coordinate) {
	gsf.dim.setBase(base)
}

func (gsf *Util_GeometricShapeFactory) SetCentre(centre *Geom_Coordinate) {
	gsf.dim.setCentre(centre)
}

func (gsf *Util_GeometricShapeFactory) SetNumPoints(nPts int) { gsf.nPts = nPts }

func (gsf *Util_GeometricShapeFactory) SetSize(size float64) { gsf.dim.setSize(size) }

func (gsf *Util_GeometricShapeFactory) SetWidth(width float64) { gsf.dim.setWidth(width) }

func (gsf *Util_GeometricShapeFactory) SetHeight(height float64) { gsf.dim.setHeight(height) }

func (gsf *Util_GeometricShapeFactory) SetRotation(radians float64) {
	gsf.rotationAngle = radians
}

func (gsf *Util_GeometricShapeFactory) rotate(geom *Geom_Geometry) *Geom_Geometry {
	if gsf.rotationAngle != 0.0 {
		trans := GeomUtil_AffineTransformation_RotationInstance(gsf.rotationAngle,
			gsf.dim.getCentre().X, gsf.dim.getCentre().Y)
		geom.ApplyCoordinateSequenceFilter(trans)
	}
	return geom
}

func (gsf *Util_GeometricShapeFactory) CreateRectangle() *Geom_Polygon {
	var i int
	ipt := 0
	nSide := gsf.nPts / 4
	if nSide < 1 {
		nSide = 1
	}
	xSegLen := gsf.dim.getEnvelope().GetWidth() / float64(nSide)
	ySegLen := gsf.dim.getEnvelope().GetHeight() / float64(nSide)

	pts := make([]*Geom_Coordinate, 4*nSide+1)
	env := gsf.dim.getEnvelope()

	//double maxx = env.getMinX() + nSide * XsegLen;
	//double maxy = env.getMinY() + nSide * XsegLen;

	for i = 0; i < nSide; i++ {
		x := env.GetMinX() + float64(i)*xSegLen
		y := env.GetMinY()
		pts[ipt] = gsf.coord(x, y)
		ipt++
	}
	for i = 0; i < nSide; i++ {
		x := env.GetMaxX()
		y := env.GetMinY() + float64(i)*ySegLen
		pts[ipt] = gsf.coord(x, y)
		ipt++
	}
	for i = 0; i < nSide; i++ {
		x := env.GetMaxX() - float64(i)*xSegLen
		y := env.GetMaxY()
		pts[ipt] = gsf.coord(x, y)
		ipt++
	}
	for i = 0; i < nSide; i++ {
		x := env.GetMinX()
		y := env.GetMaxY() - float64(i)*ySegLen
		pts[ipt] = gsf.coord(x, y)
		ipt++
	}
	pts[ipt] = Geom_NewCoordinateFromCoordinate(pts[0])
	ipt++

	ring := gsf.geomFact.CreateLinearRingFromCoordinates(pts)
	poly := gsf.geomFact.CreatePolygonFromLinearRing(ring)
	return java.Cast[*Geom_Polygon](gsf.rotate(poly.Geom_Geometry))
}

//* @deprecated use createEllipse instead

func (gsf *Util_GeometricShapeFactory) CreateCircle() *Geom_Polygon {
	return gsf.CreateEllipse()
}

func (gsf *Util_GeometricShapeFactory) CreateEllipse() *Geom_Polygon {
	env := gsf.dim.getEnvelope()
	xRadius := env.GetWidth() / 2.0
	yRadius := env.GetHeight() / 2.0

	centreX := env.GetMinX() + xRadius
	centreY := env.GetMinY() + yRadius

	pts := make([]*Geom_Coordinate, gsf.nPts+1)
	iPt := 0
	for i := 0; i < gsf.nPts; i++ {
		ang := float64(i) * (2 * math.Pi / float64(gsf.nPts))
		x := xRadius*Algorithm_Angle_CosSnap(ang) + centreX
		y := yRadius*Algorithm_Angle_SinSnap(ang) + centreY
		pts[iPt] = gsf.coord(x, y)
		iPt++
	}
	pts[iPt] = Geom_NewCoordinateFromCoordinate(pts[0])

	ring := gsf.geomFact.CreateLinearRingFromCoordinates(pts)
	poly := gsf.geomFact.CreatePolygonFromLinearRing(ring)
	return java.Cast[*Geom_Polygon](gsf.rotate(poly.Geom_Geometry))
}

func (gsf *Util_GeometricShapeFactory) CreateSquircle() *Geom_Polygon {
	return gsf.CreateSupercircle(4)
}

func (gsf *Util_GeometricShapeFactory) CreateSupercircle(power float64) *Geom_Polygon {
	recipPow := 1.0 / power

	radius := gsf.dim.getMinSize() / 2
	centre := gsf.dim.getCentre()

	r4 := math.Pow(radius, power)
	y0 := radius

	xyInt := math.Pow(r4/2, recipPow)

	nSegsInOct := gsf.nPts / 8
	totPts := nSegsInOct*8 + 1
	pts := make([]*Geom_Coordinate, totPts)
	xInc := xyInt / float64(nSegsInOct)

	for i := 0; i <= nSegsInOct; i++ {
		x := 0.0
		y := y0
		if i != 0 {
			x = xInc * float64(i)
			x4 := math.Pow(x, power)
			y = math.Pow(r4-x4, recipPow)
		}
		pts[i] = gsf.coordTrans(x, y, centre)
		pts[2*nSegsInOct-i] = gsf.coordTrans(y, x, centre)

		pts[2*nSegsInOct+i] = gsf.coordTrans(y, -x, centre)
		pts[4*nSegsInOct-i] = gsf.coordTrans(x, -y, centre)

		pts[4*nSegsInOct+i] = gsf.coordTrans(-x, -y, centre)
		pts[6*nSegsInOct-i] = gsf.coordTrans(-y, -x, centre)

		pts[6*nSegsInOct+i] = gsf.coordTrans(-y, x, centre)
		pts[8*nSegsInOct-i] = gsf.coordTrans(-x, y, centre)
	}
	pts[len(pts)-1] = Geom_NewCoordinateFromCoordinate(pts[0])

	ring := gsf.geomFact.CreateLinearRingFromCoordinates(pts)
	poly := gsf.geomFact.CreatePolygonFromLinearRing(ring)
	return java.Cast[*Geom_Polygon](gsf.rotate(poly.Geom_Geometry))
}

func (gsf *Util_GeometricShapeFactory) CreateArc(startAng, angExtent float64) *Geom_LineString {
	env := gsf.dim.getEnvelope()
	xRadius := env.GetWidth() / 2.0
	yRadius := env.GetHeight() / 2.0

	centreX := env.GetMinX() + xRadius
	centreY := env.GetMinY() + yRadius

	angSize := angExtent
	if angSize <= 0.0 || angSize > Algorithm_Angle_PiTimes2 {
		angSize = Algorithm_Angle_PiTimes2
	}
	angInc := angSize / float64(gsf.nPts-1)

	pts := make([]*Geom_Coordinate, gsf.nPts)
	iPt := 0
	for i := 0; i < gsf.nPts; i++ {
		ang := startAng + float64(i)*angInc
		x := xRadius*Algorithm_Angle_CosSnap(ang) + centreX
		y := yRadius*Algorithm_Angle_SinSnap(ang) + centreY
		pts[iPt] = gsf.coord(x, y)
		iPt++
	}
	line := gsf.geomFact.CreateLineStringFromCoordinates(pts)
	return java.Cast[*Geom_LineString](gsf.rotate(line.Geom_Geometry))
}

func (gsf *Util_GeometricShapeFactory) CreateArcPolygon(startAng, angExtent float64) *Geom_Polygon {
	env := gsf.dim.getEnvelope()
	xRadius := env.GetWidth() / 2.0
	yRadius := env.GetHeight() / 2.0

	centreX := env.GetMinX() + xRadius
	centreY := env.GetMinY() + yRadius

	angSize := angExtent
	if angSize <= 0.0 || angSize > Algorithm_Angle_PiTimes2 {
		angSize = Algorithm_Angle_PiTimes2
	}
	angInc := angSize / float64(gsf.nPts-1)
	// double check = angInc * nPts;
	// double checkEndAng = startAng + check;

	pts := make([]*Geom_Coordinate, gsf.nPts+2)

	iPt := 0
	pts[iPt] = gsf.coord(centreX, centreY)
	iPt++
	for i := 0; i < gsf.nPts; i++ {
		ang := startAng + angInc*float64(i)

		x := xRadius*Algorithm_Angle_CosSnap(ang) + centreX
		y := yRadius*Algorithm_Angle_SinSnap(ang) + centreY
		pts[iPt] = gsf.coord(x, y)
		iPt++
	}
	pts[iPt] = gsf.coord(centreX, centreY)
	iPt++
	ring := gsf.geomFact.CreateLinearRingFromCoordinates(pts)
	poly := gsf.geomFact.CreatePolygonFromLinearRing(ring)
	return java.Cast[*Geom_Polygon](gsf.rotate(poly.Geom_Geometry))
}

func (gsf *Util_GeometricShapeFactory) coord(x, y float64) *Geom_Coordinate {
	pt := Geom_NewCoordinateWithXY(x, y)
	gsf.precModel.MakePreciseCoordinate(pt)
	return pt
}

func (gsf *Util_GeometricShapeFactory) coordTrans(x, y float64, trans *Geom_Coordinate) *Geom_Coordinate {
	return gsf.coord(x+trans.X, y+trans.Y)
}

// util_GeometricShapeFactory_Dimensions is the inner Dimensions class.
type util_GeometricShapeFactory_Dimensions struct {
	base   *Geom_Coordinate
	centre *Geom_Coordinate
	width  float64
	height float64
}

func util_newGeometricShapeFactory_Dimensions() *util_GeometricShapeFactory_Dimensions {
	return &util_GeometricShapeFactory_Dimensions{}
}

func (d *util_GeometricShapeFactory_Dimensions) setBase(base *Geom_Coordinate) { d.base = base }
func (d *util_GeometricShapeFactory_Dimensions) getBase() *Geom_Coordinate     { return d.base }

func (d *util_GeometricShapeFactory_Dimensions) setCentre(centre *Geom_Coordinate) {
	d.centre = centre
}

func (d *util_GeometricShapeFactory_Dimensions) getCentre() *Geom_Coordinate {
	if d.centre == nil {
		d.centre = Geom_NewCoordinateWithXY(d.base.X+d.width/2, d.base.Y+d.height/2)
	}
	return d.centre
}

func (d *util_GeometricShapeFactory_Dimensions) setSize(size float64) {
	d.height = size
	d.width = size
}

func (d *util_GeometricShapeFactory_Dimensions) getMinSize() float64 {
	return math.Min(d.width, d.height)
}

func (d *util_GeometricShapeFactory_Dimensions) setWidth(width float64)   { d.width = width }
func (d *util_GeometricShapeFactory_Dimensions) getWidth() float64        { return d.width }
func (d *util_GeometricShapeFactory_Dimensions) getHeight() float64       { return d.height }
func (d *util_GeometricShapeFactory_Dimensions) setHeight(height float64) { d.height = height }

func (d *util_GeometricShapeFactory_Dimensions) setEnvelope(env *Geom_Envelope) {
	d.width = env.GetWidth()
	d.height = env.GetHeight()
	d.base = Geom_NewCoordinateWithXY(env.GetMinX(), env.GetMinY())
	d.centre = Geom_NewCoordinateFromCoordinate(env.Centre())
}

func (d *util_GeometricShapeFactory_Dimensions) getEnvelope() *Geom_Envelope {
	if d.base != nil {
		return Geom_NewEnvelopeFromXY(d.base.X, d.base.X+d.width, d.base.Y, d.base.Y+d.height)
	}
	if d.centre != nil {
		return Geom_NewEnvelopeFromXY(d.centre.X-d.width/2, d.centre.X+d.width/2,
			d.centre.Y-d.height/2, d.centre.Y+d.height/2)
	}
	return Geom_NewEnvelopeFromXY(0, d.width, 0, d.height)
}
