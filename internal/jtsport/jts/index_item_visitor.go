package jts

// Index_ItemVisitor is a visitor for items in a SpatialIndex.
type Index_ItemVisitor interface {
	// VisitItem visits an item in the index.
	VisitItem(item any)

	// IsIndex_ItemVisitor is a marker method for interface identification.
	IsIndex_ItemVisitor()
}

// TRANSLITERATION NOTE: Index_NewItemVisitorFunc and index_funcVisitor provide
// a Go convenience for creating simple visitors from functions. Not present in
// Java source.

// Index_NewItemVisitorFunc creates an ItemVisitor from a function.
// This is useful for simple visitors that don't need their own type.
func Index_NewItemVisitorFunc(visitFn func(item any)) Index_ItemVisitor {
	return &index_funcVisitor{visitFn: visitFn}
}

type index_funcVisitor struct {
	visitFn func(item any)
}

var _ Index_ItemVisitor = (*index_funcVisitor)(nil)

func (fv *index_funcVisitor) IsIndex_ItemVisitor() {}

func (fv *index_funcVisitor) VisitItem(item any) {
	fv.visitFn(item)
}
