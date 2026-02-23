package bundle

import (
	"strings"
)

// RedactKeys is the list of keys to redact (case-insensitive match).
var RedactKeys = []string{"api_key", "token", "password", "secret", "key"}

// RedactMap recursively redacts map values for sensitive keys. Values are replaced with "[redacted]".
func RedactMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		if isSensitiveKey(k) {
			out[k] = "[redacted]"
			continue
		}
		out[k] = redactValue(v)
	}
	return out
}

func isSensitiveKey(k string) bool {
	lower := strings.ToLower(k)
	for _, r := range RedactKeys {
		if lower == r || strings.Contains(lower, r) {
			return true
		}
	}
	return false
}

func redactValue(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		return RedactMap(x)
	case []interface{}:
		arr := make([]interface{}, len(x))
		for i, e := range x {
			arr[i] = redactValue(e)
		}
		return arr
	default:
		return v
	}
}
