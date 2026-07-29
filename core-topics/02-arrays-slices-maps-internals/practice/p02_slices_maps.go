package practice

// SafeAppend3Index creates a sub-slice with restricted capacity to force reallocation on append.
func SafeAppend3Index(src []int, start, end int, valToAppend int) []int {
	sub := src[start:end:end] // Cap restricted to end - start
	return append(sub, valToAppend)
}

// InvertMap swaps keys and values of a map[string]int.
func InvertMap(m map[string]int) map[int]string {
	res := make(map[int]string, len(m))
	for k, v := range m {
		res[v] = k
	}
	return res
}
