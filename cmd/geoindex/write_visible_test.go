package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWriteThenRangeImmediatelyVisible reproduces the reported regression: a
// point written through the HTTP surface must be returned by a range query
// issued immediately afterwards, without a process restart. Before the fix
// AddPoint only touched the WAL and the range tree while the query service
// intersects grid and tree candidates, so freshly written points were filtered
// out until a replay/rebuild repopulated the grid.
func TestWriteThenRangeImmediatelyVisible(t *testing.T) {
	server, _ := buildServer()
	ts := httptest.NewServer(server.Routes())
	defer ts.Close()

	point := map[string]any{"id": "probe-point", "x": 1.5, "y": 2.5, "category": "probe"}
	payload, _ := json.Marshal(point)
	resp, err := http.Post(ts.URL+"/api/v1/points", "application/json", bytesReader(payload))
	if err != nil {
		t.Fatalf("add point: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("add point returned %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp, err = http.Get(ts.URL + "/api/v1/query/range?minx=0&miny=0&maxx=10&maxy=10")
	if err != nil {
		t.Fatalf("post-write range: %v", err)
	}
	var result struct {
		Count float64 `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode range response: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK || result.Count < 1 {
		t.Fatalf("post-write range returned %d with count %v, want count >= 1", resp.StatusCode, result.Count)
	}
}
