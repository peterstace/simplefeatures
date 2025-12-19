package jts

// Geom_Puntal identifies Geometry subclasses which are 0-dimensional and with
// components which are Points.
type Geom_Puntal interface {
	IsPuntal()
}
