package jts

// IndexStrtree_ItemBoundable is a Boundable wrapper for a non-Boundable spatial
// object. Used internally by AbstractSTRtree.
type IndexStrtree_ItemBoundable struct {
	bounds any
	item   any
}

// IndexStrtree_NewItemBoundable creates a new ItemBoundable wrapping the given
// bounds and item.
func IndexStrtree_NewItemBoundable(bounds, item any) *IndexStrtree_ItemBoundable {
	return &IndexStrtree_ItemBoundable{
		bounds: bounds,
		item:   item,
	}
}

// GetBounds returns the bounds of this ItemBoundable.
func (ib *IndexStrtree_ItemBoundable) GetBounds() any {
	return ib.bounds
}

// TRANSLITERATION NOTE: Marker method for Boundable interface. Not present in
// Java source.
func (ib *IndexStrtree_ItemBoundable) IsIndexStrtree_Boundable() {}

// GetItem returns the item wrapped by this ItemBoundable.
func (ib *IndexStrtree_ItemBoundable) GetItem() any {
	return ib.item
}
