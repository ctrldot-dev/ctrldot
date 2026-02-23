package rules

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/futurematic/kernel/internal/config"
	"github.com/futurematic/kernel/internal/domain"
)

// Engine evaluates domain rules
type Engine struct {
	config *config.Config
}

// NewEngine creates a new rules engine
func NewEngine(cfg *config.Config) *Engine {
	return &Engine{
		config: cfg,
	}
}

// Evaluate evaluates rules and returns decision and reason (uses engine config).
func (e *Engine) Evaluate(ctx context.Context, proposal domain.ActionProposal) (domain.Decision, string) {
	return e.EvaluateWithConfig(ctx, proposal, e.config)
}

// EvaluateWithConfig evaluates rules using the given config (e.g. effective config when panic is on).
func (e *Engine) EvaluateWithConfig(ctx context.Context, proposal domain.ActionProposal, cfg *config.Config) (domain.Decision, string) {
	if cfg == nil {
		cfg = e.config
	}
	actionType := proposal.Action.Type

	// Check require_resolution rules
	for _, requiredAction := range cfg.Rules.RequireResolution {
		if actionType == requiredAction || strings.HasPrefix(actionType, requiredAction+".") {
			// Check if resolution token provided
			if proposal.ResolutionToken == "" {
				return domain.DecisionDeny, fmt.Sprintf("Requires resolution for %s", actionType)
			}
			// Token validation happens in resolution manager
		}
	}

	// Check filesystem rules
	if strings.HasPrefix(actionType, "filesystem.") {
		if !e.checkFilesystemRulesWithConfig(proposal, cfg) {
			return domain.DecisionDeny, "Filesystem access denied by rules"
		}
	}

	// Check network rules
	if strings.HasPrefix(actionType, "network.") || strings.HasPrefix(actionType, "http.") || strings.HasPrefix(actionType, "web.") {
		if !e.checkNetworkRulesWithConfig(proposal, cfg) {
			return domain.DecisionDeny, "Network access denied by rules"
		}
	}

	return domain.DecisionAllow, ""
}

func (e *Engine) checkFilesystemRules(proposal domain.ActionProposal) bool {
	return e.checkFilesystemRulesWithConfig(proposal, e.config)
}

func (e *Engine) checkFilesystemRulesWithConfig(proposal domain.ActionProposal, cfg *config.Config) bool {
	if cfg == nil || len(cfg.Rules.Filesystem.AllowRoots) == 0 {
		return true // No restrictions
	}

	target, ok := proposal.Action.Target["path"].(string)
	if !ok {
		return false
	}

	for _, allowedRoot := range cfg.Rules.Filesystem.AllowRoots {
		if strings.HasPrefix(target, allowedRoot) || strings.HasPrefix(target, expandHome(allowedRoot)) {
			return true
		}
	}

	return false
}

func (e *Engine) checkNetworkRules(proposal domain.ActionProposal) bool {
	return e.checkNetworkRulesWithConfig(proposal, e.config)
}

func (e *Engine) checkNetworkRulesWithConfig(proposal domain.ActionProposal, cfg *config.Config) bool {
	if cfg == nil || !cfg.Rules.Network.DenyAll {
		return true
	}

	domainVal, ok := proposal.Action.Target["domain"].(string)
	if !ok {
		url, ok := proposal.Action.Target["url"].(string)
		if ok {
			domainVal = extractDomain(url)
		} else {
			return false
		}
	}

	for _, allowedDomain := range cfg.Rules.Network.AllowDomains {
		if domainVal == allowedDomain || strings.HasSuffix(domainVal, "."+allowedDomain) {
			return true
		}
	}

	return false
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home := os.Getenv("HOME")
		if home != "" {
			return strings.Replace(path, "~", home, 1)
		}
	}
	return path
}

func extractDomain(url string) string {
	// Simple domain extraction
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}
	if idx := strings.Index(url, "/"); idx >= 0 {
		url = url[:idx]
	}
	return url
}
