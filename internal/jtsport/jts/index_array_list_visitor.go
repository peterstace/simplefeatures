package jts

// Index_ArrayListVisitor builds a slice of all visited items.
type Index_ArrayListVisitor struct {
	items []any
}

var _ Index_ItemVisitor = (*Index_ArrayListVisitor)(nil)

func (alv *Index_ArrayListVisitor) IsIndex_ItemVisitor() {}

// Index_NewArrayListVisitor creates a new ArrayListVisitor.
func Index_NewArrayListVisitor() *Index_ArrayListVisitor {
	return &Index_ArrayListVisitor{items: []any{}}
}

// VisitItem visits an item and adds it to the collection.
func (alv *Index_ArrayListVisitor) VisitItem(item any) {
	alv.items = append(alv.items, item)
}

// GetItems gets the slice of visited items.
func (alv *Index_ArrayListVisitor) GetItems() []any {
	return alv.items
}
