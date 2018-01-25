package application

func Tables() []interface{} {
	var tables []interface{}
	tables = append(tables, new(PodConfigure))
	return tables
}


