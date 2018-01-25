package registry



func Tables() []interface{} {
	var tables []interface{}
	tables = append(tables, new(ACL), new(Image))
	return tables
}

