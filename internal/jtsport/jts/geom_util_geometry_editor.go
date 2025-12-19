package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomUtil_GeometryEditor is a class which supports creating new Geometries
// which are modifications of existing ones, maintaining the same type
// structure. Geometry objects are intended to be treated as immutable. This
// class "modifies" Geometries by traversing them, applying a user-defined
// GeometryEditorOperation, CoordinateSequenceOperation or CoordinateOperation
// and creating new Geometries with the same structure but (possibly) modified
// components.
//
// Examples of the kinds of modifications which can be made are:
//   - the values of the coordinates may be changed. The editor does not check
//     whether changing coordinate values makes the result Geometry invalid
//   - the coordinate lists may be changed (e.g. by adding, deleting or
//     modifying coordinates). The modified coordinate lists must be consistent
//     with their original parent component (e.g. a LinearRing must always have
//     at least 4 coordinates, and the first and last coordinate must be equal)
//   - components of the original geometry may be deleted (e.g. holes may be
//     removed from a Polygon, or LineStrings removed from a MultiLineString).
//     Deletions will be propagated up the component tree appropriately.
//
// All changes must be consistent with the original Geometry's structure (e.g. a
// Polygon cannot be collapsed into a LineString). If changing the structure is
// required, use a GeometryTransformer.
//
// This class supports creating an edited Geometry using a different
// GeometryFactory via the GeomUtil_NewGeometryEditorWithFactory constructor.
// Examples of situations where this is required is if the geometry is
// transformed to a new SRID and/or a new PrecisionModel.
//
// Usage Notes:
//   - The resulting Geometry is not checked for validity. If validity needs to
//     be enforced, the new Geometry's IsValid method should be called.
//   - By default the UserData of the input geometry is not copied to the result.
type GeomUtil_GeometryEditor struct {
	// factory is the factory used to create the modified Geometry. If nil the
	// GeometryFactory of the input is used.
	factory          *Geom_GeometryFactory
	isUserDataCopied bool
}

// GeomUtil_NewGeometryEditor creates a new GeometryEditor object which will
// create edited Geometries with the same GeometryFactory as the input Geometry.
func GeomUtil_NewGeometryEditor() *GeomUtil_GeometryEditor {
	return &GeomUtil_GeometryEditor{}
}

// GeomUtil_NewGeometryEditorWithFactory creates a new GeometryEditor object
// which will create edited Geometries with the given GeometryFactory.
func GeomUtil_NewGeometryEditorWithFactory(factory *Geom_GeometryFactory) *GeomUtil_GeometryEditor {
	return &GeomUtil_GeometryEditor{
		factory: factory,
	}
}

// SetCopyUserData sets whether the User Data is copied to the edit result. Only
// the object reference is copied.
func (ge *GeomUtil_GeometryEditor) SetCopyUserData(isUserDataCopied bool) {
	ge.isUserDataCopied = isUserDataCopied
}

// Edit edits the input Geometry with the given edit operation. Clients can
// create implementations of GeometryEditorOperation or CoordinateOperation to
// perform required modifications.
func (ge *GeomUtil_GeometryEditor) Edit(geometry *Geom_Geometry, operation GeomUtil_GeometryEditor_GeometryEditorOperation) *Geom_Geometry {
	// Nothing to do.
	if geometry == nil {
		return nil
	}

	result := ge.editInternal(geometry, operation)
	if ge.isUserDataCopied {
		result.SetUserData(geometry.GetUserData())
	}
	return result
}

func (ge *GeomUtil_GeometryEditor) editInternal(geometry *Geom_Geometry, operation GeomUtil_GeometryEditor_GeometryEditorOperation) *Geom_Geometry {
	// If client did not supply a GeometryFactory, use the one from the input Geometry.
	if ge.factory == nil {
		ge.factory = geometry.GetFactory()
	}

	if java.InstanceOf[*Geom_GeometryCollection](geometry) {
		return ge.editGeometryCollection(java.Cast[*Geom_GeometryCollection](geometry), operation).Geom_Geometry
	}

	if java.InstanceOf[*Geom_Polygon](geometry) {
		return ge.editPolygon(java.Cast[*Geom_Polygon](geometry), operation).Geom_Geometry
	}

	if java.InstanceOf[*Geom_Point](geometry) {
		return operation.Edit(geometry, ge.factory)
	}

	if java.InstanceOf[*Geom_LineString](geometry) {
		return operation.Edit(geometry, ge.factory)
	}

	Util_Assert_ShouldNeverReachHereWithMessage("Unsupported Geometry class: " + geometry.GetGeometryType())
	return nil
}

func (ge *GeomUtil_GeometryEditor) editPolygon(polygon *Geom_Polygon, operation GeomUtil_GeometryEditor_GeometryEditorOperation) *Geom_Polygon {
	newPolygon := operation.Edit(polygon.Geom_Geometry, ge.factory)
	if newPolygon == nil {
		newPolygon = ge.factory.CreatePolygon().Geom_Geometry
	}
	newPolygonTyped := java.Cast[*Geom_Polygon](newPolygon)
	if newPolygonTyped.IsEmpty() {
		// RemoveSelectedPlugIn relies on this behaviour. [Jon Aquino]
		return newPolygonTyped
	}

	shell := ge.Edit(newPolygonTyped.GetExteriorRing().Geom_LineString.Geom_Geometry, operation)
	if shell == nil || shell.IsEmpty() {
		// RemoveSelectedPlugIn relies on this behaviour. [Jon Aquino]
		return ge.factory.CreatePolygon()
	}

	var holes []*Geom_LinearRing
	for i := 0; i < newPolygonTyped.GetNumInteriorRing(); i++ {
		hole := ge.Edit(newPolygonTyped.GetInteriorRingN(i).Geom_LineString.Geom_Geometry, operation)
		if hole == nil || hole.IsEmpty() {
			continue
		}
		holes = append(holes, java.Cast[*Geom_LinearRing](hole))
	}

	return ge.factory.CreatePolygonWithLinearRingAndHoles(java.Cast[*Geom_LinearRing](shell), holes)
}

func (ge *GeomUtil_GeometryEditor) editGeometryCollection(collection *Geom_GeometryCollection, operation GeomUtil_GeometryEditor_GeometryEditorOperation) *Geom_GeometryCollection {
	// First edit the entire collection.
	// MD - not sure why this is done - could just check original collection?
	editedGeom := operation.Edit(collection.Geom_Geometry, ge.factory)
	collectionForType := java.Cast[*Geom_GeometryCollection](editedGeom)

	// Edit the component geometries.
	var geometries []*Geom_Geometry
	for i := 0; i < collectionForType.GetNumGeometries(); i++ {
		geometry := ge.Edit(collectionForType.GetGeometryN(i), operation)
		if geometry == nil || geometry.IsEmpty() {
			continue
		}
		geometries = append(geometries, geometry)
	}

	// Use editedGeom for type checks to get the leaf type.
	if java.InstanceOf[*Geom_MultiPoint](editedGeom) {
		points := make([]*Geom_Point, len(geometries))
		for i, g := range geometries {
			points[i] = java.Cast[*Geom_Point](g)
		}
		return ge.factory.CreateMultiPointFromPoints(points).Geom_GeometryCollection
	}
	if java.InstanceOf[*Geom_MultiLineString](editedGeom) {
		lineStrings := make([]*Geom_LineString, len(geometries))
		for i, g := range geometries {
			lineStrings[i] = java.Cast[*Geom_LineString](g)
		}
		return ge.factory.CreateMultiLineStringFromLineStrings(lineStrings).Geom_GeometryCollection
	}
	if java.InstanceOf[*Geom_MultiPolygon](editedGeom) {
		polygons := make([]*Geom_Polygon, len(geometries))
		for i, g := range geometries {
			polygons[i] = java.Cast[*Geom_Polygon](g)
		}
		return ge.factory.CreateMultiPolygonFromPolygons(polygons).Geom_GeometryCollection
	}
	return ge.factory.CreateGeometryCollectionFromGeometries(geometries)
}

// GeomUtil_GeometryEditor_GeometryEditorOperation is an interface which
// specifies an edit operation for Geometries.
type GeomUtil_GeometryEditor_GeometryEditorOperation interface {
	// Edit edits a Geometry by returning a new Geometry with a modification.
	// The returned geometry may be:
	//   - the input geometry itself. The returned Geometry might be the same as
	//     the Geometry passed in.
	//   - nil if the geometry is to be deleted.
	Edit(geometry *Geom_Geometry, factory *Geom_GeometryFactory) *Geom_Geometry
}

// GeomUtil_GeometryEditor_NoOpGeometryOperation is a GeometryEditorOperation
// which does not modify the input geometry. This can be used for simple changes
// of GeometryFactory (including PrecisionModel and SRID).
type GeomUtil_GeometryEditor_NoOpGeometryOperation struct{}

// Edit returns the geometry unchanged.
func (op *GeomUtil_GeometryEditor_NoOpGeometryOperation) Edit(geometry *Geom_Geometry, factory *Geom_GeometryFactory) *Geom_Geometry {
	return geometry
}

// GeomUtil_GeometryEditor_CoordinateOperation is a GeometryEditorOperation
// which edits the coordinate list of a Geometry. Operates on Geometry
// subclasses which contain a single coordinate list.
type GeomUtil_GeometryEditor_CoordinateOperation interface {
	GeomUtil_GeometryEditor_GeometryEditorOperation

	// EditCoordinates edits the array of Coordinates from a Geometry. If it is
	// desired to preserve the immutability of Geometries, if the coordinates
	// are changed a new array should be created and returned.
	EditCoordinates(coordinates []*Geom_Coordinate, geometry *Geom_Geometry) []*Geom_Coordinate
}

// GeomUtil_GeometryEditor_CoordinateOperationBase provides the base
// implementation for CoordinateOperation. Subclasses should embed this and
// implement EditCoordinates.
type GeomUtil_GeometryEditor_CoordinateOperationBase struct {
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (op *GeomUtil_GeometryEditor_CoordinateOperationBase) GetChild() java.Polymorphic {
	return op.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (op *GeomUtil_GeometryEditor_CoordinateOperationBase) GetParent() java.Polymorphic {
	return nil
}

// Edit implements the GeometryEditorOperation interface.
func (op *GeomUtil_GeometryEditor_CoordinateOperationBase) Edit(geometry *Geom_Geometry, factory *Geom_GeometryFactory) *Geom_Geometry {
	// Get the concrete implementation for EditCoordinates.
	impl, ok := java.GetLeaf(op).(interface {
		EditCoordinates([]*Geom_Coordinate, *Geom_Geometry) []*Geom_Coordinate
	})
	if !ok {
		panic("CoordinateOperation implementation must provide EditCoordinates method")
	}

	if java.InstanceOf[*Geom_LinearRing](geometry) {
		return factory.CreateLinearRingFromCoordinates(impl.EditCoordinates(geometry.GetCoordinates(), geometry)).Geom_LineString.Geom_Geometry
	}

	if java.InstanceOf[*Geom_LineString](geometry) {
		return factory.CreateLineStringFromCoordinates(impl.EditCoordinates(geometry.GetCoordinates(), geometry)).Geom_Geometry
	}

	if java.InstanceOf[*Geom_Point](geometry) {
		newCoordinates := impl.EditCoordinates(geometry.GetCoordinates(), geometry)
		if len(newCoordinates) > 0 {
			return factory.CreatePointFromCoordinate(newCoordinates[0]).Geom_Geometry
		}
		return factory.CreatePoint().Geom_Geometry
	}

	return geometry
}

// GeomUtil_GeometryEditor_CoordinateSequenceOperation is a
// GeometryEditorOperation which edits the CoordinateSequence of a Geometry.
// Operates on Geometry subclasses which contain a single coordinate list.
type GeomUtil_GeometryEditor_CoordinateSequenceOperation interface {
	GeomUtil_GeometryEditor_GeometryEditorOperation

	// EditCoordinateSequence edits a CoordinateSequence from a Geometry.
	EditCoordinateSequence(coordSeq Geom_CoordinateSequence, geometry *Geom_Geometry) Geom_CoordinateSequence
}

// GeomUtil_GeometryEditor_CoordinateSequenceOperationBase provides the base
// implementation for CoordinateSequenceOperation. Subclasses should embed this
// and implement EditCoordinateSequence.
type GeomUtil_GeometryEditor_CoordinateSequenceOperationBase struct {
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (op *GeomUtil_GeometryEditor_CoordinateSequenceOperationBase) GetChild() java.Polymorphic {
	return op.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (op *GeomUtil_GeometryEditor_CoordinateSequenceOperationBase) GetParent() java.Polymorphic {
	return nil
}

// Edit implements the GeometryEditorOperation interface.
func (op *GeomUtil_GeometryEditor_CoordinateSequenceOperationBase) Edit(geometry *Geom_Geometry, factory *Geom_GeometryFactory) *Geom_Geometry {
	// Get the concrete implementation for EditCoordinateSequence.
	impl, ok := java.GetLeaf(op).(interface {
		EditCoordinateSequence(Geom_CoordinateSequence, *Geom_Geometry) Geom_CoordinateSequence
	})
	if !ok {
		panic("CoordinateSequenceOperation implementation must provide EditCoordinateSequence method")
	}

	if java.InstanceOf[*Geom_LinearRing](geometry) {
		lr := java.Cast[*Geom_LinearRing](geometry)
		return factory.CreateLinearRingFromCoordinateSequence(impl.EditCoordinateSequence(lr.GetCoordinateSequence(), geometry)).Geom_LineString.Geom_Geometry
	}

	if java.InstanceOf[*Geom_LineString](geometry) {
		ls := java.Cast[*Geom_LineString](geometry)
		return factory.CreateLineStringFromCoordinateSequence(impl.EditCoordinateSequence(ls.GetCoordinateSequence(), geometry)).Geom_Geometry
	}

	if java.InstanceOf[*Geom_Point](geometry) {
		pt := java.Cast[*Geom_Point](geometry)
		return factory.CreatePointFromCoordinateSequence(impl.EditCoordinateSequence(pt.GetCoordinateSequence(), geometry)).Geom_Geometry
	}

	return geometry
}
