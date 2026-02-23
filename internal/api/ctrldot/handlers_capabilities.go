package ctrldot

import (
	"net/http"
)

// Capabilities handles GET /v1/capabilities (agent discovery).
func (h *Handlers) Capabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	caps, err := h.service.GetCapabilities(r.Context())
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, caps, http.StatusOK)
}
