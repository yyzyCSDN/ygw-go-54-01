package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"geoindex/internal/model"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("json encode failed: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/map.html" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "web/map.html")
}

type pointRequest struct {
	ID       string            `json:"id"`
	X        float64           `json:"x"`
	Y        float64           `json:"y"`
	Category string            `json:"category"`
	Weight   float64           `json:"weight"`
	Meta     map[string]string `json:"meta"`
}

func (s *Server) handleAddPoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request pointRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	point := model.Point{
		ID:       request.ID,
		X:        request.X,
		Y:        request.Y,
		Category: request.Category,
		Weight:   request.Weight,
		Meta:     request.Meta,
	}
	if err := s.write.AddPoint(point); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": point.ID, "status": "indexed"})
}

type deleteRequest struct {
	ID string `json:"id"`
}

func (s *Server) handleDeletePoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request deleteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := s.write.DeletePoint(request.ID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": request.ID, "status": "deleted"})
}

type polygonRequest struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Vertices []model.Point     `json:"vertices"`
	Tags     []string          `json:"tags"`
	Extra    map[string]string `json:"extra"`
}

func (s *Server) handleUpdatePolygon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var request polygonRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	polygon := model.Polygon{
		ID:       request.ID,
		Name:     request.Name,
		Vertices: request.Vertices,
		Tags:     request.Tags,
	}
	if err := s.write.UpdatePolygon(polygon); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": polygon.ID, "status": "committed"})
}

func (s *Server) handleRangeQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	values := r.URL.Query()
	rect, err := rectFromQuery(values.Get("minx"), values.Get("miny"), values.Get("maxx"), values.Get("maxy"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	points, err := s.query.Range(rect)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.query.Stats().ObserveServed(len(points))
	writeJSON(w, http.StatusOK, map[string]any{"count": len(points), "points": points})
}

func (s *Server) handleNearestQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	values := r.URL.Query()
	x, err := strconv.ParseFloat(values.Get("x"), 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid x")
		return
	}
	y, err := strconv.ParseFloat(values.Get("y"), 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid y")
		return
	}
	k := 5
	if raw := values.Get("k"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			k = parsed
		}
	}
	points, err := s.query.Nearest(x, y, k)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": len(points), "points": points})
}

func (s *Server) handleTile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	z, x, y, err := tileCoords(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	payload, err := s.renderer.Render(z, x, y)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(payload)
}

func (s *Server) handleInvalidateTiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	cleared := s.renderer.InvalidateAll()
	writeJSON(w, http.StatusOK, map[string]any{"status": "invalidated", "cleared": cleared})
}

func (s *Server) handleRebuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := s.rebuild.Rebuild(); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "rebuilt", "stats": s.rebuild.Stats()})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"grid":         s.grid.Metrics(),
		"query":        s.query.Stats().Snapshot(),
		"tiles":        map[string]any{"cached": s.renderer.CacheSize(), "hitRate": s.renderer.HitRate(), "jobs": s.renderer.JobStats()},
		"polygons":     s.write.PolygonSummary(),
		"rebuild":      s.rebuild.Stats(),
		"pendingDelta": s.rebuild.PendingDelta(),
		"snapshot":     s.rebuild.SnapshotCount(),
		"walEntries":   s.write.Log().Size(),
		"rtreeIds":     len(s.write.Tree().Union()),
		"regionCells":  len(s.query.Grid().CellsForRect(model.WorldExtent())),
	})
}

func rectFromQuery(minX, minY, maxX, maxY string) (model.Rect, error) {
	parse := func(name, raw string) (float64, error) {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return 0, err
		}
		return value, nil
	}
	loX, err := parse("minx", minX)
	if err != nil {
		return model.Rect{}, err
	}
	loY, err := parse("miny", minY)
	if err != nil {
		return model.Rect{}, err
	}
	hiX, err := parse("maxx", maxX)
	if err != nil {
		return model.Rect{}, err
	}
	hiY, err := parse("maxy", maxY)
	if err != nil {
		return model.Rect{}, err
	}
	return model.Rect{MinX: loX, MinY: loY, MaxX: hiX, MaxY: hiY}, nil
}

func tileCoords(path string) (int, int, int, error) {
	const prefix = "/api/v1/tiles/"
	rest := path[len(prefix):]
	parts := splitPath(rest)
	if len(parts) != 3 {
		return 0, 0, 0, errBadTilePath
	}
	z, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, errBadTilePath
	}
	x, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, errBadTilePath
	}
	y, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, errBadTilePath
	}
	return z, x, y, nil
}
