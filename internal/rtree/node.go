package rtree

import "geoindex/internal/model"

// nodeCapacity is the maximum number of points in one leaf node before a
// split is triggered.
const nodeCapacity = 16

// node is one node of the range tree. Internal nodes own children and track
// the union of their bounds; leaf nodes own points directly.
type node struct {
	leaf     bool
	bounds   model.Rect
	points   []model.Point
	children []*node
}

func newNode(leaf bool) *node {
	return &node{leaf: leaf}
}

// insert adds a point to the subtree and refreshes the node bounds.
func (n *node) insert(point model.Point) {
	n.bounds = n.bounds.Union(model.PointRect(point))
	if n.leaf {
		n.points = append(n.points, point)
		if len(n.points) > nodeCapacity {
			n.split()
		}
		return
	}
	child := n.bestChild(point)
	child.insert(point)
}

// bestChild selects the child whose bounds grow least when extended to the
// point; ties are broken by area.
func (n *node) bestChild(point model.Point) *node {
	best := n.children[0]
	bestGrowth := best.growth(point)
	for _, child := range n.children[1:] {
		growth := child.growth(point)
		if growth < bestGrowth {
			best = child
			bestGrowth = growth
		}
	}
	return best
}

func (n *node) growth(point model.Point) float64 {
	extended := n.bounds.Union(model.PointRect(point))
	return extended.Area() - n.bounds.Area()
}

// split promotes the node from leaf to internal and distributes its points
// among two children selected by maximum bounding area heuristic.
func (n *node) split() {
	first, second := pickSeeds(n.points)
	left := newNode(true)
	right := newNode(true)
	left.insert(first)
	right.insert(second)
	for _, point := range n.points {
		if point.ID == first.ID || point.ID == second.ID {
			continue
		}
		if left.growth(point) < right.growth(point) {
			left.insert(point)
		} else {
			right.insert(point)
		}
	}
	n.leaf = false
	n.points = nil
	n.children = []*node{left, right}
}

// pickSeeds returns the two points whose combined rectangle has the largest
// area, used to initialize the split.
func pickSeeds(points []model.Point) (model.Point, model.Point) {
	first, second := points[0], points[1]
	worst := -1.0
	for i := 0; i < len(points); i++ {
		for j := i + 1; j < len(points); j++ {
			combined := model.PointRect(points[i]).Union(model.PointRect(points[j]))
			if area := combined.Area(); area > worst {
				worst = area
				first, second = points[i], points[j]
			}
		}
	}
	return first, second
}

// remove deletes a point by id from the subtree and reports success.
func (n *node) remove(id string) bool {
	if n.leaf {
		for index, point := range n.points {
			if point.ID != id {
				continue
			}
			n.points = append(n.points[:index], n.points[index+1:]...)
			n.recomputeBounds()
			return true
		}
		return false
	}
	for _, child := range n.children {
		if child.remove(id) {
			n.recomputeBounds()
			return true
		}
	}
	return false
}

// recomputeBounds recalculates the node bounding box from its contents.
func (n *node) recomputeBounds() {
	if n.leaf {
		n.bounds = model.Rect{}
		for _, point := range n.points {
			n.bounds = n.bounds.Union(model.PointRect(point))
		}
		return
	}
	n.bounds = model.Rect{}
	for _, child := range n.children {
		n.bounds = n.bounds.Union(child.bounds)
	}
}

// depth returns the maximum child depth below this node.
func (n *node) depth() int {
	if n.leaf {
		return 0
	}
	maxDepth := 0
	for _, child := range n.children {
		if childDepth := child.depth(); childDepth > maxDepth {
			maxDepth = childDepth
		}
	}
	return maxDepth + 1
}

// count returns the number of points stored below this node.
func (n *node) count() int {
	if n.leaf {
		return len(n.points)
	}
	total := 0
	for _, child := range n.children {
		total += child.count()
	}
	return total
}
