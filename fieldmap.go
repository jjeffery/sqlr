package sqlrow

type fieldMap struct {
	prev   *fieldMap
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
	if columnName := fm.fields[fieldName]; columnName != "" {
		return columnName
	}
	if fm.prev != nil {
		return fm.prev.lookup(fieldName)
	}
	return ""
}
