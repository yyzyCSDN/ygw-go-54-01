package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"geoindex/internal/grid"
	"geoindex/internal/query"
	"geoindex/internal/rebuild"
	"geoindex/internal/rtree"
	"geoindex/internal/tile"
	"geoindex/internal/wal"
	"geoindex/internal/write"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	probe := flag.Bool("probe", false, "start, self probe and exit")
	flag.Parse()

	if *probe {
		if err := runProbe(*addr); err != nil {
			log.Fatalf("probe failed: %v", err)
		}
		return
	}

	server, _ := buildServer()
	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           server.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
	}()

	log.Printf("geoindex listening on %s", *addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

// buildServer wires the shared indexes and services.
func buildServer() (*Server, *write.Service) {
	g := grid.NewGrid(8, 64)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	renderer := tile.NewRenderer(g, []int{0, 1, 2, 3})
	writeService := write.NewService(g, tree, log, write.Options{Tiles: renderer})
	queryService := query.NewService(g, tree)
	rebuilder := rebuild.NewRebuilder(g, tree, log)
	return NewServer(writeService, queryService, rebuilder, renderer, g), writeService
}
