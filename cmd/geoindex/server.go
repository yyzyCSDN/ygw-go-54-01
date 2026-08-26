package main

import (
	"net/http"

	"geoindex/internal/grid"
	"geoindex/internal/query"
	"geoindex/internal/rebuild"
	"geoindex/internal/tile"
	"geoindex/internal/write"
)

// Server owns the HTTP surface of the GeoIndex service.
type Server struct {
	write    *write.Service
	query    *query.Service
	rebuild  *rebuild.Rebuilder
	renderer *tile.Renderer
	grid     *grid.Grid
}

// NewServer wires the HTTP server onto the application services.
func NewServer(
	writeService *write.Service,
	queryService *query.Service,
	rebuilder *rebuild.Rebuilder,
	renderer *tile.Renderer,
	g *grid.Grid,
) *Server {
	return &Server{
		write:    writeService,
		query:    queryService,
		rebuild:  rebuilder,
		renderer: renderer,
		grid:     g,
	}
}

// Routes assembles the HTTP route table.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/v1/points", s.handleAddPoint)
	mux.HandleFunc("/api/v1/points/delete", s.handleDeletePoint)
	mux.HandleFunc("/api/v1/polygons", s.handleUpdatePolygon)
	mux.HandleFunc("/api/v1/query/range", s.handleRangeQuery)
	mux.HandleFunc("/api/v1/query/nearest", s.handleNearestQuery)
	mux.HandleFunc("/api/v1/tiles/", s.handleTile)
	mux.HandleFunc("/api/v1/tiles/invalidate", s.handleInvalidateTiles)
	mux.HandleFunc("/api/v1/rebuild", s.handleRebuild)
	mux.HandleFunc("/api/v1/stats", s.handleStats)
	return logRequests(mux)
}

// logRequests wraps the mux with a minimal access log.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-GeoIndex", "1")
		next.ServeHTTP(w, r)
	})
}
