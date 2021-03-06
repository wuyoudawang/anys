package iterator

type IteratorSeeker interface {
	// First moves the iterator to the first key/value pair. If the iterator
	// only contains one key/value pair then First and Last whould moves
	// to the same key/value pair.
	// It returns whether such pair exist.
	First()

	// Last moves the iterator to the last key/value pair. If the iterator
	// only contains one key/value pair then First and Last whould moves
	// to the same key/value pair.
	// It returns whether such pair exist.
	Last()

	// Seek moves the iterator to the first key/value pair whose key is greater
	// than or equal to the given key.
	// It returns whether such pair exist.
	//
	// It is safe to modify the contents of the argument after Seek returns.
	Seek(key []byte)

	// Next moves the iterator to the next key/value pair.
	// It returns whether the iterator is exhausted.
	Next()

	// Prev moves the iterator to the previous key/value pair.
	// It returns whether the iterator is exhausted.
	Prev()
}

type CommonIterator interface {
	IteratorSeeker

	// util.Releaser is the interface that wraps basic Release method.
	// When called Release will releases any resources associated with the
	// iterator.
	// util.Releaser

	// util.ReleaseSetter is the interface that wraps the basic SetReleaser
	// method.
	// util.ReleaseSetter

	// TODO: Remove this when ready.
	Valid() bool

	//error
	Error() error
}

type Interface interface {
	CommonIterator

	// Key returns the key of the current key/value pair, or nil if done.
	// The caller should not modify the contents of the returned slice, and
	// its contents may change on the next call to any 'seeks method'.
	Key() []byte

	// Value returns the key of the current key/value pair, or nil if done.
	// The caller should not modify the contents of the returned slice, and
	// its contents may change on the next call to any 'seeks method'.
	Value() []byte
}

type emptyIterator struct {
	err error
}

func (*emptyIterator) Valid() bool            { return false }
func (i *emptyIterator) First() bool          { return false }
func (i *emptyIterator) Last() bool           { return false }
func (i *emptyIterator) Seek(key []byte) bool { return false }
func (i *emptyIterator) Next() bool           { return false }
func (i *emptyIterator) Prev() bool           { return false }
func (*emptyIterator) Key() []byte            { return nil }
func (*emptyIterator) Value() []byte          { return nil }
func (i *emptyIterator) Error() error         { return i.err }

// NewEmptyIterator creates an empty iterator. The err parameter can be
// nil, but if not nil the given err will be returned by Error method.
func NewEmptyIterator(err error) Iterator {
	return &emptyIterator{err: err}
}
