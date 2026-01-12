package policy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/store"
	"gopkg.in/yaml.v3"
)

// Engine evaluates policies against proposed changes
type Engine interface {
	// Evaluate evaluates a policy set against proposed changes
	Evaluate(ctx context.Context, policySet *domain.PolicySet, changes []domain.Change, namespaceID *string, asofSeq int64, store store.Store) (*domain.PolicyReport, error)

	// ComputePolicyHash computes a hash for a policy set YAML
	ComputePolicyHash(policyYAML string) (string, error)
}

// NewEngine creates a new policy engine
func NewEngine() Engine {
	return &engine{
		predicates: NewPredicateEvaluator(),
	}
}

type engine struct {
	predicates *PredicateEvaluator
}

// Evaluate implements the Engine interface
func (e *engine) Evaluate(ctx context.Context, policySet *domain.PolicySet, changes []domain.Change, namespaceID *string, asofSeq int64, store store.Store) (*domain.PolicyReport, error) {
	if policySet == nil {
		// No policy set means no restrictions
		return &domain.PolicyReport{}, nil
	}

	// Parse policy YAML
	policy, err := e.parsePolicySet(policySet.PolicyYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse policy set: %w", err)
	}

	// Open a read transaction for policy evaluation
	tx, err := store.OpenTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open transaction: %w", err)
	}
	defer tx.Rollback() // Always rollback read transactions

	report := &domain.PolicyReport{
		Denies: []domain.PolicyViolation{},
		Warns:  []domain.PolicyViolation{},
		Infos:  []domain.PolicyViolation{},
	}

	// Evaluate each change against matching rules
	for _, change := range changes {
		for _, rule := range policy.Rules {
			if e.ruleMatches(rule, change) {
				violations, err := e.evaluateRule(ctx, rule, change, namespaceID, asofSeq, tx)
				if err != nil {
					return nil, fmt.Errorf("failed to evaluate rule %s: %w", rule.ID, err)
				}

				// Add violations to report based on effect
				for _, violation := range violations {
					switch rule.Effect.Type {
					case "deny":
						report.Denies = append(report.Denies, violation)
					case "warn":
						report.Warns = append(report.Warns, violation)
					case "info":
						report.Infos = append(report.Infos, violation)
					}
				}
			}
		}
	}

	return report, nil
}

// ComputePolicyHash computes a SHA256 hash of the policy YAML
func (e *engine) ComputePolicyHash(policyYAML string) (string, error) {
	hash := sha256.Sum256([]byte(policyYAML))
	hashStr := hex.EncodeToString(hash[:])
	return "sha256:" + hashStr, nil
}

// ruleMatches checks if a rule's "when" conditions match a change
func (e *engine) ruleMatches(rule Rule, change domain.Change) bool {
	// Check operation type
	if len(rule.When.Op) > 0 {
		matched := false
		for _, op := range rule.When.Op {
			if op == change.Kind {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check link type (if applicable)
	if rule.When.LinkType != "" {
		linkType, ok := change.Payload["type"].(string)
		if !ok || linkType != rule.When.LinkType {
			return false
		}
	}

	return true
}

// evaluateRule evaluates a rule's predicates against a change
func (e *engine) evaluateRule(ctx context.Context, rule Rule, change domain.Change, namespaceID *string, asofSeq int64, tx store.Tx) ([]domain.PolicyViolation, error) {
	var violations []domain.PolicyViolation

	// Evaluate each predicate requirement
	for _, req := range rule.Require {
		passed, err := e.predicates.Evaluate(ctx, req.Predicate, req.Args, change, namespaceID, asofSeq, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate predicate %s: %w", req.Predicate, err)
		}

		if !passed {
			violation := domain.PolicyViolation{
				RuleID:  rule.ID,
				Message: rule.Effect.Message,
			}
			violations = append(violations, violation)
		}
	}

	return violations, nil
}

// parsePolicySet parses a YAML policy set
func (e *engine) parsePolicySet(yamlStr string) (*PolicySet, error) {
	var policy PolicySet
	if err := yaml.Unmarshal([]byte(yamlStr), &policy); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return &policy, nil
}
