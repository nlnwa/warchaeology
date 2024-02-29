package filewalker

// StringSet is a set of strings
type StringSet map[string]struct{}

// NewStringSet creates a new StringSet
func NewStringSet() StringSet {
	return make(StringSet)
}

// Add adds a string to the set
func (s StringSet) Add(path string) {
	s[path] = struct{}{}
}

// Contains checks if a string is in the set
func (s StringSet) Contains(path string) bool {
	_, ok := s[path]
	return ok
}
