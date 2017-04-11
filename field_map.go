package sqlrow

// fieldMap is used to lookup column names associated with fields.
// There is no mutex because once a schema has been initialized, its
// field map should be immutable.
//
// When a schema is cloned from another schema, its fieldMap points to
// the fieldMap of the previous schema.
type fieldMap struct {
	prev   *fieldMap // points to the fieldMap for the previous
	fields map[string]string
}

func newFieldMap(prev *fieldMap) *fieldMap {
	return &fieldMap{
		prev:   prev,
		fields: make(map[string]string),
	}
}

func (fm *fieldMap) add(fieldName string, columnName string) {
	fm.fields[fieldName] = columnName
}

func (fm *fieldMap) lookup(fieldName string) string {
	if columnName, ok := fm.fields[fieldName]; ok {
		return columnName
	}
	if fm.prev != nil {
		return fm.prev.lookup(fieldName)
	}
	return ""
}
