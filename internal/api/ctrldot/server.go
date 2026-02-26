package ctrldot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/futurematic/kernel/internal/ctrldot"
	"github.com/futurematic/kernel/internal/ledger/autobundle"
)

// recoveryMiddleware recovers from handler panics and returns 500 instead of closing the connection.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic in handler %s %s: %v\n%s", r.Method, r.URL.Path, err, debug.Stack())
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Server represents the Ctrl Dot HTTP server
type Server struct {
	httpServer    *http.Server
	handlers      *Handlers
	autobundleMgr *autobundle.Manager
}

// NewServer creates a new Ctrl Dot HTTP server. autobundleMgr may be nil.
func NewServer(port int, ctrldotService ctrldot.Service, autobundleMgr *autobundle.Manager) *Server {
	handlers := NewHandlers(ctrldotService, autobundleMgr)

	mux := http.NewServeMux()

	// UI routes (served from embedded files or filesystem)
	mux.HandleFunc("/ui", handlers.ServeUI)
	mux.HandleFunc("/ui/", handlers.ServeUI)

	// Register API routes
	mux.HandleFunc("/v1/health", handlers.Health)
	mux.HandleFunc("/v1/agents/register", handlers.RegisterAgent)
	mux.HandleFunc("/v1/agents", handlers.ListAgents)
	mux.HandleFunc("/v1/agents/", handlers.AgentByID)
	mux.HandleFunc("/v1/sessions/start", handlers.StartSession)
	mux.HandleFunc("/v1/sessions/", handlers.SessionByID)
	mux.HandleFunc("/v1/actions/propose", handlers.ProposeAction)
	mux.HandleFunc("/v1/events", handlers.GetEvents)
	mux.HandleFunc("/v1/events/", handlers.GetEvent)
	mux.HandleFunc("/v1/panic/on", handlers.PanicOn)
	mux.HandleFunc("/v1/panic/off", handlers.PanicOff)
	mux.HandleFunc("/v1/panic", handlers.PanicStatus)
	mux.HandleFunc("/v1/autobundle", handlers.AutobundleStatus)
	mux.HandleFunc("/v1/autobundle/test", handlers.AutobundleTest)
	mux.HandleFunc("/v1/capabilities", handlers.Capabilities)
	mux.HandleFunc("/v1/limits/config", handlers.LimitsConfig)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      recoveryMiddleware(mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer:    httpServer,
		handlers:      handlers,
		autobundleMgr: autobundleMgr,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting Ctrl Dot server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server. Calls MaybeBundleOnShutdown before closing.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down Ctrl Dot server...")
	if s.autobundleMgr != nil {
		if path, err := s.autobundleMgr.MaybeBundleOnShutdown(ctx); err != nil {
			log.Printf("autobundle on shutdown: %v", err)
		} else if path != "" {
			log.Printf("Shutdown bundle written: %s", path)
		}
	}
	return s.httpServer.Shutdown(ctx)
}
