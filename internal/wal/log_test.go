package wal

import (
	"testing"

	"geoindex/internal/model"
)

func TestLogAppendAndSince(t *testing.T) {
	log := NewLog()
	first, err := log.Append(model.NewAddOp(0, model.NewPoint("a", 1, 2)))
	if err != nil {
		t.Fatal(err)
	}
	second, err := log.Append(model.NewDeleteOp(0, "a"))
	if err != nil {
		t.Fatal(err)
	}
	if first != 1 || second != 2 {
		t.Fatalf("unexpected sequence numbers: %d %d", first, second)
	}
	ops := log.Since(2)
	if len(ops) != 1 || ops[0].Seq != 2 {
		t.Fatalf("Since(2) returned %+v", ops)
	}
}

func TestLogReplayOrder(t *testing.T) {
	log := NewLog()
	ids := []string{"a", "b", "c"}
	for _, id := range ids {
		if _, err := log.Append(model.NewAddOp(0, model.NewPoint(id, 1, 1))); err != nil {
			t.Fatal(err)
		}
	}
	var seen []string
	if _, err := log.Replay(func(op model.Op) error {
		seen = append(seen, op.Point.ID)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for i, id := range ids {
		if seen[i] != id {
			t.Fatalf("replay order mismatch at %d: %s", i, seen[i])
		}
	}
}

func TestLogTruncate(t *testing.T) {
	log := NewLog()
	for i := 0; i < 4; i++ {
		_, _ = log.Append(model.NewAddOp(0, model.NewPoint("p", float64(i), 0)))
	}
	log.Truncate(3)
	if log.Size() != 2 {
		t.Fatalf("expected 2 entries after truncate, got %d", log.Size())
	}
	if log.Compacted() != 2 {
		t.Fatalf("expected 2 compacted ops, got %d", log.Compacted())
	}
}
