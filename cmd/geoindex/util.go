package main

import "errors"

var errBadTilePath = errors.New("tile path must be /api/v1/tiles/{z}/{x}/{y}")

// splitPath splits a slash separated path into non-empty segments.
func splitPath(path string) []string {
	var parts []string
	start := -1
	for index := 0; index <= len(path); index++ {
		if index == len(path) || path[index] == '/' {
			if start >= 0 {
				parts = append(parts, path[start:index])
				start = -1
			}
			continue
		}
		if start < 0 {
			start = index
		}
	}
	return parts
}
