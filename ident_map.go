package sqlr

// identMap is used to lookup identifiers that need to be replaced.
// There is no mutex because once a schema has been initialized, its
// identifier map should be immutable.
//
// When a schema is cloned from another schema, its identMap points to
// the identMap of the previous schema.
type identMap struct {
	prev        *identMap // points to the identMap for the previous schema
	identifiers map[string]string
}

func newIdentMap(prev *identMap) *identMap {
	return &identMap{
		prev:        prev,
		identifiers: make(map[string]string),
	}
}

func (im *identMap) add(identifier string, replacement string) {
	im.identifiers[identifier] = replacement
}

// lookup the field name in the field map and return a column name.
// The boolean returns true if there was a match. If there is a match
// and the string is empty, this means to fallback to the naming convention.
func (im *identMap) lookup(identifier string) (string, bool) {
	if replacement, ok := im.identifiers[identifier]; ok {
		return replacement, ok
	}
	if im.prev != nil {
		return im.prev.lookup(identifier)
	}
	return "", false
}
