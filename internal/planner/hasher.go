package planner

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/futurematic/kernel/internal/domain"
)

// Hasher computes deterministic hashes for plans
type Hasher struct{}

// NewHasher creates a new hasher
func NewHasher() *Hasher {
	return &Hasher{}
}

// HashPlan computes a deterministic hash for a plan
// Hash includes: expanded changes (canonical JSON) + policy set hash + namespace ID + intents (canonical JSON)
func (h *Hasher) HashPlan(namespaceID *string, intents []domain.Intent, expanded []domain.Change, policyHash string) (string, error) {
	// Canonicalize intents
	intentsJSON, err := h.canonicalizeJSON(intents)
	if err != nil {
		return "", fmt.Errorf("failed to canonicalize intents: %w", err)
	}

	// Canonicalize expanded changes
	expandedJSON, err := h.canonicalizeJSON(expanded)
	if err != nil {
		return "", fmt.Errorf("failed to canonicalize expanded: %w", err)
	}

	// Build hash input
	nsIDStr := ""
	if namespaceID != nil {
		nsIDStr = *namespaceID
	}

	hashInput := fmt.Sprintf("%s|%s|%s|%s", expandedJSON, policyHash, nsIDStr, intentsJSON)

	// Compute SHA256 hash
	hash := sha256.Sum256([]byte(hashInput))
	hashStr := hex.EncodeToString(hash[:])

	return "sha256:" + hashStr, nil
}

// canonicalizeJSON converts a value to canonical JSON
// This ensures deterministic output by sorting map keys and arrays
func (h *Hasher) canonicalizeJSON(v interface{}) (string, error) {
	// First, marshal to JSON
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	// Unmarshal to a generic interface to normalize
	var normalized interface{}
	if err := json.Unmarshal(jsonBytes, &normalized); err != nil {
		return "", err
	}

	// Recursively sort maps and arrays
	normalized = h.sortJSON(normalized)

	// Marshal again with sorted keys
	sortedBytes, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}

	return string(sortedBytes), nil
}

// sortJSON recursively sorts map keys and arrays for deterministic output
func (h *Hasher) sortJSON(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		// Sort map keys
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		sorted := make(map[string]interface{})
		for _, k := range keys {
			sorted[k] = h.sortJSON(val[k])
		}
		return sorted

	case []interface{}:
		// Sort array elements (if they're comparable)
		// For arrays, we keep order but sort nested structures
		sorted := make([]interface{}, len(val))
		for i, item := range val {
			sorted[i] = h.sortJSON(item)
		}
		return sorted

	default:
		return v
	}
}
