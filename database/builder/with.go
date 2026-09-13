package builder

// With registers relationships to be eager loaded when the query returns
// results. Nested relationships can be loaded by separating them with dots.
//
//	builder.From[User]().With("posts", "posts.comments")
func (b *ModelBuilder[T]) With(withs ...string) *ModelBuilder[T] {
	b.withs = append(b.withs, withs...)
	return b
}
