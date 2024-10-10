package carto

import "github.com/peterstace/simplefeatures/geom"

type Equirectangular struct {
	CentralMeridian   float64
	StandardParallels float64
	CentralParallel   float64
	Radius            float64
}

func (e *Equirectangular) To(lonLat geom.XY) geom.XY {
	var (
		λd  = lonLat.X
		φd  = lonLat.Y
		λ0d = e.CentralMeridian
		φ1d = e.StandardParallels
		φ0d = e.CentralParallel
		R   = e.Radius
		λr  = dtor(λd)
		φr  = dtor(φd)
		λ0r = dtor(λ0d)
		φ1r = dtor(φ1d)
		φ0r = dtor(φ0d)
	)
	var (
		x = R * (λr - λ0r) * cos(φ1r)
		y = R * (φr - φ0r)
	)
	return geom.XY{X: x, Y: y}
}

func (e *Equirectangular) From(xy geom.XY) geom.XY {
	var (
		x   = xy.X
		y   = xy.Y
		λ0d = e.CentralMeridian
		φ1d = e.StandardParallels
		φ0d = e.CentralParallel
		R   = e.Radius
		λ0r = dtor(λ0d)
		φ1r = dtor(φ1d)
		φ0r = dtor(φ0d)
	)
	var (
		λr = x/(R*cos(φ1r)) + λ0r
		φr = y/R + φ0r
		λd = rtod(λr)
		φd = rtod(φr)
	)
	return geom.XY{X: λd, Y: φd}
}
