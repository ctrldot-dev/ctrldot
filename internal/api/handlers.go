package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/kernel"
	"github.com/futurematic/kernel/internal/query"
)

// Handlers contains HTTP handlers
type Handlers struct {
	kernelService kernel.Service
	queryEngine   query.Engine
}

// NewHandlers creates new HTTP handlers
func NewHandlers(kernelService kernel.Service, queryEngine query.Engine) *Handlers {
	return &Handlers{
		kernelService: kernelService,
		queryEngine:   queryEngine,
	}
}

// Plan handles POST /v1/plan
func (h *Handlers) Plan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ActorID     string          `json:"actor_id"`
		Capabilities []string        `json:"capabilities"`
		NamespaceID *string          `json:"namespace_id"`
		AsOf        domain.AsOf     `json:"asof"`
		Intents     []domain.Intent  `json:"intents"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, NewError(ErrorCodeValidation, "Invalid request body", nil), http.StatusBadRequest)
		return
	}

	kernelReq := kernel.PlanRequest{
		ActorID:     req.ActorID,
		Capabilities: req.Capabilities,
		NamespaceID: req.NamespaceID,
		AsOf:        req.AsOf,
		Intents:     req.Intents,
	}

	plan, err := h.kernelService.Plan(r.Context(), kernelReq)
	if err != nil {
		respondError(w, NewError(ErrorCodeInternal, err.Error(), nil), http.StatusInternalServerError)
		return
	}

	respondJSON(w, plan, http.StatusOK)
}

// Apply handles POST /v1/apply
func (h *Handlers) Apply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ActorID     string   `json:"actor_id"`
		Capabilities []string `json:"capabilities"`
		PlanID      string   `json:"plan_id"`
		PlanHash    string   `json:"plan_hash"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, NewError(ErrorCodeValidation, "Invalid request body", nil), http.StatusBadRequest)
		return
	}

	kernelReq := kernel.ApplyRequest{
		ActorID:     req.ActorID,
		Capabilities: req.Capabilities,
		PlanID:      req.PlanID,
		PlanHash:    req.PlanHash,
	}

	operation, err := h.kernelService.Apply(r.Context(), kernelReq)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "plan hash mismatch") || strings.Contains(err.Error(), "already applied") {
			respondError(w, NewError(ErrorCodeConflict, err.Error(), nil), http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "policy denied") {
			respondError(w, NewError(ErrorCodePolicyDenied, err.Error(), nil), http.StatusForbidden)
			return
		}
		respondError(w, NewError(ErrorCodeInternal, err.Error(), nil), http.StatusInternalServerError)
		return
	}

	respondJSON(w, operation, http.StatusOK)
}

// Expand handles GET /v1/expand
func (h *Handlers) Expand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idsStr := r.URL.Query().Get("ids")
	if idsStr == "" {
		respondError(w, NewError(ErrorCodeValidation, "ids parameter is required", nil), http.StatusBadRequest)
		return
	}

	nodeIDs := strings.Split(idsStr, ",")
	for i := range nodeIDs {
		nodeIDs[i] = strings.TrimSpace(nodeIDs[i])
	}

	namespaceID := r.URL.Query().Get("namespace_id")
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}

	depth := 1
	if depthStr := r.URL.Query().Get("depth"); depthStr != "" {
		var err error
		depth, err = strconv.Atoi(depthStr)
		if err != nil {
			respondError(w, NewError(ErrorCodeValidation, "Invalid depth parameter", nil), http.StatusBadRequest)
			return
		}
	}

	// Resolve asof
	asofSeqStr := r.URL.Query().Get("asof_seq")
	asofTimeStr := r.URL.Query().Get("asof_time")

	var asofSeq int64
	if asofSeqStr != "" {
		seq, err := strconv.ParseInt(asofSeqStr, 10, 64)
		if err != nil {
			respondError(w, NewError(ErrorCodeValidation, "Invalid asof_seq parameter", nil), http.StatusBadRequest)
			return
		}
		asofSeq = seq
	} else if asofTimeStr != "" {
		// For now, we'll require asof_seq - time resolution would need store access
		respondError(w, NewError(ErrorCodeValidation, "asof_time not yet supported, use asof_seq", nil), http.StatusBadRequest)
		return
	} else {
		// Default to latest
		asofSeq = 0 // Will be resolved by query engine
	}

	expandReq := query.ExpandRequest{
		NodeIDs:     nodeIDs,
		NamespaceID: nsID,
		Depth:       depth,
		AsOfSeq:     asofSeq,
	}

	result, err := h.queryEngine.Expand(r.Context(), expandReq)
	if err != nil {
		respondError(w, NewError(ErrorCodeInternal, err.Error(), nil), http.StatusInternalServerError)
		return
	}

	respondJSON(w, result, http.StatusOK)
}

// History handles GET /v1/history
func (h *Handlers) History(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	target := r.URL.Query().Get("target")
	if target == "" {
		respondError(w, NewError(ErrorCodeValidation, "target parameter is required", nil), http.StatusBadRequest)
		return
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			respondError(w, NewError(ErrorCodeValidation, "Invalid limit parameter", nil), http.StatusBadRequest)
			return
		}
	}

	historyReq := query.HistoryRequest{
		Target: target,
		Limit:  limit,
	}

	operations, err := h.queryEngine.History(r.Context(), historyReq)
	if err != nil {
		respondError(w, NewError(ErrorCodeInternal, err.Error(), nil), http.StatusInternalServerError)
		return
	}

	respondJSON(w, operations, http.StatusOK)
}

// Diff handles GET /v1/diff
func (h *Handlers) Diff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	aSeqStr := r.URL.Query().Get("a_seq")
	bSeqStr := r.URL.Query().Get("b_seq")
	target := r.URL.Query().Get("target")

	if aSeqStr == "" || bSeqStr == "" {
		respondError(w, NewError(ErrorCodeValidation, "a_seq and b_seq parameters are required", nil), http.StatusBadRequest)
		return
	}

	aSeq, err := strconv.ParseInt(aSeqStr, 10, 64)
	if err != nil {
		respondError(w, NewError(ErrorCodeValidation, "Invalid a_seq parameter", nil), http.StatusBadRequest)
		return
	}

	bSeq, err := strconv.ParseInt(bSeqStr, 10, 64)
	if err != nil {
		respondError(w, NewError(ErrorCodeValidation, "Invalid b_seq parameter", nil), http.StatusBadRequest)
		return
	}

	diffReq := query.DiffRequest{
		ASeq:   aSeq,
		BSeq:   bSeq,
		Target: target,
	}

	result, err := h.queryEngine.Diff(r.Context(), diffReq)
	if err != nil {
		respondError(w, NewError(ErrorCodeInternal, err.Error(), nil), http.StatusInternalServerError)
		return
	}

	respondJSON(w, result, http.StatusOK)
}

// Healthz handles GET /v1/healthz
func (h *Handlers) Healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	respondJSON(w, map[string]bool{"ok": true}, http.StatusOK)
}

// Helper functions

func respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, err *ErrorResponse, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err.JSON()
	json.NewEncoder(w).Encode(err)
}
