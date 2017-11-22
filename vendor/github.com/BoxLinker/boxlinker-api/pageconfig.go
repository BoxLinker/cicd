package boxlinker

type PageConfig struct {
	CurrentPage int
	PageCount   int
	TotalCount  int
}

func (pc PageConfig) Limit() int {
	return pc.PageCount
}
func (pc PageConfig) Offset() int {
	return pc.PageCount * (pc.CurrentPage - 1)
}

func (pc PageConfig) PaginationJSON() map[string]int {
	m := map[string]int{}
	m["currentPage"] = pc.CurrentPage
	m["pageCount"] = pc.PageCount
	m["totalCount"] = pc.TotalCount
	return m
}

func (pc PageConfig) PaginationResult(output interface{}) map[string]interface{} {
	return map[string]interface{}{
		"pagination": pc.PaginationJSON(),
		"data": output,
	}
}

func (pc PageConfig) FormatOutput(output interface{}) map[string]interface{} {
	return map[string]interface{}{
		"pagination": pc.PaginationJSON(),
		"data":       output,
	}
}