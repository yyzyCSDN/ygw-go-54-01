package write

import "sort"

func sortStrings(values []string) {
	sort.Strings(values)
}

func joinStrings(values []string, separator string) string {
	output := ""
	for index, value := range values {
		if index > 0 {
			output += separator
		}
		output += value
	}
	return output
}
