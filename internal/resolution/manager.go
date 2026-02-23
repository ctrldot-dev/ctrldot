package resolution

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/futurematic/kernel/internal/store"
)

// Manager manages resolution tokens
type Manager struct {
	store     store.Store
	secretKey string
}

// NewManager creates a new resolution manager
func NewManager(store store.Store, secretKey string) *Manager {
	if secretKey == "" {
		secretKey = "default-secret-key-change-in-production" // TODO: use proper secret
	}
	return &Manager{
		store:     store,
		secretKey: secretKey,
	}
}

// GenerateToken generates a resolution token
func (m *Manager) GenerateToken(ctx context.Context, agentID string, actionType string, ttl time.Duration) (string, error) {
	expiresAt := time.Now().Add(ttl)
	
	data := fmt.Sprintf("%s:%s:%d", agentID, actionType, expiresAt.Unix())
	mac := hmac.New(sha256.New, []byte(m.secretKey))
	mac.Write([]byte(data))
	signature := hex.EncodeToString(mac.Sum(nil))
	
	token := fmt.Sprintf("%s:%s", data, signature)
	return token, nil
}

// ValidateToken validates a resolution token
func (m *Manager) ValidateToken(ctx context.Context, token string, agentID string, actionType string) (bool, error) {
	parts := strings.Split(token, ":")
	if len(parts) != 4 {
		return false, nil
	}

	tokenAgentID := parts[0]
	tokenActionType := parts[1]
	expiresAtStr := parts[2]
	signature := parts[3]

	// Verify agent and action match
	if tokenAgentID != agentID || tokenActionType != actionType {
		return false, nil
	}

	// Verify expiration
	var expiresAt int64
	fmt.Sscanf(expiresAtStr, "%d", &expiresAt)
	if time.Now().Unix() > expiresAt {
		return false, nil
	}

	// Verify signature
	data := fmt.Sprintf("%s:%s:%s", tokenAgentID, tokenActionType, expiresAtStr)
	mac := hmac.New(sha256.New, []byte(m.secretKey))
	mac.Write([]byte(data))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature)), nil
}
