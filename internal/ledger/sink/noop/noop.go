package noop

import (
	"context"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/ledger/sink"
)

// Sink is a no-op ledger sink (ledger_sink: none).
type Sink struct{}

// New returns a no-op sink.
func New() *Sink {
	return &Sink{}
}

// EmitDecision does nothing.
func (s *Sink) EmitDecision(ctx context.Context, d *sink.DecisionRecord) error {
	return nil
}

// EmitEvent does nothing.
func (s *Sink) EmitEvent(ctx context.Context, e *domain.Event) error {
	return nil
}

// Close does nothing.
func (s *Sink) Close() error {
	return nil
}

// Ensure Sink implements sink.LedgerSink.
var _ sink.LedgerSink = (*Sink)(nil)
