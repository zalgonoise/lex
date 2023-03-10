package lex

// Item represents a set of any type of tokens identified by a comparable type
type Item[T comparable, V any] struct {
	Pos   int
	Type  T
	Value []V
}

// NewItem creates an Item with type `T` and values `[]V`
func NewItem[T comparable, V any](pos int, itemType T, value ...V) Item[T, V] {
	return Item[T, V]{
		Pos:   pos,
		Type:  itemType,
		Value: value,
	}
}
