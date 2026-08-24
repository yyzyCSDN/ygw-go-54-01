package rtree

// Union returns all distinct point ids currently indexed by the tree.
func (t *RTree) Union() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	seen := make(map[string]struct{})
	var out []string
	var walk func(*node)
	walk = func(current *node) {
		if current == nil {
			return
		}
		if current.leaf {
			for _, point := range current.points {
				if _, ok := seen[point.ID]; ok {
					continue
				}
				seen[point.ID] = struct{}{}
				out = append(out, point.ID)
			}
			return
		}
		for _, child := range current.children {
			walk(child)
		}
	}
	walk(t.root)
	return out
}
