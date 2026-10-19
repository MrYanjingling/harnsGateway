package audit

import (
	"encoding/json"
	"strings"
)

var sensitiveKeys = map[string]struct{}{
	"password":  {},
	"authkey":   {},
	"token":     {},
	"accesskey": {},
	"secretkey": {},
	"ak":        {},
	"sk":        {},
	"secret":    {},
}

// MaskJSON masks sensitive fields recursively. ok is false when raw is not JSON.
func MaskJSON(raw []byte) (string, bool) {
	if len(raw) == 0 {
		return "", true
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", false
	}
	maskValue(v)
	b, err := json.Marshal(v)
	if err != nil {
		return "", false
	}
	return string(b), true
}

func maskValue(v interface{}) {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			if _, ok := sensitiveKeys[strings.ToLower(k)]; ok {
				t[k] = "***"
				continue
			}
			maskValue(val)
		}
	case []interface{}:
		for _, item := range t {
			maskValue(item)
		}
	}
}

func truncate(s string, limit int) string {
	if limit <= 0 || len(s) <= limit {
		return s
	}
	return s[:limit] + "...(truncated)"
}
