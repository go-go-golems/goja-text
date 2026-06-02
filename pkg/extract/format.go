package extract

import (
	"encoding/json"
	"strings"
)

func inferFormatFromPayload(payload string) string {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return "unknown"
	}
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) || (strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		var v any
		if json.Unmarshal([]byte(trimmed), &v) == nil {
			return "json"
		}
		return "json"
	}
	if strings.Contains(trimmed, "\n") && strings.Contains(trimmed, ":") {
		return "yaml"
	}
	return "unknown"
}

func inferFormatFromLabel(label string) string {
	label = strings.ToLower(strings.TrimSpace(label))
	if label == "" {
		return "unknown"
	}
	fields := strings.Fields(label)
	if len(fields) > 0 {
		label = fields[0]
	}
	switch label {
	case "json", "jsonc":
		return "json"
	case "yaml", "yml":
		return "yaml"
	case "xml":
		return "xml"
	case "toml":
		return "toml"
	case "data", "result", "answer", "payload", "arguments", "tool_call":
		return "unknown"
	default:
		return label
	}
}
