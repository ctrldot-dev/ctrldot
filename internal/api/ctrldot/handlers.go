package ctrldot

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/futurematic/kernel/internal/ctrldot"
	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/ledger/autobundle"
)

// Handlers contains HTTP handlers for Ctrl Dot API
type Handlers struct {
	service       ctrldot.Service
	autobundleMgr *autobundle.Manager
}

// NewHandlers creates new HTTP handlers. autobundleMgr may be nil.
func NewHandlers(service ctrldot.Service, autobundleMgr *autobundle.Manager) *Handlers {
	return &Handlers{
		service:       service,
		autobundleMgr: autobundleMgr,
	}
}

// Health handles GET /v1/health
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	respondJSON(w, map[string]interface{}{
		"ok":      true,
		"version": "0.1.0",
	}, http.StatusOK)
}

// RegisterAgent handles POST /v1/agents/register
func (h *Handlers) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AgentID     string `json:"agent_id"`
		DisplayName string `json:"display_name"`
		DefaultMode string `json:"default_mode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	agent, err := h.service.RegisterAgent(r.Context(), req.AgentID, req.DisplayName, req.DefaultMode)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, agent, http.StatusOK)
}

// ListAgents handles GET /v1/agents
func (h *Handlers) ListAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agents, err := h.service.ListAgents(r.Context())
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, agents, http.StatusOK)
}

// AgentByID handles GET /v1/agents/{agent_id} and POST /v1/agents/{agent_id}/halt|resume
func (h *Handlers) AgentByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/agents/")
	parts := strings.Split(path, "/")
	agentID := parts[0]

	if len(parts) == 1 {
		// GET /v1/agents/{agent_id}
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		agent, err := h.service.GetAgent(r.Context(), agentID)
		if err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if agent == nil {
			respondError(w, "Agent not found", http.StatusNotFound)
			return
		}
		respondJSON(w, agent, http.StatusOK)
		return
	}

	action := parts[1]
	switch action {
	case "halt":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.Reason == "" {
			req.Reason = "Halted via API"
		}
		if err := h.service.HaltAgent(r.Context(), agentID, req.Reason); err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, map[string]string{"status": "halted"}, http.StatusOK)

	case "resume":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := h.service.ResumeAgent(r.Context(), agentID); err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, map[string]string{"status": "resumed"}, http.StatusOK)

	case "limits":
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		lim, err := h.service.GetAgentLimits(r.Context(), agentID)
		if err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if lim == nil {
			respondError(w, "limits not available", http.StatusNotFound)
			return
		}
		respondJSON(w, lim, http.StatusOK)

	default:
		http.NotFound(w, r)
	}
}

// StartSession handles POST /v1/sessions/start
func (h *Handlers) StartSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AgentID  string                 `json:"agent_id"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session, err := h.service.StartSession(r.Context(), req.AgentID, req.Metadata)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, session, http.StatusOK)
}

// SessionByID handles POST /v1/sessions/{session_id}/end
func (h *Handlers) SessionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v1/sessions/")
	parts := strings.Split(path, "/")
	sessionID := parts[0]

	if len(parts) == 2 && parts[1] == "end" {
		if err := h.service.EndSession(r.Context(), sessionID); err != nil {
			respondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, map[string]string{"status": "ended"}, http.StatusOK)
		return
	}

	http.NotFound(w, r)
}

// ProposeAction handles POST /v1/actions/propose
func (h *Handlers) ProposeAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var proposal domain.ActionProposal
	if err := json.NewDecoder(r.Body).Decode(&proposal); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	decision, err := h.service.ProposeAction(r.Context(), proposal)
	if err != nil {
		log.Printf("ProposeAction error: %v", err)
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, decision, http.StatusOK)
}

// GetEvents handles GET /v1/events
func (h *Handlers) GetEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var agentID *string
	if id := r.URL.Query().Get("agent_id"); id != "" {
		agentID = &id
	}

	var sinceTS *int64
	if tsStr := r.URL.Query().Get("since_ts"); tsStr != "" {
		if ts, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
			sinceTS = &ts
		}
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	events, err := h.service.GetEvents(r.Context(), agentID, sinceTS, limit)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, events, http.StatusOK)
}

// GetEvent handles GET /v1/events/{event_id}
func (h *Handlers) GetEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	eventID := strings.TrimPrefix(r.URL.Path, "/v1/events/")
	// TODO: Implement GetEvent in service
	respondJSON(w, domain.Event{EventID: eventID}, http.StatusOK)
}

// PanicStatus handles GET /v1/panic
func (h *Handlers) PanicStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	state, err := h.service.GetPanicState(r.Context())
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, state, http.StatusOK)
}

// PanicOn handles POST /v1/panic/on
func (h *Handlers) PanicOn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		TTLSeconds int    `json:"ttl_seconds"`
		Reason     string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	now := time.Now()
	state := domain.PanicState{
		Enabled:    true,
		EnabledAt:  now,
		TTLSeconds: body.TTLSeconds,
		Reason:     body.Reason,
	}
	if body.TTLSeconds > 0 {
		exp := now.Add(time.Duration(body.TTLSeconds) * time.Second)
		state.ExpiresAt = &exp
	}
	if err := h.service.SetPanicState(r.Context(), state); err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if h.autobundleMgr != nil {
		if path, err := h.autobundleMgr.MaybeBundleOnPanicToggle(r.Context(), true); err != nil {
			log.Printf("autobundle on panic_on: %v", err)
		} else if path != "" {
			log.Printf("Panic-on bundle: %s", path)
		}
	}
	respondJSON(w, state, http.StatusOK)
}

// PanicOff handles POST /v1/panic/off
func (h *Handlers) PanicOff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	state, _ := h.service.GetPanicState(r.Context())
	if state == nil {
		state = &domain.PanicState{}
	}
	state.Enabled = false
	state.EnabledAt = time.Time{}
	state.ExpiresAt = nil
	state.TTLSeconds = 0
	state.Reason = ""
	if err := h.service.SetPanicState(r.Context(), *state); err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if h.autobundleMgr != nil {
		if path, err := h.autobundleMgr.MaybeBundleOnPanicToggle(r.Context(), false); err != nil {
			log.Printf("autobundle on panic_off: %v", err)
		} else if path != "" {
			log.Printf("Panic-off bundle: %s", path)
		}
	}
	respondJSON(w, map[string]interface{}{"enabled": false}, http.StatusOK)
}

// AutobundleStatus handles GET /v1/autobundle
func (h *Handlers) AutobundleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	status, err := h.service.GetAutobundleStatus(r.Context())
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, status, http.StatusOK)
}

// AutobundleTest handles POST /v1/autobundle/test
func (h *Handlers) AutobundleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.autobundleMgr == nil {
		respondError(w, "autobundle not configured", http.StatusBadRequest)
		return
	}
	path, err := h.autobundleMgr.MaybeBundleTest(r.Context())
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if path == "" {
		respondJSON(w, map[string]string{"message": "autobundle disabled"}, http.StatusOK)
		return
	}
	respondJSON(w, map[string]string{"path": path}, http.StatusOK)
}

// LimitsConfig handles GET /v1/limits/config — default limits from config (read-only).
func (h *Handlers) LimitsConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg, err := h.service.GetLimitsConfig(r.Context())
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, cfg, http.StatusOK)
}

// Helper functions

func respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// ServeUI serves the web UI
func (h *Handlers) ServeUI(w http.ResponseWriter, r *http.Request) {
	// Get the working directory (where the binary is run from)
	wd, err := os.Getwd()
	if err != nil {
		// Fallback: try to find web-ui relative to common locations
		wd = "."
	}

	// Try multiple possible paths
	possiblePaths := []string{
		filepath.Join(wd, "web-ui", "public"),
		filepath.Join(wd, "..", "web-ui", "public"),
		filepath.Join(".", "web-ui", "public"),
	}

	var uiDir string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			uiDir = path
			break
		}
	}

	if uiDir == "" {
		// If we can't find the directory, try relative to executable
		exe, _ := os.Executable()
		exeDir := filepath.Dir(exe)
		uiDir = filepath.Join(exeDir, "web-ui", "public")
	}

	// Handle path
	path := r.URL.Path
	if path == "/ui" || path == "/ui/" {
		path = "/index.html"
	} else {
		path = strings.TrimPrefix(path, "/ui")
	}

	filePath := filepath.Join(uiDir, path)
	
	// Security: ensure we're only serving from the uiDir
	if !strings.HasPrefix(filePath, uiDir) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// If file doesn't exist, try index.html
		filePath = filepath.Join(uiDir, "index.html")
	}

	http.ServeFile(w, r, filePath)
}
