package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/futurematic/kernel/internal/kernel"
	"github.com/futurematic/kernel/internal/query"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	handlers   *Handlers
}

// NewServer creates a new HTTP server
func NewServer(port int, kernelService kernel.Service, queryEngine query.Engine) *Server {
	handlers := NewHandlers(kernelService, queryEngine)

	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/v1/plan", handlers.Plan)
	mux.HandleFunc("/v1/apply", handlers.Apply)
	mux.HandleFunc("/v1/expand", handlers.Expand)
	mux.HandleFunc("/v1/history", handlers.History)
	mux.HandleFunc("/v1/diff", handlers.Diff)
	mux.HandleFunc("/v1/healthz", handlers.Healthz)
	mux.HandleFunc("/v1/namespaces", handlers.Namespaces)
	mux.HandleFunc("/v1/namespace_root", handlers.NamespaceRoot)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		handlers:   handlers,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting kernel server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down kernel server...")
	return s.httpServer.Shutdown(ctx)
}
