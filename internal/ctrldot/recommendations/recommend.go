package recommendations

import (
	"context"
	"fmt"
	"strings"

	"github.com/futurematic/kernel/internal/domain"
)

// Stable reason codes for agent logic
const (
	CodeResolutionRequired   = "PANIC_RESOLUTION_REQUIRED"
	CodeNetworkDomainDenied  = "NETWORK_DOMAIN_DENIED"
	CodeBudgetStopThreshold  = "BUDGET_STOP_THRESHOLD"
	CodeLoopStopThreshold    = "LOOP_STOP_THRESHOLD"
	CodeAgentHalted          = "AGENT_HALTED"
	CodeFilesystemDenied     = "FILESYSTEM_DENIED"
	CodeResolutionMissing    = "RESOLUTION_REQUIRED"
)

// RecommendOptions supplies inputs for building a recommendation
type RecommendOptions struct {
	Decision       domain.Decision
	ReasonText     string
	ReasonCodes    []string
	ActionType     string
	PanicEnabled   bool
	ResolutionAbsent bool
	AgentID        string
	SessionID      string
}

// Recommend returns a deterministic Recommendation for DENY/STOP/THROTTLE (and optionally WARN).
func Recommend(ctx context.Context, opts RecommendOptions) *domain.Recommendation {
	if opts.AgentID == "" {
		opts.AgentID = "<agent_id>"
	}
	codeSet := make(map[string]bool)
	for _, c := range opts.ReasonCodes {
		codeSet[c] = true
	}

	switch opts.Decision {
	case domain.DecisionDeny, domain.DecisionStop:
		// Resolution required (rules or panic)
		if codeSet[CodeResolutionRequired] || codeSet[CodeResolutionMissing] || strings.Contains(opts.ReasonText, "resolution") || strings.Contains(opts.ReasonText, "Requires resolution") {
			return &domain.Recommendation{
				Kind:    "use_resolution",
				Title:   "Resolution required",
				Summary: opts.ReasonText,
				NextSteps: []string{
					fmt.Sprintf("ctrldot resolve allow-once --agent %s --ttl 120s", opts.AgentID),
					"# Or disable panic: ctrldot panic off",
				},
				DocsHint: "docs/SETUP_GUIDE.md#panic-mode",
				Tags:     []string{"resolution", "panic"},
			}
		}
		// Network denied
		if codeSet[CodeNetworkDomainDenied] || strings.Contains(strings.ToLower(opts.ReasonText), "network") {
			return &domain.Recommendation{
				Kind:    "tighten_scope",
				Title:   "Network access denied",
				Summary: opts.ReasonText,
				NextSteps: []string{
					"# Add domain to config rules.network.allow_domains, or: ctrldot panic off",
				},
				DocsHint: "docs/SETUP_GUIDE.md",
				Tags:     []string{"network", "rules"},
			}
		}
		// Filesystem denied
		if codeSet[CodeFilesystemDenied] || strings.Contains(strings.ToLower(opts.ReasonText), "filesystem") {
			return &domain.Recommendation{
				Kind:    "tighten_scope",
				Title:   "Filesystem access denied",
				Summary: opts.ReasonText,
				NextSteps: []string{
					"# Add path to config rules.filesystem.allow_roots, or: ctrldot panic off",
				},
				Tags: []string{"filesystem", "rules"},
			}
		}
		// Loop stop
		if codeSet[CodeLoopStopThreshold] || strings.Contains(opts.ReasonText, "Loop") {
			return &domain.Recommendation{
				Kind:    "reduce_loop",
				Title:   "Loop detected",
				Summary: opts.ReasonText,
				NextSteps: []string{
					"# Action repeated too many times; vary the action or: ctrldot panic off",
				},
				Tags: []string{"loop"},
			}
		}
		// Budget stop
		if codeSet[CodeBudgetStopThreshold] || strings.Contains(strings.ToLower(opts.ReasonText), "budget") {
			return &domain.Recommendation{
				Kind:    "enable_panic",
				Title:   "Budget limit reached",
				Summary: opts.ReasonText,
				NextSteps: []string{
					"# Daily budget exceeded; wait for reset or: ctrldot panic off (reduces cap)",
				},
				Tags: []string{"budget", "limits"},
			}
		}
		// Agent halted
		if codeSet[CodeAgentHalted] || strings.Contains(opts.ReasonText, "halted") {
			return &domain.Recommendation{
				Kind:    "enable_ctrldot",
				Title:   "Agent is halted",
				Summary: opts.ReasonText,
				NextSteps: []string{
					fmt.Sprintf("ctrldot agents %s resume  # or via API POST /v1/agents/%s/resume", opts.AgentID, opts.AgentID),
				},
				Tags: []string{"halt"},
			}
		}
		// Generic deny/stop
		return &domain.Recommendation{
			Kind:    "tighten_scope",
			Title:   "Action denied or stopped",
			Summary: opts.ReasonText,
			NextSteps: []string{
				"ctrldot panic off  # if appropriate",
				"# Or provide resolution token for this action type",
			},
			Tags: []string{"deny", "stop"},
		}
	case domain.DecisionThrottle:
		return &domain.Recommendation{
			Kind:    "reduce_loop",
			Title:   "Throttled",
			Summary: opts.ReasonText,
			NextSteps: []string{
				"# Approaching limits; reduce rate or: ctrldot panic off",
			},
			Tags: []string{"throttle"},
		}
	default:
		return nil
	}
}
