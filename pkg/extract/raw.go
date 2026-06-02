package extract

import (
	"encoding/json"
	"regexp"
	"strings"

	jsonsanitize "github.com/go-go-golems/sanitize/pkg/json"
	yamlsanitize "github.com/go-go-golems/sanitize/pkg/yaml"
)

var yamlMappingLine = regexp.MustCompile(`(?m)^\s*[^#\s][^:\n]+:\s+.+$`)
var yamlListLine = regexp.MustCompile(`(?m)^\s*-\s+.+$`)

// RawStructured recognizes whole-input raw JSON or YAML candidates conservatively.
func RawStructured(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) {
	options = optionsOrDefault(options)
	trimmed, start, end := trimSpan(input)
	if trimmed == "" {
		return nil, nil
	}
	li := newLineIndex(input)
	var candidates []*ExtractionCandidate

	if looksJSON(trimmed) {
		var confidence float64
		var v any
		if json.Unmarshal([]byte(trimmed), &v) == nil {
			confidence = 0.98
		} else {
			result := jsonsanitize.Sanitize(trimmed)
			if !result.StrictParseClean {
				return nil, nil
			}
			confidence = 0.75
		}
		c := &ExtractionCandidate{Kind: "raw", Format: "json", Text: trimmed, Raw: trimmed, Wrapper: "none", Label: "json", Confidence: confidence}
		li.fillSpan(c, start, end, start, end)
		candidates = append(candidates, c)
		return filterCandidates(candidates, options), nil
	}

	if looksYAML(trimmed) {
		result := yamlsanitize.Sanitize(trimmed)
		if result.ParseClean || result.LintClean || len(result.Fixes) > 0 {
			confidence := 0.7
			if result.ParseClean && result.LintClean {
				confidence = 0.9
			}
			c := &ExtractionCandidate{Kind: "raw", Format: "yaml", Text: trimmed, Raw: trimmed, Wrapper: "none", Label: "yaml", Confidence: confidence}
			li.fillSpan(c, start, end, start, end)
			candidates = append(candidates, c)
		}
	}
	return filterCandidates(candidates, options), nil
}

func trimSpan(input string) (string, int, int) {
	start := 0
	end := len(input)
	for start < end {
		switch input[start] {
		case ' ', '\t', '\n', '\r':
			start++
		default:
			goto trimEnd
		}
	}
trimEnd:
	for end > start {
		switch input[end-1] {
		case ' ', '\t', '\n', '\r':
			end--
		default:
			return input[start:end], start, end
		}
	}
	return input[start:end], start, end
}

func looksJSON(input string) bool {
	return (strings.HasPrefix(input, "{") && strings.Contains(input, "}")) || (strings.HasPrefix(input, "[") && strings.Contains(input, "]"))
}

func looksYAML(input string) bool {
	if strings.HasPrefix(input, "{") || strings.HasPrefix(input, "[") {
		return false
	}
	mapping := yamlMappingLine.FindAllString(input, -1)
	list := yamlListLine.FindAllString(input, -1)
	if len(mapping) >= 2 || (len(mapping) >= 1 && len(list) >= 1) {
		return true
	}
	return false
}
