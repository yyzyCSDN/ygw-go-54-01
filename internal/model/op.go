package model

// OpKind identifies the type of a spatial change recorded in the write-ahead
// log. Kinds are ordered and stable so that replay is deterministic.
type OpKind uint8

const (
	OpAddPoint OpKind = iota + 1
	OpDeletePoint
	OpUpdatePolygon
)

// Op is a single durable spatial change. The sequence number is assigned by
// the write-ahead log and is unique within one process lifetime.
type Op struct {
	Seq     uint64
	Kind    OpKind
	Point   Point
	Polygon Polygon
	Deleted string
}

// NewAddOp builds an add-point operation.
func NewAddOp(seq uint64, point Point) Op {
	return Op{Seq: seq, Kind: OpAddPoint, Point: point}
}

// NewDeleteOp builds a delete-point operation.
func NewDeleteOp(seq uint64, id string) Op {
	return Op{Seq: seq, Kind: OpDeletePoint, Deleted: id}
}

// NewPolygonOp builds a polygon update operation.
func NewPolygonOp(seq uint64, polygon Polygon) Op {
	return Op{Seq: seq, Kind: OpUpdatePolygon, Polygon: polygon}
}
