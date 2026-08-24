package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// runProbe starts the server on the given address, exercises the core HTTP
// surface and shuts down again. It is used by the startup verification flow.
func runProbe(addr string) error {
	server, _ := buildServer()
	httpServer := &http.Server{Addr: addr, Handler: server.Routes()}
	listener, err := netListen(addr)
	if err != nil {
		return err
	}
	go func() {
		_ = httpServer.Serve(listener)
	}()

	base := fmt.Sprintf("http://%s", listener.Addr().String())
	defer func() {
		_ = httpServer.Close()
	}()

	checks := []struct {
		name string
		url  string
	}{
		{"health", base + "/healthz"},
		{"index", base + "/"},
		{"range", base + "/api/v1/query/range?minx=0&miny=0&maxx=10&maxy=10"},
		{"nearest", base + "/api/v1/query/nearest?x=5&y=5&k=3"},
		{"tile", base + "/api/v1/tiles/0/0/0"},
		{"stats", base + "/api/v1/stats"},
	}
	for _, check := range checks {
		response, err := http.Get(check.url)
		if err != nil {
			return fmt.Errorf("probe %s: %w", check.name, err)
		}
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		_ = response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("probe %s returned %d: %s", check.name, response.StatusCode, body)
		}
	}

	point := map[string]any{"id": "probe-point", "x": 1.5, "y": 2.5, "category": "probe"}
	payload, _ := json.Marshal(point)
	response, err := http.Post(base+"/api/v1/points", "application/json", bytesReader(payload))
	if err != nil {
		return fmt.Errorf("probe add point: %w", err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("probe add point returned %d", response.StatusCode)
	}

	// 数据写入后再查一次，确认索引立即可见。
	response, err = http.Get(base + "/api/v1/query/range?minx=0&miny=0&maxx=10&maxy=10")
	if err != nil {
		return fmt.Errorf("probe post-write range: %w", err)
	}
	var result map[string]any
	_ = json.NewDecoder(response.Body).Decode(&result)
	_ = response.Body.Close()
	count, _ := result["count"].(float64)
	if response.StatusCode != http.StatusOK || count < 1 {
		return fmt.Errorf("probe post-write range returned %d with count %v", response.StatusCode, result["count"])
	}

	return nil
}
