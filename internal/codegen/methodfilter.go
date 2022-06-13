package codegen

// MethodFilter is a filter for methods to be generated.
type MethodFilter struct {
	filters []string
}

// NewMethodFilter creates a new MethodFilter.
func NewMethodFilter(filters []string) *MethodFilter {
	return &MethodFilter{filters}
}

// Filter returns true if the method matches one of the filters.
// In case no filter matches the method, the method is allowed.
func (md *MethodFilter) Filter(method string) bool {
	for _, filter := range md.filters {
		result := true
		if filter[0] == '!' {
			filter = filter[1:]
			result = false
		}

		if filter == "*" || filter == method {
			return result
		}
	}
	return true // everything matches by default
}
