package tags

// Tag Tag structure.
type Tag struct {
	Key   string
	Value string
}

// TagDelta Tag delta with to add and to delete tag lists.
type TagDelta struct {
	AddList    []*Tag
	DeleteList []*Tag
}
