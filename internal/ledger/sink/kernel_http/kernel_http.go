package kernel_http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/ledger/sink"
)

// Sink POSTs decision records to the Kernel HTTP API.
type Sink struct {
	baseURL   string
	apiKey    string
	timeout   time.Duration
	required  bool
	client    *http.Client
}

// NewSink creates a Kernel HTTP sink. baseURL is the Kernel root (e.g. http://127.0.0.1:8080).
func NewSink(baseURL, apiKey string, timeoutMs int, required bool) *Sink {
	if timeoutMs <= 0 {
		timeoutMs = 2000
	}
	return &Sink{
		baseURL:  baseURL,
		apiKey:   apiKey,
		timeout:  time.Duration(timeoutMs) * time.Millisecond,
		required: required,
		client: &http.Client{
			Timeout: time.Duration(timeoutMs) * time.Millisecond,
		},
	}
}

// EmitDecision POSTs the decision record to Kernel. Best-effort unless Required; on failure logs and returns error only if Required.
func (s *Sink) EmitDecision(ctx context.Context, d *sink.DecisionRecord) error {
	body, err := json.Marshal(d)
	if err != nil {
		if s.required {
			return fmt.Errorf("marshal decision: %w", err)
		}
		log.Printf("kernel_http: marshal decision: %v", err)
		return nil
	}
	url := s.baseURL + "/v1/ctrldot/decisions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		if s.required {
			return err
		}
		log.Printf("kernel_http: new request: %v", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		// One retry with fresh body
		req2, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		req2.Header.Set("Content-Type", "application/json")
		if s.apiKey != "" {
			req2.Header.Set("Authorization", "Bearer "+s.apiKey)
		}
		resp, err = s.client.Do(req2)
	}
	if err != nil {
		if s.required {
			return fmt.Errorf("POST %s: %w", url, err)
		}
		log.Printf("kernel_http: POST %s: %v", url, err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("POST %s: %d", url, resp.StatusCode)
		if s.required {
			return err
		}
		log.Printf("kernel_http: %v", err)
		return nil
	}
	return nil
}

// EmitEvent is a no-op for Kernel HTTP (decisions only).
func (s *Sink) EmitEvent(ctx context.Context, e *domain.Event) error {
	return nil
}

// Close is a no-op.
func (s *Sink) Close() error {
	return nil
}

var _ sink.LedgerSink = (*Sink)(nil)
