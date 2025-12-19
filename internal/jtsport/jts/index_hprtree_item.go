package jts

import "fmt"

// IndexHprtree_Item wraps an envelope and item for HPRtree storage.
type IndexHprtree_Item struct {
	env  *Geom_Envelope
	item any
}

// IndexHprtree_NewItem creates a new Item with the given envelope and item.
func IndexHprtree_NewItem(env *Geom_Envelope, item any) *IndexHprtree_Item {
	return &IndexHprtree_Item{
		env:  env,
		item: item,
	}
}

// GetEnvelope returns the envelope of this item.
func (i *IndexHprtree_Item) GetEnvelope() *Geom_Envelope {
	return i.env
}

// GetItem returns the item.
func (i *IndexHprtree_Item) GetItem() any {
	return i.item
}

// String returns a string representation of this item.
func (i *IndexHprtree_Item) String() string {
	return fmt.Sprintf("Item: %s", i.env.String())
}
