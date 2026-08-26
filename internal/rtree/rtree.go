package rtree

import (
	"sync"

	"geoindex/internal/model"
)

// RTree is a small in-memory R-tree used as an approximate secondary index
// over every indexed point. It supports inclusive range search and point
// removal, and every mutation is protected by a read-write lock.
type RTree struct {
	mu   sync.RWMutex
	root *node
	size int
}

// NewRTree builds an empty tree.
func NewRTree() *RTree {
	return &RTree{}
}

// Insert adds a point to the tree.
func (t *RTree) Insert(point model.Point) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.root == nil {
		t.root = newNode(true)
	}
	t.root.insert(point)
	t.size++
}

// Remove deletes a point by id and returns whether it existed.
func (t *RTree) Remove(id string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.root == nil {
		return false
	}
	if t.root.remove(id) {
		t.size--
		return true
	}
	return false
}

// RangeSearch returns all points whose coordinates fall inside the rectangle.
// The read lock is held for the whole traversal and released on return.
func (t *RTree) RangeSearch(rect model.Rect) []model.Point {
	t.mu.RLock()
	var out []model.Point
	t.searchLocked(t.root, rect, &out)
	return out
}

// Search returns the indexed points inside the rectangle. It is the query
// service entry point and always releases the read lock before returning.
func (t *RTree) Search(rect model.Rect) []model.Point {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []model.Point
	t.searchLocked(t.root, rect, &out)
	return out
}

func (t *RTree) searchLocked(current *node, rect model.Rect, out *[]model.Point) {
	if current == nil || !current.bounds.Overlaps(rect) {
		return
	}
	if current.leaf {
		for _, point := range current.points {
			if rect.ContainsPoint(point) {
				*out = append(*out, point)
			}
		}
		return
	}
	for _, child := range current.children {
		t.searchLocked(child, rect, out)
	}
}

// Size returns the number of indexed points.
func (t *RTree) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.size
}

// Depth returns the tree height or zero for an empty tree.
func (t *RTree) Depth() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.root == nil {
		return 0
	}
	return t.root.depth()
}

// Rebuild replaces the whole tree with the given point set. It is used by the
// index rebuilder after a full snapshot reconstruction.
func (t *RTree) Rebuild(points []model.Point) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.root = newNode(true)
	t.size = 0
	for _, point := range points {
		t.root.insert(point)
		t.size++
	}
}
